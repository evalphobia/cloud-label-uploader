package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mkideal/cli"
)

// vott command
type vottT struct {
	cli.Helper
	JSONDir     string `cli:"*j,json" usage:"VoTT json results dir path --image='/path/to/vott_json_dir'"`
	Output      string `cli:"*o,output" usage:"output CSV file path --output='./output.csv'" dft:"./output.csv"`
	PathPrefix  string `cli:"*p,prefix" usage:"prefix for file path --prefix='gs://<your-bucket-name>'" dft:"gs://"`
	IsRecursive bool   `cli:"r,recursive" usage:"read files in sub directories" dft:"false"`
}

var vott = &cli.Command{
	Name: "vott",
	Desc: "Create object-detection list file from VoTT results",
	Argv: func() interface{} { return new(vottT) },
	Fn:   execVott,
}

func execVott(ctx *cli.Context) error {
	argv := ctx.Argv().(*vottT)

	r := newVottRunner(*argv)
	return r.Run()
}

type VottRunner struct {
	// parameters
	JSONDir     string
	Output      string
	PathPrefix  string
	IsRecursive bool

	Formatter formatter
}

func newVottRunner(p vottT) VottRunner {
	return VottRunner{
		JSONDir:     p.JSONDir,
		Output:      p.Output,
		PathPrefix:  p.PathPrefix,
		IsRecursive: p.IsRecursive,
	}
}

func (r *VottRunner) Run() error {
	// try to open before starting process
	f, err := NewFileHandler(r.Output)
	if err != nil {
		return err
	}
	r.Formatter = &automlObjectDetectionFormatter{
		pathPrefix: r.PathPrefix,
	}

	// get json file list
	baseDir = fmt.Sprintf("%s/", filepath.Clean(r.JSONDir))
	jsonFiles, err := r.FindJSONFilesFromDir(baseDir)
	if err != nil {
		return err
	}

	// read VoTT JSON and convert to AutoML format.
	results, err := r.ReadDataFromJSONFiles(jsonFiles)
	// save to CSV file
	return f.WriteAll(results)
}

func (r VottRunner) FindJSONFilesFromDir(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	types := newFileType([]string{"json"})

	var list []string
	for _, file := range files {
		fileName := file.Name()
		if r.IsRecursive && file.IsDir() {
			sublist, err := r.FindJSONFilesFromDir(filepath.Join(dir, fileName))
			if err != nil {
				return nil, err
			}
			list = append(list, sublist...)
			continue
		}

		if !types.isTarget(fileName) {
			continue
		}

		list = append(list, filepath.Join(dir, fileName))
	}
	return list, nil
}

func (r VottRunner) ReadDataFromJSONFiles(list []string) ([]string, error) {
	var results []string
	for _, path := range list {
		byt, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		data := VottFormat{}
		if err := json.Unmarshal(byt, &data); err != nil {
			return nil, err
		}

		w := data.Asset.Size.Width
		h := data.Asset.Size.Height
		for _, reg := range data.Regions {
			tagName, points := reg.FullVertices(w, h)
			if tagName == "" {
				continue
			}

			// e.g. label,0.1,0.1,0.1,0.2,0.2,0.2,0.2,0.1
			labelData := strings.Join(append([]string{tagName}, points...), ",")
			results = append(results, r.Formatter.format(data.Asset.Name, labelData))
		}
	}
	return results, nil
}

// type mappings for VoTT JSON

type VottFormat struct {
	Asset   vottAsset    `json:"asset"`
	Regions []vottRegion `json:"regions"`
}

type vottAsset struct {
	Format string   `json:"format"`
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Path   string   `json:"path"`
	Size   vottSize `json:"size"`
	State  int64    `json:"state"`
	Type   int64    `json:"type"`
}

type vottSize struct {
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
}

type vottRegion struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Tags        []string        `json:"tags"`
	BoundingBox vottBoundingBox `json:"boundingBox"`
	Points      []vottPoint     `json:"points"`
}

func (v vottRegion) FullVertices(width, height int64) (tagName string, vertices []string) {
	if len(v.Tags) == 0 {
		return "", nil
	}
	tagName = v.Tags[0]

	const verticesSize = 4 * 2
	results := make([]string, 0, verticesSize)
	for _, p := range v.Points {
		results = append(results, strconv.FormatFloat(p.X/float64(width), 'f', -1, 64))
		results = append(results, strconv.FormatFloat(p.Y/float64(height), 'f', -1, 64))
	}
	return tagName, results
}

type vottBoundingBox struct {
	Height float64 `json:"height"`
	Width  float64 `json:"width"`
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
}

type vottPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
