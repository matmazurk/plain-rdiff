package main

import (
	"errors"
	"io"
)

type ReaderSeeker interface {
	ReadAt(p []byte, off int64) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
}

type bufferedReader struct {
	ReaderSeeker
	windowLength uint32
	buffer       []byte
	length       uint32
	offset       int64
}

func NewBufferedReader(windowLength uint32, readerSeeker ReaderSeeker) bufferedReader {
	br := bufferedReader{
		windowLength: windowLength,
		buffer:       make([]byte, windowLength),
	}
	return br
}

func (br *bufferedReader) ReadWindow() error {
	readBytes, err := br.ReadAt(br.buffer, br.offset)
	if err != nil && errors.Is(err, io.EOF) {
		return err
	}
	br.length = uint32(readBytes)
	br.offset, err = br.Seek(0, 1)
	return err
}

func (br *bufferedReader) Offset() int64 {
	return br.offset
}

func (br *bufferedReader) Len() uint32 {
	return br.length
}
