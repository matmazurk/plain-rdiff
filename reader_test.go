package main

import (
	"crypto/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/md4"
)

func TestReadWindow(t *testing.T) {
	t.Run("should successfully read whole input", func(t *testing.T) {
		windowSize := 10
		shift := 5
		testString := string("15 chars string")

		br := NewBufferedReader(uint32(windowSize), strings.NewReader(testString))
		anyReads, err := br.ReadWindow()
		assert.NoError(t, err)
		assert.True(t, anyReads)

		expectedLen := uint32(windowSize)
		assert.Equal(t, expectedLen, br.Len())
		expectedBytes := []byte(testString[:windowSize])
		assert.ElementsMatch(t, expectedBytes, br.Buf())

		anyReads, err = br.ReadWindow()
		assert.NoError(t, err)
		assert.True(t, anyReads)
		assert.True(t, br.isEOF())
		expectedLen = uint32(shift)
		assert.Equal(t, expectedLen, br.Len())
		expectedBytes = []byte(testString[windowSize:])
		assert.ElementsMatch(t, expectedBytes, br.Buf())

		anyReads, err = br.ReadWindow()
		assert.NoError(t, err)
		assert.False(t, anyReads)
		assert.True(t, br.isEOF())
	})
}

func TestSumBufferBytes(t *testing.T) {
	t.Run("should properly sum bytes counts", func(t *testing.T) {
		bytes := make([]byte, 10)
		rand.Read(bytes)
		expectedSum := uint32(0)
		for _, b := range bytes {
			expectedSum += uint32(b)
		}

		br := NewBufferedReader(15, strings.NewReader(string(bytes)))
		anyReads, err := br.ReadWindow()
		assert.NoError(t, err)
		assert.True(t, anyReads)
		assert.Equal(t, expectedSum, br.SumBufferBytes())
	})
}

func TestMD4(t *testing.T) {
	t.Run("should properly calculate MD4", func(t *testing.T) {
		bytes := make([]byte, 10)
		rand.Read(bytes)
		h := md4.New()
		h.Write(bytes)
		expectedMD4 := h.Sum(nil)

		br := NewBufferedReader(15, strings.NewReader(string(bytes)))
		anyReads, err := br.ReadWindow()
		assert.NoError(t, err)
		assert.True(t, anyReads)
		assert.Equal(t, expectedMD4, br.MD4())
	})
}

func TestBuf(t *testing.T) {
	t.Run("should properly read input that is not multiplier of its buf len", func(t *testing.T) {
		bytes := make([]byte, 10)
		rand.Read(bytes)

		br := NewBufferedReader(15, strings.NewReader(string(bytes)))
		anyReads, err := br.ReadWindow()
		assert.NoError(t, err)
		assert.True(t, anyReads)
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
