package main

import (
	"context"
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
	return s.BlobClient.Get("certs/" + url.QueryEscape(key))
}

func (s SSLblobCache) Put(ctx context.Context, key string, data []byte) error {
	return s.BlobClient.Put("certs/"+url.QueryEscape(key), data)
}

func (s SSLblobCache) Delete(ctx context.Context, key string) error {
	return nil
}
