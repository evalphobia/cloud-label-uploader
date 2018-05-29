package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/mkideal/cli"
)

// list command
type listT struct {
	cli.Helper
	Input      string `cli:"*i,input" usage:"image dir path --input='/path/to/image_dir'"`
	Output     string `cli:"*o,output" usage:"output TSV file path --output='./output.csv'" dft:"./output.csv"`
	Type       string `cli:"t,type" usage:"comma separate file extensions --type='jpg,jpeg,png,gif'" dft:"jpg,jpeg,png,gif"`
	Headers    string `cli:"headers" usage:"comma separate header columns --headers='label,path'" dft:"label,path"`
	PathPrefix string `cli:"p,prefix" usage:"prefix for file path --prefix='gs://<your-bucket-name>'" dft:""`
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
	types      map[string]struct{}
)

func execList(ctx *cli.Context) error {
	argv := ctx.Argv().(*listT)

	types = make(map[string]struct{})
	for _, s := range strings.Split(argv.Type, ",") {
		types["."+strings.TrimSpace(s)] = struct{}{}
	}

	f, err := NewFileHandler(argv.Output)
	if err != nil {
		return err
	}

	pathPrefix = argv.PathPrefix
	baseDir = fmt.Sprintf("%s/", filepath.Clean(argv.Input))
	result := append([]string{argv.Headers}, getFilesFromDir(baseDir)...)
	return f.WriteAll(result)
}

func getFilesFromDir(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() {
			paths = append(paths, getFilesFromDir(filepath.Join(dir, fileName))...)
			continue
		}

		if _, ok := types[strings.ToLower(filepath.Ext(fileName))]; !ok {
			continue
		}

		d := strings.TrimPrefix(dir, baseDir)
		paths = append(paths, fmt.Sprintf("%s,%s", d, filepath.Join(pathPrefix, d, fileName)))
	}

	return paths
}
