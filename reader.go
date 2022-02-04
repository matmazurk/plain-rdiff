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
	windowLength int
	buffer       []byte
	length       int
	offset       int64
	eof          bool
}

var ErrEmptyBuffer = errors.New("empty buffer- cannot pop")

func NewBufferedReader(windowLength int, readerSeeker ReaderSeeker) bufferedReader {
	br := bufferedReader{
		r:            readerSeeker,
		windowLength: windowLength,
		buffer:       make([]byte, windowLength),
	}
	return br
}

func (br *bufferedReader) ReadWindow() (int, error) {
	readBytes, err := br.r.ReadAt(br.buffer, br.offset+int64(br.length))
	if err != nil && !errors.Is(err, io.EOF) {
		return readBytes, err
	}
	if errors.Is(err, io.EOF) {
		br.eof = true
	}
	br.offset += int64(br.length)
	br.length = readBytes
	return readBytes, nil
}

func (br *bufferedReader) PopAndShift() (byte, error) {
	if br.length == 0 {
		return byte(0), ErrEmptyBuffer
	}
	buf := make([]byte, 1)
	readBytes, err := br.r.ReadAt(buf, br.offset+int64(br.length))
	if err != nil && !errors.Is(err, io.EOF) {
		return byte(0), err
	}
	if errors.Is(err, io.EOF) {
		br.eof = true
	}
	newByte := buf[0]
	pop := br.buffer[0]
	for i := 0; i < len(br.buffer)-1; i++ {
		br.buffer[i] = br.buffer[i+1]
	}
	br.buffer[len(br.buffer)-1] = newByte
	br.offset++
	if readBytes == 0 {
		br.length--
	}

	return pop, nil
}

func (br *bufferedReader) Offset() int64 {
	return br.offset
}

func (br *bufferedReader) Len() int {
	return br.length
}

func (br *bufferedReader) isEOF() bool {
	return br.eof
}

func (br *bufferedReader) GetHash(calculations func([]byte) []byte) []byte {
	return calculations(br.buffer[:br.length])
}

func (br *bufferedReader) Get(index int) byte {
	return br.buffer[index]
}

func (br *bufferedReader) Buf() []byte {
	return br.buffer[:br.length]
}

func (br *bufferedReader) WindowLen() int {
	return br.windowLength
}
