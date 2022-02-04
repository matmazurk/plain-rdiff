package main

import (
	"crypto/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadWindow(t *testing.T) {
	t.Run("should successfully read whole input", func(t *testing.T) {
		windowSize := 10
		shift := 5
		testString := string("15 chars string")

		br := NewBufferedReader(windowSize, strings.NewReader(testString))
		anyReads, err := br.ReadWindow()
		assert.NoError(t, err)
		assert.True(t, anyReads > 0)

		assert.Equal(t, windowSize, br.Len())
		expectedBytes := []byte(testString[:windowSize])
		assert.ElementsMatch(t, expectedBytes, br.Buf())

		anyReads, err = br.ReadWindow()
		assert.NoError(t, err)
		assert.True(t, anyReads > 0)
		assert.True(t, br.isEOF())
		assert.Equal(t, shift, br.Len())
		expectedBytes = []byte(testString[windowSize:])
		assert.ElementsMatch(t, expectedBytes, br.Buf())

		anyReads, err = br.ReadWindow()
		assert.NoError(t, err)
		assert.False(t, anyReads > 0)
		assert.True(t, br.isEOF())
	})
}

func TestBuf(t *testing.T) {
	t.Run("should properly read input that is not multiplier of its buf len", func(t *testing.T) {
		bytes := make([]byte, 10)
		rand.Read(bytes)

		br := NewBufferedReader(15, strings.NewReader(string(bytes)))
		anyReads, err := br.ReadWindow()
		assert.NoError(t, err)
		assert.True(t, anyReads > 0)
		assert.Equal(t, bytes, br.Buf())
	})

	t.Run("should properly read big input", func(t *testing.T) {
		bytes := make([]byte, 1005)
		rand.Read(bytes)

		br := NewBufferedReader(10, strings.NewReader(string(bytes)))

		var actualBytes []byte
		for i := 0; i < len(bytes); i += 10 {
			_, err := br.ReadWindow()
			assert.NoError(t, err)
			actualBytes = append(actualBytes, br.Buf()...)
		}

		assert.ElementsMatch(t, bytes, actualBytes)
	})
}

func TestPopAndShift(t *testing.T) {
	t.Run(
		`should properly pop and shift through whole input, setting EOF and returning read bytes equal to 0 at the end`,
		func(t *testing.T) {
			bytes := make([]byte, 19)
			rand.Read(bytes)
			windowSize := 10

			br := NewBufferedReader(windowSize, strings.NewReader(string(bytes)))
			read, err := br.ReadWindow()
			assert.NoError(t, err)
			assert.Equal(t, windowSize, read)
			bytesFromPops := []byte{}
			for i := 1; i <= len(bytes)-windowSize; i++ {
				pop, err := br.PopAndShift()
				assert.NoError(t, err)
				assert.Equal(t, int64(i), br.Offset())
				assert.Equal(t, windowSize, br.Len())
				bytesFromPops = append(bytesFromPops, pop)
			}
			_, err = br.PopAndShift()
			assert.NoError(t, err)
			assert.True(t, br.isEOF())
			assert.Equal(t, int64(len(bytes)-windowSize+1), br.Offset())
			assert.Equal(t, windowSize-1, br.Len())
			assert.Equal(t, bytes[:len(bytes)-windowSize], bytesFromPops)
		},
	)

	t.Run("should return error when buffer is empty", func(t *testing.T) {
		br := NewBufferedReader(5, strings.NewReader(string("")))
		_, err := br.PopAndShift()
		assert.ErrorIs(t, err, ErrEmptyBuffer)
	})
}
