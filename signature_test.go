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

		br := NewBufferedReader(uint32(windowLength), strings.NewReader(base))
		checksumsChan := make(chan []byte, 3)
		go func() {
			err := CalculateAndSendChecksums(br, checksumsChan, func(b []byte, i int64) uint32 {
				return uint32(i)
			})
			assert.NoError(t, err)
		}()

		var checksums [][]byte
		for c := range checksumsChan {
			checksums = append(checksums, c)
		}
		firstChecksum := binary.BigEndian.Uint32(checksums[0][:4])
		secondChecksum := binary.BigEndian.Uint32(checksums[1][:4])
		thirdChecksum := binary.BigEndian.Uint32(checksums[2][:4])
		fourthChecksum := binary.BigEndian.Uint32(checksums[3][:4])
		assert.Equal(t, uint32(0), firstChecksum)
		assert.Equal(t, uint32(3), secondChecksum)
		assert.Equal(t, uint32(6), thirdChecksum)
		assert.Equal(t, uint32(9), fourthChecksum)
	})

	t.Run("should properly calculate checksums for input being multiple of window size", func(t *testing.T) {
		base := "abcdefghi"
		windowLength := 3

		br := NewBufferedReader(uint32(windowLength), strings.NewReader(base))
		checksumsChan := make(chan []byte, 3)
		go func() {
			err := CalculateAndSendChecksums(br, checksumsChan, func(b []byte, i int64) uint32 {
				return uint32(i)
			})
			assert.NoError(t, err)
		}()

		var checksums [][]byte
		for checksum := range checksumsChan {
			checksums = append(checksums, checksum)
		}

		firstChecksum := binary.BigEndian.Uint32(checksums[0][:4])
		secondChecksum := binary.BigEndian.Uint32(checksums[1][:4])
		thirdChecksum := binary.BigEndian.Uint32(checksums[2][:4])
		assert.Equal(t, uint32(0), firstChecksum)
		assert.Equal(t, uint32(3), secondChecksum)
		assert.Equal(t, uint32(6), thirdChecksum)
	})
}
