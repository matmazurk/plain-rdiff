package main

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateAndSendChecksums(t *testing.T) {
	t.Run("should calculate proper amount of checksums for various inputs", func(t *testing.T) {
		tcs := []struct {
			input        string
			windowLength int
		}{
			{
				input:        "abcdefghijk",
				windowLength: 3,
			},
			{
				input:        "abcdefghi",
				windowLength: 3,
			},
		}
		for _, tc := range tcs {
			br := NewBufferedReader(tc.windowLength, strings.NewReader(tc.input))
			checksumsChan := make(chan []byte, 3)
			go func() {
				err := CalculateAndSendChecksums(
					br,
					checksumsChan,
					func(b []byte) (uint32, *uint32, *uint32) {
						s := uint32(0)
						return uint32(0), &s, &s
					},
				)
				assert.NoError(t, err)
			}()

			var checksums [][]byte
			for c := range checksumsChan {
				checksums = append(checksums, c)
			}
			div := float64(len(tc.input)) / float64(tc.windowLength)
			assert.Len(t, checksums, int(math.Ceil(div)))
			for _, c := range checksums {
				assert.Len(t, c, 20)
			}
		}
	})
}
