package main

import (
	"context"
	"log"
	"net/url"

	"golang.org/x/crypto/acme/autocert"
)

func NewSSLmanager(c BlobClient) *autocert.Manager {
	return &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      SSLblobCache{c},
		HostPolicy: autocert.HostWhitelist("www.sif.io", "webmail.sif.io"),
	}
}

type SSLblobCache struct {
	BlobClient BlobClient
}

var _ autocert.Cache = SSLblobCache{}

func (s SSLblobCache) Get(ctx context.Context, key string) ([]byte, error) {

	d, err := s.BlobClient.Get("certs/" + url.QueryEscape(key))
	if err != nil {
		return []byte{}, autocert.ErrCacheMiss
	}

	return d, err
}

func (s SSLblobCache) Put(ctx context.Context, key string, data []byte) error {
	log.Println("saving certificate")
	return s.BlobClient.Put("certs/"+url.QueryEscape(key), data)
}

func (s SSLblobCache) Delete(ctx context.Context, key string) error {
	return nil
}
