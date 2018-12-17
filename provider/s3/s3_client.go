package s3

import (
	"fmt"
	"os"

	"github.com/evalphobia/aws-sdk-go-wrapper/config"
	"github.com/evalphobia/aws-sdk-go-wrapper/s3"

	"github.com/evalphobia/cloud-label-uploader/provider"
)

const providerName = "s3"

func init() {
	provider.AddProvider(providerName, newProvider)
}

// S3Client is client for AWS S3.
type S3Client struct {
	*s3.S3
}

func New() (S3Client, error) {
	cli, err := s3.New(config.Config{})
	return S3Client{
		S3: cli,
	}, err
}

func newProvider() (provider.Provider, error) {
	return New()
}

// CheckBucket checks bucket existence.
func (c S3Client) CheckBucket(bucketName string) error {
	ok, err := c.S3.IsExistBucket(bucketName)
	switch {
	case err != nil:
		return err
	case !ok:
		return fmt.Errorf("Bucket does not exists: [%s]", bucketName)
	}
	return nil
}

// IsExists checks file existence from GCS Bucket.
func (c S3Client) IsExists(opt provider.FileOption) (isExist bool, err error) {
	b, err := c.S3.GetBucket(opt.BucketName)
	if err != nil {
		return false, err
	}

	return b.IsExists(opt.DstPath), nil
}

// UploadFromLocalFile uploads from local file to S3 Bucket.
func (c S3Client) UploadFromLocalFile(opt provider.FileOption) error {
	b, err := c.S3.GetBucket(opt.BucketName)
	if err != nil {
		return err
	}

	file, err := os.Open(opt.SrcPath)
	if err != nil {
		return err
	}
	obj := s3.NewPutObject(file)
	err = b.PutOne(obj, opt.DstPath, s3.ACLPrivate)
	return err
}
