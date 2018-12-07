package provider

import (
	"fmt"
	"strings"
)

var providerGenerator = map[string]func() (Provider, error){}

type Provider interface {
	CheckBucket(bucketName string) error
	IsExists(FileOption) (isExist bool, err error)
	UploadFromLocalFile(FileOption) error
}

type FileOption struct {
	SrcPath    string
	BucketName string
	DstPath    string
}

// AddProvider adds the Provider constructor to the list.
func AddProvider(providerName string, fn func() (Provider, error)) {
	providerGenerator[providerName] = fn
}

// Create creates the Provider from the list.
func Create(providerName string) (Provider, error) {
	if fn, ok := providerGenerator[strings.ToLower(providerName)]; ok {
		return fn()
	}
	return nil, fmt.Errorf("Unknown Provider: [%s]", providerName)
}
