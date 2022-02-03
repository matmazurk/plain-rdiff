package main

import (
	"errors"
	"io"

	"golang.org/x/crypto/md4"
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
	eof          bool
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
	if errors.Is(err, io.EOF) {
		br.eof = true
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

func (br *bufferedReader) SumBufferBytes() uint32 {
	var sum uint32
	for i := uint32(0); i < br.length; i++ {
		sum += uint32(br.buffer[i])
	}
	return sum
}

func (br *bufferedReader) isEOF() bool {
	return br.eof
}

func (br *bufferedReader) MD4() []byte {
	h := md4.New()
	h.Write(br.buffer[:len(br.buffer)-int(br.length)])
	return h.Sum(nil)
}

func (br *bufferedReader) Get(index int) byte {
	return br.buffer[index]
}
