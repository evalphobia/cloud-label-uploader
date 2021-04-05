package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/mkideal/cli"

	"github.com/evalphobia/cloud-label-uploader/provider"
	_ "github.com/evalphobia/cloud-label-uploader/provider/gcs"
	_ "github.com/evalphobia/cloud-label-uploader/provider/s3"
)

// uploader command
type uploaderT struct {
	cli.Helper
	Input          string `cli:"*i,input" usage:"image dir path --input='/path/to/image_dir'"`
	Type           string `cli:"t,type" usage:"comma separate file extensions --type='jpg,jpeg,png,gif'" dft:"jpg,jpeg,png,gif"`
	IncludeAllType bool   `cli:"a,all" usage:"use all files"`
	InputLabelFile string `cli:"l,label" usage:"label file for training (outputted CSV file) --label='/path/to/output.csv'"`
	CloudProvider  string `cli:"*c,provider" usage:"cloud provider name for the bucket --provider='[s3,gcs]'"`
	Bucket         string `cli:"*b,bucket" usage:"bucket name of S3/GCS --bucket='<your-bucket-name>'"`
	PathPrefix     string `cli:"*p,prefix" usage:"prefix for S3/GCS --prefix='foo/bar'"`
	Parallel       int    `cli:"m,parallel" usage:"parallel number (multiple upload) --parallel=2" dft:"2"`
}

var uploader = &cli.Command{
	Name: "upload",
	Desc: "Upload files to Cloud Bucket(S3, GCS) from --input dir",
	Argv: func() interface{} { return new(uploaderT) },
	Fn:   execUpload,
}

func execUpload(ctx *cli.Context) error {
	argv := ctx.Argv().(*uploaderT)

	// create Cloud Provider client from env vars
	cli, err := provider.Create(argv.CloudProvider)
	if err != nil {
		panic(err)
	}
	if err := cli.CheckBucket(argv.Bucket); err != nil {
		panic(err)
	}

	types := newFileType(strings.Split(argv.Type, ","))
	if argv.IncludeAllType {
		types.setIncludeAll(argv.IncludeAllType)
	}

	u := Uploader{
		Provider:   cli,
		FileTypes:  types,
		BaseDir:    fmt.Sprintf("%s/", filepath.Clean(argv.Input)),
		Bucket:     argv.Bucket,
		PathPrefix: strings.TrimLeft(argv.PathPrefix, "/"),
		maxReq:     make(chan struct{}, argv.Parallel),
	}
	if argv.InputLabelFile != "" {
		u.UploadFileFromPath(argv.InputLabelFile)
	}
	u.UploadFilesFromDir(u.BaseDir)
	u.wg.Wait()
	return nil
}

type Uploader struct {
	Provider   provider.Provider
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
			fmt.Printf("exec #%d: [%s] [%s]\n", num, dir, fileName)

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

func (u *Uploader) UploadFileFromPath(path string) {
	_, err := u.upload(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		panic(err)
	}
}

func (u *Uploader) upload(dir, fileName string) (skip bool, err error) {
	label := u.getLabel(dir)
	objectPath := path.Join(u.PathPrefix, label, fileName)

	// check file existance
	ok, err := u.Provider.IsExists(provider.FileOption{
		BucketName: u.Bucket,
		DstPath:    objectPath,
	})
	switch {
	case err != nil:
		return false, err
	case ok:
		return true, nil
	}

	// upload file
	return false, u.Provider.UploadFromLocalFile(provider.FileOption{
		SrcPath:    filepath.Join(dir, fileName),
		BucketName: u.Bucket,
		DstPath:    objectPath,
	})
}

func (u Uploader) getLabel(path string) string {
	return strings.TrimPrefix(path, u.BaseDir)
}
