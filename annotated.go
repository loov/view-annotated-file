package main

import (
	"errors"
	"io/ioutil"
	"strings"
)

type AnnotatedFile struct {
	Path    string `json:"path"`
	AbsPath string `json:"path"`
	Lines   []Line `json:"lines"`
}

type Line struct {
	Source string     `json:"source"`
	Notes  []LineNote `json:"notes"`
}

type LineNote struct {
	Column  int    `json:"column"`
	Message string `json:"message"`
}

func (index *Index) LoadAnnotatedFile(path string) (*AnnotatedFile, error) {
	info, ok := index.Files[path]
	if !ok {
		return nil, errors.New("not found")
	}

	data, err := ioutil.ReadFile(info.AbsPath)
	if err != nil {
		return nil, err
	}

	file := &AnnotatedFile{}
	file.Path = info.Path
	file.AbsPath = info.AbsPath

	noteidx := 0
	sourceLines := strings.Split(string(data), "\n")
	for i, sourceLine := range sourceLines {
		line := Line{}
		line.Source = sourceLine
		line.Notes = []LineNote{}

		for noteidx < len(info.Notes) && i > info.Notes[noteidx].Line {
			noteidx++
		}
		for noteidx < len(info.Notes) && i == info.Notes[noteidx].Line {
			x := info.Notes[noteidx]
			note := LineNote{
				Column:  x.Column,
				Message: string(x.Message),
			}
			line.Notes = append(line.Notes, note)
			noteidx++
		}

		file.Lines = append(file.Lines, line)
	}

	return file, nil
}
