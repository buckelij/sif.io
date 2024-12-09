package main

import (
	"log"
	"net/http"

	"github.com/buckelij/sif.io/internal/www"
)

func main() {
	log.Println("starting")
	http.HandleFunc("/", www.Index(www.IndexHtml))
	http.HandleFunc("/resume", www.Page(www.ResumeHtml))
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
