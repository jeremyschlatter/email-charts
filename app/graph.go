package app

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/dustin/go-humanize"
)

var epoch = time.Unix(0, 0)
var TempGraphDir = path.Join(os.TempDir(), "email_chart_graphs")
var TempAnalysisDir = path.Join(os.TempDir(), "analysis_records")

func init() {
	for _, dir := range []string{TempGraphDir, TempAnalysisDir} {
		err := os.Mkdir(dir, os.ModeDir|0764)
		if err != nil && !os.IsNotExist(err) {
			CallExitFuncs()
			log.Fatalf("Failed to create temp dir. -- %s\n", err.Error())
		}
		local_dir := dir // Give the deferred function its own copy of this string.
		RunAtExit(func() {
			log.Printf("Deleting %s.\n", local_dir)
			os.RemoveAll(local_dir)
		})
	}
}

type AnalysisData struct {
	GraphFile, DayGraph, BarGraph string
	Analysis                      []string
}

func formatHour(hour int) string {
	t, _ := time.Parse("15", fmt.Sprintf("%d", hour))
	return t.Format("3 PM")
}

func makeDateFile(user, authToken string) (string, [24]int, int) {
	headers, err := fetchAllHeaders(user, authToken, "[Gmail]/Sent Mail")
	// TODO
	//log.Panicln("Handle this error")

	parseFailures := 0
	var buf bytes.Buffer
	var emailCount int
	var hourBuckets [24]int
	for _, h := range headers {
		emailCount++
		if emailCount%5000 == 0 {
			log.Printf("Processed %d messages.\n", emailCount)
		}
		date, err := h.Date()
		if err != nil {
			parseFailures++
			continue
		}
		buf.WriteString(fmt.Sprintf("%d, %d\n", int(date.Sub(epoch).Hours()/24),
			(date.Hour()*60+date.Minute())*60+date.Second()))
		hourBuckets[date.Hour()]++
	}
	if parseFailures > 0 {
		log.Printf("Failed to parse %d headers.\n", parseFailures)
	}
	f, err := ioutil.TempFile(os.TempDir(), "tmp-csv-")
	checkError(err)
	defer func() {
		f.Close()
	}()
	_, err = f.WriteString("day, seconds\n")
	checkError(err)
	_, err = buf.WriteTo(f)
	checkError(err)
	return f.Name(), hourBuckets, emailCount
}

func makeGraph(rFile, csvFile string) string {
	out, err := ioutil.TempFile(TempGraphDir, "")
	checkError(err)
	defer out.Close()
	stderr, err := exec.Command(
		"R", "--no-save", "-f", "r-files/"+rFile, "--args", csvFile, out.Name()).CombinedOutput()
	if err != nil {
		log.Panicf("Letting this command die. Error running command: %s\n%s", err.Error(), string(stderr))
	}
	return path.Join("/graph", path.Base(out.Name()))
}

func RunAnalysis(user, authToken string) string {
	csvFile, hourBuckets, emailCount := makeDateFile(user, authToken)
	defer os.Remove(csvFile)
	record := AnalysisData{}
	record.GraphFile = makeGraph("graph.r", csvFile)
	record.DayGraph = makeGraph("daygraph.r", csvFile)
	record.BarGraph = makeGraph("bargraph.r", csvFile)
	f, err := ioutil.TempFile(TempAnalysisDir, "")
	checkError(err)
	defer f.Close()
	var maxHour, daytimeCount int
	for i := 9; i < 17; i++ {
		daytimeCount += hourBuckets[i]
	}
	for i := range hourBuckets {
		if hourBuckets[i] > hourBuckets[maxHour] {
			maxHour = i
		}
	}
	record.Analysis = []string{fmt.Sprintf("You send %d%% of your email between the hours of 9 AM and 5 PM.",
		100*daytimeCount/emailCount),
		fmt.Sprintf("You've sent %s emails in total.", humanize.Comma(int64(emailCount))),
		fmt.Sprintf("Your most active hour for sending emails is between %s and %s.",
			formatHour(maxHour), formatHour((maxHour+1)%24))}
	checkError(gob.NewEncoder(f).Encode(record))
	return path.Base(f.Name())
}
