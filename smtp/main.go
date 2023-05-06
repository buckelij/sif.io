// mail receiver and web interface
// mail is stored in blob storage under the `mail/` prefix
// webmail is authenticated against blob storage hashes under `bcrypt/<username>` keys
// credentials can be generated with e.g. `go run . genpass passw0rd`

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	smtp "github.com/emersion/go-smtp"
	"golang.org/x/crypto/bcrypt"
)

type configuration struct {
	listenAddress string
	domain        string
	mxDomains     string
	blobAccount   string
	blobContainer string
	blobKey       string
	blobClient    BlobClient
}

var config configuration

/*
EHLO localhost
MAIL FROM:<root@example.com>
RCPT TO:<toor@example.com>
DATA
Hey <3
.
*/
func newServer() *smtp.Server {
	be := &Backend{}

	s := smtp.NewServer(be)

	s.Addr = config.listenAddress
	s.Domain = config.domain
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024 * 5
	s.MaxRecipients = 500
	s.AuthDisabled = true

	return s
}

func main() {
	if len(os.Args) > 2 && os.Args[1] == "genpass" {
		v, _ := bcrypt.GenerateFromPassword([]byte(os.Args[2]), bcrypt.DefaultCost)
		fmt.Println(string(v))
		return
	}

	blobClient, err := NewAzureBlobClient(os.Getenv("BLOB_ACCOUNT"), os.Getenv("BLOB_CONTAINER"), os.Getenv("BLOB_KEY"))
	if err != nil {
		panic("failed to create blob client")
	}
	err = blobClient.Put("ping", []byte("pong"))
	if err != nil {
		log.Println("failed to upload ping", err)
	}

	config = configuration{
		"0.0.0.0:1025",
		"mx.sif.io",
		os.Getenv("MX_DOMAINS"),
		os.Getenv("BLOB_ACCOUNT"),
		os.Getenv("BLOB_CONTAINER"),
		os.Getenv("BLOB_KEY"),
		blobClient,
	}
	s := newServer()
	log.Println("Starting server at", s.Addr)

	xsrfSecret := os.Getenv("XSRF_SECRET")
	if xsrfSecret == "" {
		log.Fatal("XSRF_SECRET not set")
	}
	webmailservice := Webmail{XsrfSecret: xsrfSecret, BlobClient: blobClient}
	go webmailservice.ListenAndServeWebmail()

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
