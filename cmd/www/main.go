package main

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/buckelij/sif.io/internal/blob"
	"github.com/buckelij/sif.io/internal/ssl"
	"github.com/buckelij/sif.io/internal/www"
)

var config = struct {
	BlobAccount   string
	BlobContainer string
	BlobKey       string
	NoTls         string
}{
	BlobAccount:   os.Getenv("BLOB_ACCOUNT"),
	BlobContainer: os.Getenv("BLOB_CONTAINER"),
	BlobKey:       os.Getenv("BLOB_KEY"),
	NoTls:         os.Getenv("NO_TLS"),
}

func main() {
	log.Println("starting")

	redirMux := http.NewServeMux()
	redirMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		targetUrl := url.URL{Scheme: "https", Host: r.Host, Path: r.URL.Path}
		http.Redirect(w, r, targetUrl.String(), http.StatusMovedPermanently)
	})
	go func() { log.Println(http.ListenAndServe(":8080", redirMux)) }()
	log.Println("started redirect listener")

	http.HandleFunc("/", www.Index(www.IndexHtml))
	http.HandleFunc("/resume", www.Page(www.ResumeHtml))
	if config.NoTls != "" {
		log.Fatal(http.ListenAndServe(":8443", nil))
	} else {
		blobClient, err := blob.NewAzureBlobClient(config.BlobAccount, config.BlobContainer, config.BlobKey)
		if err != nil {
			panic("failed to create blob client")
		}
		err = blobClient.Put("pingwww", []byte("pong"))
		if err != nil {
			log.Println("failed to upload ping", err)
		}
		s := &http.Server{
			Addr:      ":8443",
			TLSConfig: ssl.NewSSLmanager(blobClient).TLSConfig(),
		}
		log.Fatal(s.ListenAndServeTLS("", ""))
	}
}
