package main

import (
	"fmt"
	"strings"
)

func createListFormat(name string) (formatter, error) {
	name = strings.ToLower(name)
	switch name {
	case "sagemaker":
		return &sagemakerFormatter{}, nil
	case "csv":
		return &csvFormatter{}, nil
	default:
		return nil, fmt.Errorf("Unknown Format: [%s]", name)
	}
}

type formatter interface {
	format(path, label string) string
}

type csvFormatter struct{}

func (csvFormatter) format(path, label string) string {
	return fmt.Sprintf("%s,%s", path, label)
}

type sagemakerFormatter struct{}

func (sagemakerFormatter) format(path, label string) string {
	return fmt.Sprintf(`{"source-ref": "%s"}`, path)
}

type automlObjectDetectionFormatter struct {
	pathPrefix string
	defaultSet string
}

func (f automlObjectDetectionFormatter) format(path, label string) string {
	set := f.defaultSet
	if set == "" {
		set = "UNASSIGNED"
	}
	return fmt.Sprintf("%s,%s,%s", set, f.pathPrefix+path, label)
}
