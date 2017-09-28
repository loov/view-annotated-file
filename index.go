package main

import (
	"bytes"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type Index struct {
	Files map[string]*File
}

type File struct {
	Path    string
	AbsPath string
	Stats   Stats
	Notes   []Note
}

type Note struct {
	Line    int // 0 is the first line
	Column  int // 0 is the first column
	Message []byte
}

func NewIndex() *Index {
	index := &Index{}
	index.Files = make(map[string]*File)
	return index
}

func NewFile(dir string, path string) *File {
	file := &File{}
	file.Path = path
	if filepath.IsAbs(path) {
		file.AbsPath = path
	} else {
		file.AbsPath = filepath.Join(dir, path)
	}
	return file
}

func (index *Index) Parse(dir string, data []byte) {
	lineStart := 0
	lineEnd := 0
	for lineStart < len(data) {
		lineEnd = IndexByteAt(data, lineStart, '\n')
		if lineEnd < 0 {
			lineEnd = len(data)
		}
		index.Add(dir, data[lineStart:lineEnd])
		lineStart = lineEnd + 1
	}

	index.Sort()
}

func (index *Index) Sort() {
	for _, file := range index.Files {
		sort.Slice(file.Notes, func(i, k int) bool {
			if file.Notes[i].Line == file.Notes[k].Line {
				return file.Notes[i].Column < file.Notes[k].Column
			}
			return file.Notes[i].Line < file.Notes[k].Line
		})
	}
}

func (index *Index) Add(dir string, line []byte) {
	if len(line) <= 2 {
		return
	}

	for _, ignore := range ignoredLines {
		if bytes.HasPrefix(line, []byte(ignore)) {
			return
		}
	}

	for _, ignore := range ignoredContent {
		if bytes.Contains(line, []byte(ignore)) {
			return
		}
	}

	pathbytes, lineno, col, msg, ok := ParseFileLine(line)
	if !ok {
		return
	}

	path := string(pathbytes)
	if runtime.GOOS == "windows" {
		path = strings.ToLower(path)
	}

	file, ok := index.Files[path]
	if !ok {
		file = NewFile(dir, path)
		index.Files[path] = file
	}

	file.Stats.Add(msg)
	file.Notes = append(file.Notes, Note{
		Line:    lineno - 1,
		Column:  col - 1,
		Message: msg,
	})
}
