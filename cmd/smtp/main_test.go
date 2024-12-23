package main

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"sync"
	"testing"

	sifsmtp "github.com/buckelij/sif.io/internal/smtp"
)

type TestBlobClient struct {
	wg       sync.WaitGroup // create a wait group, this will allow you to block later
	uploaded []string
	gets     [][]byte // stub values to be returned
}

func (c *TestBlobClient) Put(oid string, data []byte) error {
	c.uploaded = append(c.uploaded, oid)
	c.wg.Done()
	return nil
}

func (c *TestBlobClient) Get(oid string) ([]byte, error) {
	v := c.gets[0]
	c.gets = c.gets[1:]
	return v, nil
}

func (c *TestBlobClient) ListMail() ([]string, error) {
	return []string{}, nil
}

func TestStoresMail(t *testing.T) {
	testBlobClient := &TestBlobClient{}
	s := newServer(&sifsmtp.Backend{
		ListenAddress: "0.0.0.0:1025",
		Domain:        "mx.sif.io",
		MxDomains:     "sif.io",
		BlobAccount:   "blob-account",
		BlobContainer: "blob-container",
		BlobKey:       "blob-secret",
		BlobClient:    testBlobClient,
	})

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	go s.Serve(l)

	testBlobClient.wg.Add(1)

	c, _ := smtp.Dial(l.Addr().String())
	c.Mail("sender@example.org")
	c.Rcpt("recipient@sif.io")
	wc, _ := c.Data()
	fmt.Fprintf(wc, "This is the email body")
	wc.Close()
	c.Quit()

	testBlobClient.wg.Wait()

	if len(testBlobClient.uploaded) != 1 {
		t.Error("mail did not store")
	}

	if !strings.HasPrefix(testBlobClient.uploaded[0], "mail/sif.io") {
		t.Error("mail did not store with expected blob prefix")
	}
}
