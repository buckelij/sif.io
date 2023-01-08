package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type BlobClient interface {
	Put(string, []byte) error
	Get(string) ([]byte, error)
	List(string) ([]string, error)
}

type azureBlobClient struct {
	client    *azblob.Client
	container string
}

func NewAzureBlobClient(account string, container string, key string) (BlobClient, error) {
	cred, err := azblob.NewSharedKeyCredential(account, key)
	if err != nil {
		return &azureBlobClient{}, err
	}
	c, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", account), cred, nil)
	if err != nil {
		return &azureBlobClient{}, err
	}

	return &azureBlobClient{c, container}, nil
}

func (c *azureBlobClient) Put(oid string, data []byte) error {
	_, err := c.client.UploadBuffer(context.TODO(), c.container, oid, data, &azblob.UploadBufferOptions{})
	if err != nil {
		log.Println("failed to upload")
	}
	return err
}

func (c *azureBlobClient) Get(oid string) ([]byte, error) {
	log.Print(oid)
	return []byte{}, nil
}

func (c *azureBlobClient) List(prefix string) ([]string, error) {
	log.Print(prefix)
	return []string{}, nil
}
