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
	r            ReaderSeeker
	windowLength uint32
	buffer       []byte
	length       uint32
	offset       int64
	eof          bool
}

func NewBufferedReader(windowLength uint32, readerSeeker ReaderSeeker) bufferedReader {
	br := bufferedReader{
		r:            readerSeeker,
		windowLength: windowLength,
		buffer:       make([]byte, windowLength),
	}
	return br
}

func (br *bufferedReader) ReadWindow() (bool, error) {
	readBytes, err := br.r.ReadAt(br.buffer, br.offset+int64(br.length))
	if err != nil && !errors.Is(err, io.EOF) {
		return readBytes > 0, err
	}
	if errors.Is(err, io.EOF) {
		br.eof = true
	}
	br.offset += int64(br.length)
	br.length = uint32(readBytes)
	return readBytes > 0, nil
}

func (br *bufferedReader) Offset() int64 {
	return br.offset
}

func (br *bufferedReader) Len() uint32 {
	return br.length
}

func (br *bufferedReader) isEOF() bool {
	return br.eof
}

func (br *bufferedReader) GetHash(calculations func([]byte) []byte) []byte {
	return calculations(br.buffer)
}

func (br *bufferedReader) Get(index int) byte {
	return br.buffer[index]
}

func (br *bufferedReader) Buf() []byte {
	return br.buffer[:br.length]
}
