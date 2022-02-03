package main

import (
	"encoding/binary"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateAndSendChecksums(t *testing.T) {
	t.Run("should properly calculate checksums for input not being multiple of window size", func(t *testing.T) {
		base := "abcdefghijk"
		windowLength := 3
		firstExpectedChecksum := uint32(586<<16 | 294)
		secondExpectedChecksum := uint32(604<<16 | 303)
		thirdExpectedChecksum := uint32(622<<16 | 312)
		fourthExpectedChecksum := uint32(319<<16 | 213)

		br := NewBufferedReader(uint32(windowLength), strings.NewReader(base))
		checksumsChan := make(chan []byte, 3)
		go func() {
			err := CalculateAndSendChecksums(br, checksumsChan)
			assert.NoError(t, err)
		}()

		var checksums [][]byte
		for checksum := range checksumsChan {
			checksums = append(checksums, checksum)
		}

		firstChecksum := binary.BigEndian.Uint32(checksums[0][:4])
		secondChecksum := binary.BigEndian.Uint32(checksums[1][:4])
		thirdChecksum := binary.BigEndian.Uint32(checksums[2][:4])
		fourthChecksum := binary.BigEndian.Uint32(checksums[3][:4])
		assert.Equal(t, firstExpectedChecksum, firstChecksum)
		assert.Equal(t, secondExpectedChecksum, secondChecksum)
		assert.Equal(t, thirdExpectedChecksum, thirdChecksum)
		assert.Equal(t, fourthExpectedChecksum, fourthChecksum)
	})

	t.Run("should properly calculate checksums for input being multiple of window size", func(t *testing.T) {
		base := "abcdefghi"
		windowLength := 3
		firstExpectedChecksum := uint32(586<<16 | 294)
		secondExpectedChecksum := uint32(604<<16 | 303)
		thirdExpectedChecksum := uint32(622<<16 | 312)

		br := NewBufferedReader(uint32(windowLength), strings.NewReader(base))
		checksumsChan := make(chan []byte, 3)
		go func() {
			err := CalculateAndSendChecksums(br, checksumsChan)
			assert.NoError(t, err)
		}()

		var checksums [][]byte
		for checksum := range checksumsChan {
			checksums = append(checksums, checksum)
		}

		firstChecksum := binary.BigEndian.Uint32(checksums[0][:4])
		secondChecksum := binary.BigEndian.Uint32(checksums[1][:4])
		thirdChecksum := binary.BigEndian.Uint32(checksums[2][:4])
		assert.Equal(t, firstExpectedChecksum, firstChecksum)
		assert.Equal(t, secondExpectedChecksum, secondChecksum)
		assert.Equal(t, thirdExpectedChecksum, thirdChecksum)
	})
}
