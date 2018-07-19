package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/evalphobia/google-api-go-wrapper/config"
	"github.com/evalphobia/google-api-go-wrapper/storage"
	"github.com/mkideal/cli"
)

var _ = config.Config{}

// uploader command
type uploaderT struct {
	cli.Helper
	Input          string `cli:"*i,input" usage:"image dir path --input='/path/to/image_dir'"`
	Type           string `cli:"t,type" usage:"comma separate file extensions --type='jpg,jpeg,png,gif'" dft:"jpg,jpeg,png,gif"`
	IncludeAllType bool   `cli:"a,all" usage:"use all files"`
	Bucket         string `cli:"*b,bucket" usage:"bucket name of GCS --bucket='<your-bucket-name>'"`
	PathPrefix     string `cli:"*d,prefix" usage:"prefix for GCS --prefix='foo/bar'"`
	Parallel       int    `cli:"p,parallel" usage:"parallel number --parallel=2" dft:"2"`
}

var uploader = &cli.Command{
	Name: "uploader",
	Desc: "Upload files to GCS from --input dir",
	Argv: func() interface{} { return new(uploaderT) },
	Fn:   execUpload,
}

func execUpload(ctx *cli.Context) error {
	argv := ctx.Argv().(*uploaderT)

	// create GCS cient from env vars
	cli, err := storage.New(context.Background(), config.Config{})
	if err != nil {
		panic(err)
	}
	types := newFileType(strings.Split(argv.Type, ","))
	if argv.IncludeAllType {
		types.setIncludeAll(argv.IncludeAllType)
	}

	u := Uploader{
		GCSClient:  cli,
		FileTypes:  types,
		BaseDir:    fmt.Sprintf("%s/", filepath.Clean(argv.Input)),
		Bucket:     argv.Bucket,
		PathPrefix: strings.TrimLeft(argv.PathPrefix, "/"),
		maxReq:     make(chan struct{}, argv.Parallel),
	}
	u.UploadFilesFromDir(u.BaseDir)
	u.wg.Wait()
	return nil
}

type Uploader struct {
	GCSClient  *storage.Storage
	FileTypes  fileType
	Bucket     string
	PathPrefix string
	BaseDir    string

	wg      sync.WaitGroup
	maxReq  chan struct{}
	counter uint64
}

func (u *Uploader) UploadFilesFromDir(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() {
			u.UploadFilesFromDir(filepath.Join(dir, fileName))
			continue
		}

		if !u.FileTypes.isTarget(fileName) {
			continue
		}

		u.wg.Add(1)
		go func(dir, fileName string) {
			u.maxReq <- struct{}{}
			defer func() {
				<-u.maxReq
				u.wg.Done()
			}()

			num := atomic.AddUint64(&u.counter, 1)
			fmt.Printf("exec #: [%d]\n", num)

			skip, err := u.upload(dir, fileName)
			switch {
			case err != nil:
				fmt.Printf("[ERROR]: #=[%d] path=[%s] error=[%s]\n", num, filepath.Join(dir, fileName), err.Error())
			case skip:
				fmt.Printf("[SKIP] already exists #=[%d], filepath=[%s]\n", num, filepath.Join(dir, fileName))
			}
		}(dir, fileName)
	}
}

func (u *Uploader) upload(dir, fileName string) (skip bool, err error) {
	label := u.getLabel(dir)
	objectPath := path.Join(u.PathPrefix, label, fileName)

	// check file existance
	ok, err := u.GCSClient.IsExists(storage.ObjectOption{
		BucketName: u.Bucket,
		Path:       objectPath,
	})
	switch {
	case err != nil:
		return false, err
	case ok:
		return true, nil
	}

	// upload file
	return false, u.GCSClient.UploadByFile(filepath.Join(dir, fileName), storage.ObjectOption{
		BucketName: u.Bucket,
		Path:       objectPath,
	})
}

func (u Uploader) getLabel(path string) string {
	return strings.TrimPrefix(path, u.BaseDir)
}
