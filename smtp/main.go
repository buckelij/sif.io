package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	smtp "github.com/emersion/go-smtp"
)

// The Backend implements SMTP server methods
type Backend struct{}

func (bkd *Backend) Login(_ *smtp.ConnectionState, _ string, _ string) (smtp.Session, error) {
	return &Session{}, smtp.ErrAuthUnsupported
}

func (bkd *Backend) AnonymousLogin(_ *smtp.ConnectionState) (smtp.Session, error) {
	return &Session{}, nil
}

// A Session is returned after EHLO
type Session struct {
	Messages []Message
}

func (s *Session) Mail(from string, opts smtp.MailOptions) error {
	s.Messages = append(s.Messages, Message{From: from})
	return nil
}

func (s *Session) Rcpt(to string) error {
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
		if strings.HasSuffix(m.Recipient, "sif.io") {
			log.Printf("FROM: %v TO: %v MESSSAGE: %v\n", m.From, m.Recipient, string(m.Data))
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

/*
EHLO localhost
MAIL FROM:<root@example.com>
RCPT TO:<toor@example.com>
DATA
Hey <3
.
*/
func main() {
	blobAccount := os.Getenv("BLOB_ACCOUNT")
	blobContainer := os.Getenv("BLOB_CONTAINER")
	blobKey := os.Getenv("BLOB_KEY")
	cred, err := azblob.NewSharedKeyCredential(blobAccount, blobKey)
	if err != nil {
		panic("blob storage credential error")
	}
	blobClient, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", blobAccount), cred, nil)
	if err != nil {
		panic("blob client error")
	}
	_, err = blobClient.UploadStream(context.TODO(),
		blobContainer,
		"ping",
		strings.NewReader("pong"),
		&azblob.UploadStreamOptions{})
	if err != nil {
		log.Println("failed to upload")
	}

	be := &Backend{}

	s := smtp.NewServer(be)

	s.Addr = "0.0.0.0:1025"
	s.Domain = "mx.sif.io"
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024 * 5
	s.MaxRecipients = 500
	s.AuthDisabled = true

	log.Println("Starting server at", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
