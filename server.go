package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/jeremyschlatter/email-charts/app"
)

var err error
var stderr = flag.Bool("stderr", false, "log to stderr rather than log file")
var httpAddr = flag.String("http", ":8080", "port to listen to for http connections")

func visualizeHandler(w http.ResponseWriter, r *http.Request) {
	out := app.RunAnalysis(r.FormValue("user"), r.FormValue("token"))
	w.Write([]byte(out))
}

func loadingHandler(w http.ResponseWriter, r *http.Request) {
	loadingTmpl, _ := template.ParseFiles("templates/loading.html")
	err := loadingTmpl.Execute(w, map[string]string{
		"user": r.FormValue("user"), "token": r.FormValue("token")})
	if err != nil {
		log.Printf("Loading template error - %s\n", err.Error())
		w.Write([]byte("Sorry, an error occurred processing your graph. We're looking into it."))
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	dashboardTmpl, _ := template.ParseFiles("templates/dashboard.html")
	record, err := os.Open(path.Join(app.TempAnalysisDir, r.FormValue("data")))
	if err != nil {
		log.Printf("Got request for nonexistant record named '%s'.", r.FormValue("data"))
		w.Write([]byte("Sorry, an error occurred processing your graph. We're looking into it."))
		return
	}
	var analysis app.AnalysisData
	err = gob.NewDecoder(record).Decode(&analysis)
	if err != nil {
		log.Printf("Gob decode error - %s", err.Error())
		w.Write([]byte("Sorry, an error occurred processing your graph. We're looking into it."))
		return
	}
	err = dashboardTmpl.Execute(w, analysis)
	if err != nil {
		log.Printf("Dashboard template error - %s\n", err.Error())
		w.Write([]byte("Sorry, an error occurred processing your graph. We're looking into it."))
	}
}

var smtpPassword string

func signupHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Email address registered - %s\n", r.FormValue("email"))
	f, err := os.OpenFile(path.Join(os.Getenv("HOME"), "email_signups.txt"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err == nil {
		_, err = f.WriteString(r.FormValue("email") + "\n")
		if err != nil {
			log.Printf("Failure writing email signup to disk - %s\n", err)
		}
		f.Close()
	} else {
		log.Printf("Failure writing email signup to disk - %s\n", err)
	}
	http.ServeFile(w, r, "static-files/thanks.html")
	smtp.SendMail(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", "jeremy.schlatter@gmail.com", smtpPassword, "smtp.gmail.com"),
		"jeremy.schlatter@gmail.com", []string{"jeremy.schlatter@gmail.com", "mjcurzi@gmail.com"}, []byte(
			"Subject: Contact info submitted for emailcharts.com\r\n"+
				"To: jeremy.schlatter@gmail.com\r\n"+r.FormValue("email")))
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func TrapKillSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		for range c {
			log.Println("Caught a fatal signal. Terminating.")
			app.CallExitFuncs()
			os.Exit(1)
		}
	}()
}

func main() {
	flag.Parse()

	// Make sure exit functions get called. This program has no normal exit -- it either dies from panics
	// or from catching fatal signals. The former case is handled by defer. The latter case is handled
	// by TrapKillSignals.
	defer app.CallExitFuncs()
	TrapKillSignals()

	if !*stderr {
		// Set up log file.
		logname := path.Join(os.Getenv("HOME"), "server.log")
		err := os.Rename(logname, fmt.Sprintf(
			"%s.%s.%d", logname, time.Now().Format("2006-01-02"), time.Now().Unix()))
		if err != nil && !os.IsNotExist(err) {
			log.Fatalln(err)
		}
		logfile, err := os.Create(logname)
		checkInitError(err)
		app.RunAtExit(func() { logfile.Close() })
		log.Println("Subsequent log messages will go to disk, not stderr.")
		log.SetOutput(logfile)
	} else {
		log.Println("Logging to stderr.")
	}
	log.Println("Starting server.")

	// Static files.
	pwd, err := os.Getwd()
	checkInitError(err)
	http.Handle("/", http.FileServer(http.Dir(path.Join(pwd, "static-files"))))

	// Special handler for static graphs.
	http.Handle("/graph/", http.StripPrefix("/graph/", http.FileServer(http.Dir(app.TempGraphDir))))

	// Visualization builder.
	http.HandleFunc("/visualize", visualizeHandler)

	// Loading page.
	template.Must(template.ParseFiles("templates/loading.html")) // Sanity check.
	http.HandleFunc("/loading", loadingHandler)

	// Dashboard page.
	template.Must(template.ParseFiles("templates/dashboard.html")) // Sanity check.
	http.HandleFunc("/dashboard", dashboardHandler)

	// Contact info signup.
	smtpPassword = os.Getenv("EMAIL_CHARTS_SMTP_PASSWORD")
	if smtpPassword == "" {
		log.Println("Must set EMAIL_CHARTS_SMTP_PASSWORD")
		return
	}
	http.HandleFunc("/staytuned", signupHandler)
	ok := func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	}
	http.HandleFunc("/_ah/start", ok)
	http.HandleFunc("/_ah/health", ok)

	log.Printf(http.ListenAndServe(*httpAddr, Log(http.DefaultServeMux)).Error())
}

func checkInitError(err error) {
	if err != nil {
		log.Panicln(err.Error())
	}
}
