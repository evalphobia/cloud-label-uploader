package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/mkideal/cli"
)

// list command
type listT struct {
	cli.Helper
	Input          string `cli:"*i,input" usage:"image dir path --input='/path/to/image_dir'"`
	Output         string `cli:"*o,output" usage:"output TSV file path --output='./output.csv'" dft:"./output.csv"`
	IncludeAllType bool   `cli:"a,all" usage:"use all files"`
	Type           string `cli:"t,type" usage:"comma separate file extensions --type='jpg,jpeg,png,gif'" dft:"jpg,jpeg,png,gif"`
	PathPrefix     string `cli:"p,prefix" usage:"prefix for file path --prefix='gs://<your-bucket-name>'" dft:""`
}

var list = &cli.Command{
	Name: "list",
	Desc: "Create csv list file from --output dir",
	Argv: func() interface{} { return new(listT) },
	Fn:   execList,
}

var (
	baseDir    string
	pathPrefix string
)

func execList(ctx *cli.Context) error {
	argv := ctx.Argv().(*listT)

	f, err := NewFileHandler(argv.Output)
	if err != nil {
		return err
	}
	types := newFileType(strings.Split(argv.Type, ","))
	if argv.IncludeAllType {
		types.setIncludeAll(argv.IncludeAllType)
	}

	pathPrefix = argv.PathPrefix
	baseDir = fmt.Sprintf("%s/", filepath.Clean(argv.Input))
	result := getFilesFromDir(baseDir, types)
	return f.WriteAll(result)
}

func getFilesFromDir(dir string, types fileType) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() {
			paths = append(paths, getFilesFromDir(filepath.Join(dir, fileName), types)...)
			continue
		}

		if !types.isTarget(fileName) {
			continue
		}

		label := strings.TrimPrefix(dir, baseDir)
		path := getURLPath(pathPrefix, path.Join(label, fileName))
		paths = append(paths, fmt.Sprintf("%s,%s", path, label))
	}

	return paths
}

func getURLPath(prefix, filepath string) string {
	u, _ := url.Parse(prefix)
	u.Path = path.Join(u.Path, filepath)
	return u.String()
}
