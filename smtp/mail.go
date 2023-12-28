package main

import (
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	smtp "github.com/emersion/go-smtp"
)

// The Backend implements SMTP server methods
type Backend struct{}

func (bkd *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{}, nil
}

// A Session is returned after EHLO
type Session struct {
	Messages []Message
}

func (s *Session) AuthPlain(_, _ string) error {
	return smtp.ErrAuthUnsupported
}

func (s *Session) Mail(from string, _ *smtp.MailOptions) error {
	s.Messages = append(s.Messages, Message{From: from})
	return nil
}

func (s *Session) Rcpt(to string, _ *smtp.RcptOptions) error {
	msg := s.Messages[len(s.Messages)-1]
	msg.Recipient = to
	s.Messages[len(s.Messages)-1] = msg
	return nil
}

func (s *Session) Data(r io.Reader) error {
	if b, err := io.ReadAll(r); err != nil {
		return err
	} else {
		msg := s.Messages[len(s.Messages)-1]
		msg.Data = b
		s.Messages[len(s.Messages)-1] = msg
	}
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	for _, m := range s.Messages {
		for _, domain := range strings.Split(config.mxDomains, ",") {
			if strings.HasSuffix(m.Recipient, domain) {
				log.Printf("FROM: %v TO: %v MESSSAGE: %v\n", m.From, m.Recipient, string(m.Data))
				go config.blobClient.Put("mail/"+url.QueryEscape(domain)+"/"+url.QueryEscape(time.Now().String()), m.Data)
			}
		}
	}
	return nil
}

// A Message is a single message to be stored
type Message struct {
	Recipient string
	From      string
	Data      []byte
}
