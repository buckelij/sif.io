// mail receiver and web interface
// mail is stored in blob storage under the `mail/` prefix
// webmail is authenticated against blob storage hashes under `bcrypt/<username>` keys
// credentials can be generated with e.g. `go run . genpass passw0rd`
// set ENV NO_TLS to disable SSL

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/buckelij/sif.io/internal/smtp"
	gosmtp "github.com/emersion/go-smtp"
	"golang.org/x/crypto/bcrypt"
)

var config = struct {
	MxDomains     string
	BlobAccount   string
	BlobContainer string
	BlobKey       string
	XsrfSecret    string
	NoTls         string
}{
	MxDomains:     os.Getenv("MX_DOMAINS"),
	BlobAccount:   os.Getenv("BLOB_ACCOUNT"),
	BlobContainer: os.Getenv("BLOB_CONTAINER"),
	BlobKey:       os.Getenv("BLOB_KEY"),
	XsrfSecret:    os.Getenv("XSRF_SECRET"),
	NoTls:         os.Getenv("NO_TLS"),
}

/*
EHLO localhost
MAIL FROM:<root@example.com>
RCPT TO:<toor@example.com>
DATA
Hey <3
.
*/
func newServer(be *smtp.Backend) *gosmtp.Server {
	s := gosmtp.NewServer(be)

	s.Addr = be.ListenAddress
	s.Domain = be.Domain
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024 * 5
	s.MaxRecipients = 500

	return s
}

func main() {
	log.Println("starting")
	if len(os.Args) > 2 && os.Args[1] == "genpass" {
		v, _ := bcrypt.GenerateFromPassword([]byte(os.Args[2]), bcrypt.DefaultCost)
		fmt.Println(string(v))
		return
	}

	blobClient, err := smtp.NewAzureBlobClient(config.BlobAccount, config.BlobContainer, config.BlobKey)
	if err != nil {
		panic("failed to create blob client")
	}
	err = blobClient.Put("ping", []byte("pong"))
	if err != nil {
		log.Println("failed to upload ping", err)
	}

	s := newServer(&smtp.Backend{
		ListenAddress: "0.0.0.0:1025",
		Domain:        "mx.sif.io",
		MxDomains:     config.MxDomains,
		BlobAccount:   config.BlobAccount,
		BlobContainer: config.BlobContainer,
		BlobKey:       config.BlobKey,
		BlobClient:    blobClient,
	})
	log.Println("Starting server at", s.Addr)

	xsrfSecret := config.XsrfSecret
	if xsrfSecret == "" {
		log.Fatal("XSRF_SECRET not set")
	}
	webmailservice := smtp.NewWebMailer(xsrfSecret, blobClient, config.NoTls != "")
	go webmailservice.ListenAndServeWebmail()

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
