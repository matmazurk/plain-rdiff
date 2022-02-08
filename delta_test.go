package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const _WINDOW_SIZE = 10

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
			br := NewBufferedReader(_WINDOW_SIZE, strings.NewReader(tc.referenceFileContent))

			deltaChunkChan := make(chan DeltaChunk)
			go func() {
				err := CalculateAndSendDeltaChunks(
					br,
					deltaChunkChan,
					nil,
					nil,
					mockFindMatchingOffset(tc.oldFileContent),
					mockCalculateChecksum,
				)
				assert.NoError(t, err)
			}()

			deltaChunks := []DeltaChunk{}
			for chunk := range deltaChunkChan {
				deltaChunks = append(deltaChunks, chunk)
			}

			referenceFileFromDelta := getReferenceFileFromDelta(tc.oldFileContent, deltaChunks)
			assert.Equal(t, tc.referenceFileContent, referenceFileFromDelta)
		})
	}
}

func mockFindMatchingOffset(
	refFile string,
) func([]byte, [][]byte, uint32, map[uint32]int) (bool, int) {
	return func(h []byte, _ [][]byte, _ uint32, _ map[uint32]int) (bool, int) {
		if len(refFile) == 0 || len(h) == 0 {
			return false, 0
		}
		if r := strings.Index(refFile, string(h)); r != -1 {
			if r%_WINDOW_SIZE == 0 {
				return true, r / _WINDOW_SIZE
			}
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

func getReferenceFileFromDelta(oldFileContent string, deltaChunks []DeltaChunk) string {
	var originalFileFromDelta strings.Builder
	for _, chunk := range deltaChunks {
		if !chunk.rawData {
			originalFileFromDelta.WriteString(oldFileContent[*chunk.r.from:*chunk.r.to])
			continue
		}
		originalFileFromDelta.WriteString(string(chunk.d))
	}
	return originalFileFromDelta.String()
}
