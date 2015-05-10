package app

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/mail"

	"github.com/jeremyschlatter/email-charts/Godeps/_workspace/src/github.com/mxk/go-imap/imap"
)

type oauthSASL struct {
	user, token string
}

func (o oauthSASL) Start(s *imap.ServerInfo) (string, []byte, error) {
	return "XOAUTH2", []byte(fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", o.user, o.token)), nil
}

func (o oauthSASL) Next(challenge []byte) ([]byte, error) {
	return nil, errors.New("Challenge shouldn't be issued.")
}

func fetchAllHeaders(user, authToken, box string) ([]mail.Header, error) {
	c, err := imap.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		return nil, err
	}
	_, err = c.Auth(oauthSASL{user, authToken})
	if err != nil {
		return nil, err
	}
	c.Select(box, true)
	set, _ := imap.NewSeqSet("1:*")
	cmd, err := imap.Wait(c.Fetch(set, "BODY.PEEK[HEADER.FIELDS (DATE)]"))
	if err != nil {
		return nil, err
	}
	headers := make([]mail.Header, 0, len(cmd.Data))
	parseFailures := 0
	for _, rsp := range cmd.Data {
		header := imap.AsBytes(rsp.MessageInfo().Attrs["BODY[HEADER.FIELDS (DATE)]"]) // no peek
		if msg, _ := mail.ReadMessage(bytes.NewReader(header)); msg != nil {
			headers = append(headers, msg.Header)
		} else {
			parseFailures++
		}
	}
	log.Printf("Finished processing %d emails with %d parse failures.\n", len(headers), parseFailures)
	return headers, nil
}
