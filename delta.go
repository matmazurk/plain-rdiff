package main

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type Range struct {
	from *uint64
	to   *uint64
}

func (r *Range) empty() bool {
	return r.from == nil && r.to == nil
}

func (r *Range) clear() {
	r.from = nil
	r.to = nil
}

func (r *Range) set(from int, to int) {
	f := uint64(from)
	t := uint64(to)
	r.from = &f
	r.to = &t
}

func (r *Range) shiftToBy(shiftLen int) {
	*r.to += uint64(shiftLen)
}

type DeltaChunk struct {
	r       Range
	d       []byte
	rawData bool
}

func NewDeltaChunkWithRange(r Range) DeltaChunk {
	return DeltaChunk{
		r: r,
	}
}

func NewDeltaChunkWithRawData(data []byte) DeltaChunk {
	return DeltaChunk{
		d:       data,
		rawData: true,
	}
}

func (c DeltaChunk) ToBytes() []byte {
	if !c.rawData {
		bytes := make([]byte, 1+8+8)
		bytes[0] = byte(1)

		binary.BigEndian.PutUint64(bytes[1:9], *c.r.from)
		binary.BigEndian.PutUint64(bytes[9:17], *c.r.to)
		return bytes
	}
	unmatchedDataLen := len(c.d)
	bytes := make([]byte, 1+8+unmatchedDataLen)
	binary.BigEndian.PutUint64(bytes[1:9], uint64(unmatchedDataLen))
	for i, ii := 9, 0; i < len(bytes); i++ {
		bytes[i] = c.d[ii]
		ii++
	}
	return bytes
}

func CalculateAndSendDeltaChunks(
	referenceFileReader bufferedReader,
	deltaChunkChan chan<- DeltaChunk,
	rollingChecksumsToIndexes map[uint32]int,
	hashes [][]byte,
	findMatchingOffset func([]byte, [][]byte, uint32, map[uint32]int) (bool, int),
	checksumCalculation func([]byte, *byte, int, *uint32, *uint32) (uint32, *uint32, *uint32),
) error {
	defer close(deltaChunkChan)

	var unmatchedBytes []byte
	var checksum uint32
	var readBytes int
	var pop *byte
	var a *uint32
	var b *uint32
	var err error
	r := Range{}

	for {
		if a == nil && b == nil {
			readBytes, err = referenceFileReader.ReadWindow()
			if err != nil {
				return err
			}
			if readBytes == 0 {
				if !r.empty() {
					deltaChunkChan <- NewDeltaChunkWithRange(r)
				}
				return nil
			}
		}
		checksum, a, b = checksumCalculation(
			referenceFileReader.Buf(),
			pop,
			referenceFileReader.WindowLen(),
			a,
			b,
		)
		matching, offset := findMatchingOffset(
			referenceFileReader.Buf(),
			hashes,
			checksum,
			rollingChecksumsToIndexes,
		)
		if matching {
			a, b = nil, nil
			if len(unmatchedBytes) > 0 {
				deltaChunkChan <- NewDeltaChunkWithRawData(unmatchedBytes)
				unmatchedBytes = []byte{}
			}
			if r.empty() {
				r.set(offset*referenceFileReader.WindowLen(), offset*referenceFileReader.WindowLen()+readBytes)
				continue
			}
			if *r.to == uint64(offset*referenceFileReader.WindowLen()) {
				r.shiftToBy(readBytes)
				continue
			}

			deltaChunkChan <- NewDeltaChunkWithRange(r)
			r.set(offset*referenceFileReader.WindowLen(), offset*referenceFileReader.WindowLen()+readBytes)
			continue
		}

		if !r.empty() {
			deltaChunkChan <- NewDeltaChunkWithRange(r)
			r.clear()
		}

		p, err := referenceFileReader.PopAndShift()
		pop = &p
		if err != nil {
			if errors.Is(err, ErrEmptyBuffer) {
				if len(unmatchedBytes) > 0 {
					deltaChunkChan <- NewDeltaChunkWithRawData(unmatchedBytes)
				}
				return nil
			}
			return err
		}
		unmatchedBytes = append(unmatchedBytes, p)
	}
}

func findMatchingOffset(
	block []byte,
	hashes [][]byte,
	checksum uint32,
	checksums map[uint32]int,
) (bool, int) {
	index, ok := checksums[checksum]
	if ok && bytes.Equal(calculateMD4(block), hashes[index]) {
		return true, index
	}
	return false, 0
}
