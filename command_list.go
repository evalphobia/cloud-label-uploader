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
	Output         string `cli:"*o,output" usage:"output CSV file path --output='./output.csv'" dft:"./output.csv"`
	IncludeAllType bool   `cli:"a,all" usage:"use all files"`
	Type           string `cli:"t,type" usage:"comma separate file extensions --type='jpg,jpeg,png,gif'" dft:"jpg,jpeg,png,gif"`
	Format         string `cli:"f,format" usage:"set output format --format='[csv,sagemaker]'" dft:"csv"`
	PathPrefix     string `cli:"*p,prefix" usage:"prefix for file path --prefix='gs://<your-bucket-name>'" dft:""`
}

var list = &cli.Command{
	Name: "list",
	Desc: "Create list file from --input dir images",
	Argv: func() interface{} { return new(listT) },
	Fn:   execList,
}

var (
	baseDir    string
	pathPrefix string
)

func execList(ctx *cli.Context) error {
	argv := ctx.Argv().(*listT)

	r := newListRunner(*argv)
	formatter, err := createListFormat(argv.Format)
	if err != nil {
		return err
	}
	r.Formatter = formatter
	return r.Run()
}

type ListRunner struct {
	// parameters
	Input          string
	Output         string
	IncludeAllType bool
	Type           string
	Format         string
	PathPrefix     string

	Formatter formatter
}

func newListRunner(p listT) ListRunner {
	return ListRunner{
		Input:          p.Input,
		Output:         p.Output,
		IncludeAllType: p.IncludeAllType,
		Type:           p.Type,
		Format:         p.Format,
		PathPrefix:     p.PathPrefix,
	}
}

func (r *ListRunner) Run() error {
	f, err := NewFileHandler(r.Output)
	if err != nil {
		return err
	}

	types := newFileType(strings.Split(r.Type, ","))
	if r.IncludeAllType {
		types.setIncludeAll(r.IncludeAllType)
	}

	pathPrefix = r.PathPrefix
	baseDir = fmt.Sprintf("%s/", filepath.Clean(r.Input))
	result, err := r.GetFilesFromDir(baseDir, types)
	if err != nil {
		return err
	}
	return f.WriteAll(result)
}

func (r *ListRunner) GetFilesFromDir(dir string, types fileType) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() {
			sublist, err := r.GetFilesFromDir(filepath.Join(dir, fileName), types)
			if err != nil {
				return nil, err
			}
			paths = append(paths, sublist...)
			continue
		}

		if !types.isTarget(fileName) {
			continue
		}

		label := strings.TrimPrefix(dir, baseDir)
		path := getURLPath(pathPrefix, path.Join(label, fileName))
		paths = append(paths, r.Formatter.format(path, label))
	}
	return paths, nil
}

func getURLPath(prefix, filepath string) string {
	u, _ := url.Parse(prefix)
	u.Path = path.Join(u.Path, filepath)
	return u.String()
}
