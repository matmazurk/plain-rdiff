//go:build e2e

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRdiffFile(t *testing.T) {
	tcs := []struct {
		filesLen        int
		windowSize      int
		noOfDifferences int
	}{
		{
			filesLen:        60_000_000,
			windowSize:      8000,
			noOfDifferences: 20,
		},
		{
			filesLen:        math.MaxUint32 + 1000,
			windowSize:      1_000_000,
			noOfDifferences: 10,
		},
		{
			filesLen:        math.MaxUint32 + 1000,
			windowSize:      10_000_000,
			noOfDifferences: 0,
		},
		{
			filesLen:        100_000_000,
			windowSize:      20_000,
			noOfDifferences: 10000,
		},
	}
	for _, tc := range tcs {
		t.Run(
			fmt.Sprintf("filesLen:%d windowSize:%d noOfDifferences:%d", tc.filesLen, tc.windowSize, tc.noOfDifferences),
			func(t *testing.T) {
				t.Log("generating old and ref files")
				refFileName := "__test_reference_file"
				refFileChan := make(chan []byte)
				wg := sync.WaitGroup{}
				oldFileName := "__test_old_file"
				oldFileChan := make(chan []byte)
				wg.Add(1)
				go func() {
					err := CreateAndFillFile(refFileName, refFileChan)
					if err != nil {
						panic(err)
					}
					wg.Done()
				}()
				wg.Add(1)
				go func() {
					err := CreateAndFillFile(oldFileName, oldFileChan)
					if err != nil {
						panic(err)
					}
					wg.Done()
				}()

				bytes := make([]byte, tc.filesLen)
				rand.Read(bytes)
				refFileChan <- bytes
				oldFileContent := make([]byte, tc.filesLen)
				copy(oldFileContent, bytes)
				for i := 0; i < tc.noOfDifferences; i++ {
					index := rand.Intn(int(tc.filesLen))
					oldFileContent[index] = byte(rand.Int())
				}
				oldFileChan <- oldFileContent
				close(refFileChan)
				close(oldFileChan)
				wg.Wait()
				defer func() {
					err := os.Remove(refFileName)
					if err != nil {
						panic(err)
					}
					err = os.Remove(oldFileName)
					if err != nil {
						panic(err)
					}
				}()

				// signature
				signatureFileName := "__test_signature_file"
				t.Log("calculating signature")
				signatureFlow(oldFileName, signatureFileName, tc.windowSize)
				defer func() {
					err := os.Remove(signatureFileName)
					if err != nil {
						panic(err)
					}
				}()

				// delta
				deltaFileName := "__test_delta_file"
				t.Log("calculating delta")
				deltaFlow(signatureFileName, refFileName, deltaFileName, tc.windowSize)
				defer func() {
					err := os.Remove(deltaFileName)
					if err != nil {
						panic(err)
					}
				}()

				// patch
				newFileName := "__test_new_file"
				t.Log("applying patch")
				patchFlow(oldFileName, deltaFileName, newFileName)
				defer func() {
					err := os.Remove(newFileName)
					if err != nil {
						panic(err)
					}
				}()

				// compare
				t.Log("comparing file hashes")
				refFile, err := os.Open(refFileName)
				if err != nil {
					t.Fatal(err)
				}
				defer refFile.Close()

				refFileHash := sha256.New()
				if _, err := io.Copy(refFileHash, refFile); err != nil {
					t.Fatal(err)
				}

				newFile, err := os.Open(newFileName)
				if err != nil {
					t.Fatal(err)
				}
				defer newFile.Close()

				newFileHash := sha256.New()
				if _, err := io.Copy(newFileHash, newFile); err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, refFileHash, newFileHash)
			})
	}
}
