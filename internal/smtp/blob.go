package smtp

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type BlobClient interface {
	Put(string, []byte) error
	Get(string) ([]byte, error)
	ListMail() ([]string, error)
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
	s, err := c.client.DownloadStream(context.TODO(), c.container, oid, &azblob.DownloadStreamOptions{})
	if err != nil {
		return []byte{}, err
	}
	b := bytes.Buffer{}
	retryReader := s.NewRetryReader(context.TODO(), &azblob.RetryReaderOptions{})
	_, err = b.ReadFrom(retryReader)
	if err != nil {
		return []byte{}, err
	}
	defer retryReader.Close()
	return b.Bytes(), err
}

// Lists all mail blobs, which won't be too many right :|
func (c *azureBlobClient) ListMail() ([]string, error) {
	prefix := "mail/"
	lister := c.client.NewListBlobsFlatPager(c.container, &azblob.ListBlobsFlatOptions{Prefix: &prefix})
	blobs := make([]string, 1)
	for lister.More() {
		page, err := lister.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("azureBlobClient List: %v", err)
			return []string{}, err
		}
		for _, blob := range page.Segment.BlobItems {
			blobs = append(blobs, *blob.Name)
		}
	}
	return blobs, nil
}
