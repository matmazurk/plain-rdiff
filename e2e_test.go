package main

import (
	"crypto/sha256"
	"io"
	"math/rand"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRdiffFile(t *testing.T) {
	// tcs := struct {
	// 	filesLen int
	// 	noOfDifferences int
	// }
	// s := rand.NewSource(time.Now().UnixNano())
	// r := rand.New(s)

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
	var changed = 0
	fileLength := int64(40_000_000 + 1000)
	bytes := make([]byte, fileLength)
	rand.Read(bytes)
	refFileChan <- bytes
	oldFileContent := make([]byte, fileLength)
	copy(oldFileContent, bytes)
	for i := 0; i < 5; i++ {
		index := rand.Intn(int(fileLength))
		oldFileContent[index] = byte(rand.Int())
	}
	oldFileChan <- oldFileContent
	close(refFileChan)
	close(oldFileChan)
	t.Log("changed bytes:", changed)
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

	windowSize := 100_000
	// signature
	signatureFileName := "__test_signature_file"
	t.Log("calculating signature")
	signatureFlow(oldFileName, signatureFileName, windowSize)
	defer func() {
		err := os.Remove(signatureFileName)
		if err != nil {
			panic(err)
		}
	}()

	// delta
	deltaFileName := "__test_delta_file"
	t.Log("calculating delta")
	deltaFlow(signatureFileName, refFileName, deltaFileName, windowSize)
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
}
