package main

import (
	"bytes"
	"strconv"
)

// ParseFileLine
// ../../abc.go:688: cannot inline ...
// /go/src/abc.go:688: cannot inline ...
// /go/src/abc.go:688:123: cannot inline ...
// ..\..\abc.go:688: cannot inline ...
// C:\Go\src\example\abc.go:688: cannot inline ...
// C:\Go\src\example\abc.go:688:123: cannot inline ...
func ParseFileLine(line []byte) (path []byte, lineno, column int, msg []byte, ok bool) {
	lineno = -1
	column = -1

	// skip first 2 characters, to handle windows letter "C:"
	first := IndexByteAt(line, 2, ':')
	if first < 0 {
		return
	}
	second := IndexByteAt(line, first+1, ':')
	if second < 0 {
		return
	}
	third := IndexByteAt(line, second+1, ' ')
	if third < 0 {
		return
	}

	path = line[:first]
	lineno, ok = ParseInt(line[first+1 : second])
	if !ok {
		return
	}
	if second+1 < third-1 {
		if col, colok := ParseInt(line[second+1 : third-1]); colok {
			column = col
		}
	}
	msg = line[third+1:]
	return
}

func ParseInt(data []byte) (int, bool) {
	x, err := strconv.Atoi(string(data))
	if err != nil {
		return -1, false
	}
	return x, true
}

func IndexByteAt(data []byte, at int, b byte) int {
	s := bytes.IndexByte(data[at:], b)
	if s < 0 {
		return s
	}
	return s + at
}
