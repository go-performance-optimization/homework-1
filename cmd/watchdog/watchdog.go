package main

import (
	"io"
)

type (
	reader interface {
		io.Reader
		ReadSlice(delim byte) (line []byte, err error)
	}

	writer interface {
		io.Writer
		WriteByte(c byte) error
		Flush() error
	}
)

func main() {
	panic("not implemented")
}
