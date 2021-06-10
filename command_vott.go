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
	InputDir    string `cli:"*i,input" usage:"VoTT json results dir path --input='/path/to/vott_json_dir'"`
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
	InputDir    string
	Output      string
	PathPrefix  string
	IsRecursive bool

	Formatter formatter
}

func newVottRunner(p vottT) VottRunner {
	return VottRunner{
		InputDir:    p.InputDir,
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
	baseDir = fmt.Sprintf("%s/", filepath.Clean(r.InputDir))
	jsonFiles, err := r.FindJSONFilesFromDir(baseDir)
	if err != nil {
		return err
	}

	// read VoTT JSON and convert to AutoML format.
	results, err := r.ReadDataFromJSONFiles(jsonFiles)
	if err != nil {
		return err
	}

	// save to CSV file
	return f.WriteAll(results)
}

func (r VottRunner) FindJSONFilesFromDir(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
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
		if !data.HasValidBoundingBox() {
			fmt.Printf("[WARN] invalid bounding box: [%s]\n", path)
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

func (v VottFormat) HasValidBoundingBox() bool {
	for _, r := range v.Regions {
		if len(r.Points) < 2 {
			return false
		}

		minX, minY, maxX, maxY := -1.0, -1.0, -1.0, -1.0
		for _, p := range r.Points {
			if minX < 0 || p.X < minX {
				minX = p.X
			}
			if minY < 0 || p.Y < minY {
				minY = p.Y
			}
			if maxX < 0 || p.X > maxX {
				maxX = p.X
			}
			if maxY < 0 || p.Y > maxY {
				maxY = p.Y
			}
		}
		switch {
		case minX == maxX, minY == maxY:
			return false
		}
	}
	return true
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

	minX, minY, maxX, maxY := -1.0, -1.0, -1.0, -1.0
	for _, p := range v.Points {
		if minX < 0 || p.X < minX {
			minX = p.X
		}
		if minY < 0 || p.Y < minY {
			minY = p.Y
		}
		if maxX < 0 || p.X > maxX {
			maxX = p.X
		}
		if maxY < 0 || p.Y > maxY {
			maxY = p.Y
		}
	}

	// output: [(x1,y1), (x2,y1), (x2,y2), (x1,y2)]
	const verticesSize = 4 * 2
	results := make([]string, verticesSize)
	results[0] = strconv.FormatFloat(minX/float64(width), 'f', -1, 64)
	results[1] = strconv.FormatFloat(minY/float64(height), 'f', -1, 64)
	results[2] = strconv.FormatFloat(maxX/float64(width), 'f', -1, 64)
	results[3] = strconv.FormatFloat(minY/float64(height), 'f', -1, 64)
	results[4] = strconv.FormatFloat(maxX/float64(width), 'f', -1, 64)
	results[5] = strconv.FormatFloat(maxY/float64(height), 'f', -1, 64)
	results[6] = strconv.FormatFloat(minX/float64(width), 'f', -1, 64)
	results[7] = strconv.FormatFloat(maxY/float64(height), 'f', -1, 64)
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
