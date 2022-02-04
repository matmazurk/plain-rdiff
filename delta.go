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

func CalculateAndSendDeltaChunks(
	referenceFileReader bufferedReader,
	deltaChunkChan chan<- []byte,
	rollingChecksumsToIndexes map[uint32]int,
	hashes [][]byte,
	getRawDeltaData func(*Range, *[]byte) []byte,
	findMatchingOffset func([]byte, [][]byte, uint32, map[uint32]int) (bool, int),
	checksumCalculation func([]byte, *byte, int, *uint32, *uint32) (uint32, *uint32, *uint32),
	hashCalculation func([]byte) []byte,
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
					deltaChunkChan <- getRawDeltaData(&r, nil)
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
			referenceFileReader.GetHash(hashCalculation),
			hashes,
			checksum,
			rollingChecksumsToIndexes,
		)
		if matching {
			if len(unmatchedBytes) > 0 {
				deltaChunkChan <- getRawDeltaData(nil, &unmatchedBytes)
				unmatchedBytes = []byte{}
			}
			if r.empty() {
				r.set(offset, offset+referenceFileReader.Len())
			} else {
				if *r.to == uint64(offset) {
					r.shiftToBy(readBytes)
				} else {
					deltaChunkChan <- getRawDeltaData(&r, nil)
					r.set(offset, offset+readBytes)
				}
			}
			a, b = nil, nil
			continue
		}

		if !r.empty() {
			deltaChunkChan <- getRawDeltaData(&r, nil)
			r.clear()
		}

		pop, err := referenceFileReader.PopAndShift()
		if err != nil {
			if errors.Is(err, ErrEmptyBuffer) {
				if len(unmatchedBytes) > 0 {
					deltaChunkChan <- getRawDeltaData(nil, &unmatchedBytes)
				}
				return nil
			}
			return err
		}
		unmatchedBytes = append(unmatchedBytes, pop)
	}
}

func findMatchingOffset(
	hash []byte,
	hashes [][]byte,
	checksum uint32,
	checksums map[uint32]int,
) (bool, int) {
	for c, i := range checksums {
		if c == checksum {
			if bytes.Equal(hash, hashes[i]) {
				return true, i
			}
		}
	}
	return false, 0
}

// bytes: |1,0|uint64|data v uint64|
func getRawDeltaChunk(r *Range, unmatchedData *[]byte) []byte {
	if r != nil {
		bytes := make([]byte, 1+8+8)
		bytes[0] = byte(1)
		binary.BigEndian.PutUint64(bytes[1:9], *r.from)
		binary.BigEndian.PutUint64(bytes[9:17], *r.to)
		return bytes
	}
	unmatchedDataLen := len(*unmatchedData)
	bytes := make([]byte, 1+8+unmatchedDataLen)
	binary.BigEndian.PutUint64(bytes[1:9], uint64(unmatchedDataLen))
	for i, ii := 9, 0; i < len(bytes); i++ {
		bytes[i] = (*unmatchedData)[ii]
		ii++
	}
	return bytes
}
