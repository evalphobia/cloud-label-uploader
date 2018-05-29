package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/mkideal/cli"
)

// download command
type downloadT struct {
	cli.Helper
	File        string `cli:"*f,file" usage:"image list file --file='/path/to/dir/input.csv'"`
	ColumnName  string `cli:"*n,name" usage:"column name for filename --name='name'"`
	ColumnLabel string `cli:"*l,label" usage:"column name for label --label='group'"`
	ColumnURL   string `cli:"*u,url" usage:"column name for URL --url='path'"`
	Parallel    int    `cli:"p,parallel" usage:"parallel number --parallel=2" dft:"2"`
	OutputDir   string `cli:"o,out" usage:"outout dir --out='/path/to/dir/out'"`
}

var downloader = &cli.Command{
	Name: "download",
	Desc: "Download files from --file csv",
	Argv: func() interface{} { return new(downloadT) },
	Fn:   execOCR,
}

func execOCR(ctx *cli.Context) error {
	argv := ctx.Argv().(*downloadT)
	maxReq := make(chan struct{}, argv.Parallel)

	f, err := NewCSVHandler(argv.File)
	if err != nil {
		return err
	}

	colName := argv.ColumnName
	colLabel := argv.ColumnLabel
	colURL := argv.ColumnURL
	err = f.checkHeaders(colName, colLabel, colURL)
	if err != nil {
		return err
	}

	outputDir := argv.OutputDir
	err = makeDir(outputDir)
	if err != nil {
		return err
	}

	dirMap := make(map[string]struct{})

	var wg sync.WaitGroup
	var counter uint64
	for {
		line, err := f.Read()
		if err != nil {
			return err
		}
		if len(line) == 0 {
			break
		}

		wg.Add(1)
		go func(line map[string]string) {
			maxReq <- struct{}{}
			defer func() {
				<-maxReq
				wg.Done()
			}()

			num := atomic.AddUint64(&counter, 1)
			fmt.Printf("exec #: [%d]\n", num)

			url := line[colURL]
			dir := fmt.Sprintf("%s/%s", outputDir, line[colLabel])
			if _, ok := dirMap[dir]; !ok {
				dirMap[dir] = struct{}{}
				err := makeDir(dir)
				if err != nil {
					fmt.Printf("[ERRORL:mkdir] #=[%d], dir=[%s], err=[%s]\n", num, dir, err)
					return
				}
			}

			name := getFileName(line[colName], url)
			filepath := fmt.Sprintf("%s/%s", dir, name)
			if isFileExist(filepath) {
				fmt.Printf("[SKIP] already exists #=[%d], filepath=[%s]\n", num, filepath)
				return
			}

			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("[ERRORL:http] #=[%d], url=[%s], err=[%s]\n", num, url, err)
				return
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("[ERROR:ioutil] #=[%d], url=[%s], err=[%s]\n", num, url, err)
				return
			}

			fp, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				fmt.Printf("[ERROR:OpenFile] #=[%d], filepath=[%s], err=[%s]\n", num, filepath, err)
				return
			}

			defer fp.Close()
			fp.Write(body)
		}(line)
	}

	wg.Wait()
	return nil
}

func getFileName(name, url string) string {
	ext := strings.Split(filepath.Ext(url), "?")[0]
	return name + ext
}
