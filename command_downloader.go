package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/mkideal/cli"
)

// download command
type downloadT struct {
	cli.Helper
	Input       string `cli:"*i,input" usage:"image list file --input='/path/to/dir/input.csv'"`
	ColumnName  string `cli:"*n,name" usage:"column name for filename --name='name'"`
	ColumnLabel string `cli:"*l,label" usage:"column name for label --label='group'"`
	ColumnURL   string `cli:"*u,url" usage:"column name for URL --url='path'"`
	Parallel    int    `cli:"p,parallel" usage:"parallel number --parallel=2" dft:"2"`
	OutputDir   string `cli:"o,output" usage:"outout dir --output='/path/to/dir/'"`
}

var downloader = &cli.Command{
	Name: "download",
	Desc: "Download files from --file csv",
	Argv: func() interface{} { return new(downloadT) },
	Fn:   execDownload,
}

func execDownload(ctx *cli.Context) error {
	argv := ctx.Argv().(*downloadT)
	maxReq := make(chan struct{}, argv.Parallel)

	f, err := NewCSVHandler(argv.Input)
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
	if outputDir == "" {
		outputDir = "."
	}
	err = makeDir(outputDir)
	if err != nil {
		return err
	}

	dirMap := newDirectoryMap()

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
			err := dirMap.Create(dir)
			if err != nil {
				fmt.Printf("[ERRORL:mkdir] #=[%d], dir=[%s], err=[%s]\n", num, dir, err)
				return
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

// get file name with extension.
func getFileName(name, uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return name
	}

	ext := filepath.Ext(u.Path)
	return name + ext
}

// to create new dir for the label.
type directoryMap struct {
	dataMu sync.RWMutex
	data   map[string]struct{}
}

func newDirectoryMap() directoryMap {
	return directoryMap{
		data: make(map[string]struct{}),
	}
}

func (m directoryMap) Create(key string) error {
	if m.has(key) {
		return nil
	}

	m.dataMu.Lock()
	defer m.dataMu.Unlock()
	m.data[key] = struct{}{}
	return makeDir(key)
}

func (m directoryMap) has(key string) bool {
	m.dataMu.RLock()
	defer m.dataMu.RUnlock()
	_, ok := m.data[key]
	return ok
}
