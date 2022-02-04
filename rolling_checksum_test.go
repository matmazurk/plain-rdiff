package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateChecksumWithoutPreviousCompounds(t *testing.T) {
	t.Run("should properly calculate checksums", func(t *testing.T) {
		testData := map[string]uint32{
			"abc": uint32(586<<16 | 294),
			"def": uint32(604<<16 | 303),
			"ghi": uint32(622<<16 | 312),
			"jk":  uint32(319<<16 | 213),
		}
		for str, checksum := range testData {
			actualChecksum, _, _ := CalculateChecksumWithoutPreviousCompounds([]byte(str))
			assert.Equal(t, checksum, actualChecksum)
		}
	})
}

func TestCalculateChecksumUsingPreviousCompounds(t *testing.T) {
	t.Run("should properly calculate checksums when input converges to zero", func(t *testing.T) {
		input := "some random input"
		windowLength := len(input)
		_, a, b := CalculateChecksumWithoutPreviousCompounds([]byte(input))
		for i := 1; i < windowLength; i++ {
			var checksum uint32
			currInput := []byte(input[i:])
			checksum, a, b = CalculateChecksumUsingPreviousCompounds(currInput, input[i-1], windowLength, *a, *b)
			refChecksum, _, _ := CalculateChecksumWithoutPreviousCompounds(currInput)
			assert.Equal(t, refChecksum, checksum)
		}
	})
}
