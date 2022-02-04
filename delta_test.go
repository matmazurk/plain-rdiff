package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const _SEPARATOR = "~"
const _FROM_TO_SEPARATOR = ":"

func TestCalculateAndSendDeltaChunks(t *testing.T) {
	tcs := []struct {
		name                 string
		referenceFileContent string
		oldFileContent       string
	}{
		{
			name:                 "should successfully recreate identical file",
			referenceFileContent: "Imagine you have two files, A and B, and you wish to update B to be the same as A.",
			oldFileContent:       "Imagine you have two files, A and B, and you wish to update B to be the same as A.",
		},
		{
			name:                 "should successfully recreate different file",
			referenceFileContent: "Imagine you wish to uphave two files, A and B, and you wish to update B to be the same as A.",
			oldFileContent:       "Imagine you have two files, A and B, and you wish to update B to be the same as A.",
		},
		{
			name:                 "should successfully recreate completely different file",
			referenceFileContent: "Imagfwne you have twoyou have tw||oyou have tw* #@d B, ad dw wish to !@#be the sa$me as A",
			oldFileContent:       "Iamaginfe u have two sasdf, Asda andd yosdfu wio update B to be to be to bedf the sam@#s A. ",
		},
		{
			name:                 "should successfully recreate file with length equal to window length",
			referenceFileContent: "Imagine yo",
			oldFileContent:       "Imagine yo",
		},
		{
			name:                 "should successfully recreate empty file",
			referenceFileContent: "",
			oldFileContent:       "Imagine yo",
		},
		{
			name:                 "should successfully recreate file basing on empty file",
			referenceFileContent: "Imagine you have two files, A and B, and you wish to update B to be the same as A.",
			oldFileContent:       "",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			br := NewBufferedReader(10, strings.NewReader(tc.referenceFileContent))

			deltaChunkChan := make(chan []byte)
			go func() {
				err := CalculateAndSendDeltaChunks(
					br,
					deltaChunkChan,
					nil,
					nil,
					mockGetDeltaChunk,
					mockFindMatchingOffset(tc.oldFileContent),
					mockCalculateChecksum,
					mockHashCalculation,
				)
				assert.NoError(t, err)
			}()

			deltaChunks := []byte{}
			for chunk := range deltaChunkChan {
				deltaChunks = append(deltaChunks, chunk...)
			}

			referenceFileFromDelta := getReferenceFileFromDelta(tc.oldFileContent, string(deltaChunks))
			assert.Equal(t, tc.referenceFileContent, referenceFileFromDelta)
		})
	}
}

func mockGetDeltaChunk(r *Range, b *[]byte) []byte {
	if r != nil {
		return []byte(string(fmt.Sprintf("%d%s%d%s", *r.from, _FROM_TO_SEPARATOR, *r.to, _SEPARATOR)))
	}
	return append(*b, []byte(_SEPARATOR)...)
}

func mockFindMatchingOffset(
	refFile string,
) func([]byte, [][]byte, uint32, map[uint32]int) (bool, int) {
	return func(h []byte, _ [][]byte, _ uint32, _ map[uint32]int) (bool, int) {
		if r := strings.Index(refFile, string(h)); r != -1 {
			return true, r
		}
		return false, 0
	}

}

func mockCalculateChecksum(data []byte, previous *byte, length int, a, b *uint32) (uint32, *uint32, *uint32) {
	s := uint32(0)
	return uint32(0), &s, &s
}

func mockHashCalculation(data []byte) []byte {
	return data
}

func getReferenceFileFromDelta(oldFileContent, delta string) string {
	var originalFileFromDelta strings.Builder
	splits := strings.Split(delta, _SEPARATOR)
	for _, split := range splits {
		if strings.Contains(split, _FROM_TO_SEPARATOR) {
			vals := strings.Split(split, _FROM_TO_SEPARATOR)
			from, _ := strconv.Atoi(vals[0])
			to, _ := strconv.Atoi(vals[1])
			originalFileFromDelta.WriteString(oldFileContent[from:to])
			continue
		}
		originalFileFromDelta.WriteString(split)
	}
	return originalFileFromDelta.String()
}
