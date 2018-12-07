package gcs

import (
	"context"

	"github.com/evalphobia/google-api-go-wrapper/config"
	"github.com/evalphobia/google-api-go-wrapper/storage"

	"github.com/evalphobia/csv-file-downloader/provider"
)

const providerName = "gcs"

func init() {
	provider.AddProvider(providerName, newProvider)
}

// GCSClient is client for Google Cloud Storage.
type GCSClient struct {
	*storage.Storage
}

func New() (GCSClient, error) {
	cli, err := storage.New(context.Background(), config.Config{})
	return GCSClient{
		Storage: cli,
	}, err
}

func newProvider() (provider.Provider, error) {
	return New()
}

// CheckBucket checks bucket existence.
func (c GCSClient) CheckBucket(bucketName string) error {
	_, err := c.Storage.Bucket(bucketName).Attrs(context.Background())
	return err
}

// IsExists checks file existence from GCS Bucket.
func (c GCSClient) IsExists(opt provider.FileOption) (isExist bool, err error) {
	return c.Storage.IsExists(storage.ObjectOption{
		BucketName: opt.BucketName,
		Path:       opt.DstPath,
	})
}

// UploadFromLocalFile uploads from local file to GCS Bucket.
func (c GCSClient) UploadFromLocalFile(opt provider.FileOption) error {
	return c.Storage.UploadByFile(opt.SrcPath, storage.ObjectOption{
		BucketName: opt.BucketName,
		Path:       opt.DstPath,
	})
}
