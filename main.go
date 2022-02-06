package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	MODE_SIGNATURE = "signature"
	MODE_DELTA     = "delta"
	MODE_PATCH     = "patch"
)

const (
	USAGE_TEXT      = "Usage:\n rdiff signature old-file signature-file\n rdiff delta signature-file new-file delta-file\n rdiff patch basis-file delta-file new-file"
	SIGNATURE_USAGE = "Signature usage:\n rdiff signature old-file signature-file"
	DELTA_USAGE     = "Delta usage:\n rdiff delta signature-file new-file delta-file"
	PATCH_USAGE     = "Patch usage:\n rdiff patch basis-file delta-file new-file"
)

const WINDOW_LENGTH = 5000

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatal(USAGE_TEXT)
		return
	}
	switch os.Args[1] {
	case MODE_SIGNATURE:
		if len(os.Args) != 4 {
			log.Fatal(SIGNATURE_USAGE)
		}
		oldFile := os.Args[2]
		signatureFile := os.Args[3]
		if !exists(fmt.Sprintf("%s/%s", getExecutionDir(), oldFile)) {
			log.Fatalf("provided old file doesn't exist")
		}
		if exists(fmt.Sprintf("%s/%s", getExecutionDir(), signatureFile)) {
			log.Fatalf("provided signature file already exists")
		}
		signatureFlow(oldFile, signatureFile)
	case MODE_DELTA:
		if len(os.Args) != 5 {
			log.Fatal(DELTA_USAGE)
		}
		signatureFile := os.Args[2]
		newFile := os.Args[3]
		deltaFile := os.Args[4]
		if !exists(fmt.Sprintf("%s/%s", getExecutionDir(), signatureFile)) {
			log.Fatalf("provided signature file doesn't exist")
		}
		if !exists(fmt.Sprintf("%s/%s", getExecutionDir(), newFile)) {
			log.Fatalf("provided new file doesn't exist")
		}
		if exists(fmt.Sprintf("%s/%s", getExecutionDir(), deltaFile)) {
			log.Fatalf("provided delta file already exists")
		}
		deltaFlow(signatureFile, newFile, deltaFile)
	case MODE_PATCH:
		if len(os.Args) != 5 {
			log.Fatal(PATCH_USAGE)
		}
		basisFile := os.Args[2]
		deltaFile := os.Args[3]
		newFile := os.Args[4]
		if !exists(fmt.Sprintf("%s/%s", getExecutionDir(), basisFile)) {
			log.Fatalf("provided basis file doesn't exist")
		}
		if !exists(fmt.Sprintf("%s/%s", getExecutionDir(), deltaFile)) {
			log.Fatalf("provided delta file doesn't exist")
		}
		if exists(fmt.Sprintf("%s/%s", getExecutionDir(), newFile)) {
			log.Fatalf("provided new file already exists")
		}
		patchFlow(basisFile, deltaFile, newFile)
	default:
		fmt.Println(USAGE_TEXT)
	}
}

func getExecutionDir() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(ex)
}

// exists returns whether the given file or directory exists
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func signatureFlow(oldFilePath, signatureFilePath string) {
	c := make(chan []byte)

	go func() {
		reader, err := GetFileReader(oldFilePath)
		if err != nil {
			log.Fatal(err)
		}

		bufferedReader := NewBufferedReader(WINDOW_LENGTH, reader)
		err = CalculateAndSendChecksums(bufferedReader, c, CalculateChecksumWithoutPreviousCompounds)
		if err != nil {
			log.Fatal(err)
		}
	}()

	err := CreateAndFillFile(signatureFilePath, c)
	if err != nil {
		log.Fatal(err)
	}
}

func deltaFlow(signatureFilePath, newFilePath, deltaFilePath string) {
	bundles, err := ReadSignatureFile(signatureFilePath)
	if err != nil {
		log.Fatal(err)
	}

	reader, err := GetFileReader(newFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	c := make(chan DeltaChunk)
	go func() {
		checksums, hashes := getRollingChecksumAndHashes(bundles)
		br := NewBufferedReader(WINDOW_LENGTH, reader)
		err := CalculateAndSendDeltaChunks(
			br,
			c,
			checksums,
			hashes,
			findMatchingOffset,
			CalculateChecksum,
			calculateMD4,
		)
		if err != nil {
			panic(err)
		}
	}()

	err = CreateAndFillDeltaFile(deltaFilePath, c)
	if err != nil {
		log.Fatal(err)
	}
}

func patchFlow(basisFilePath, deltaFilePath, newFilePath string) {
	c := make(chan DeltaChunk)
	newFileWriterChan := make(chan []byte)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		f, err := GetFileReader(deltaFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		err = DeltaReader(f, c)
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		err := CreateAndFillFile(newFilePath, newFileWriterChan)
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()

	f, err := GetFileReader(basisFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = Patch(c, f, newFileWriterChan)
	if err != nil {
		panic(err)
	}
	wg.Wait()
}

func getRollingChecksumAndHashes(bundles [][]byte) (map[uint32]int, [][]byte) {
	rollingChecksumsToIndexes := make(map[uint32]int, len(bundles))
	hashes := make([][]byte, len(bundles))
	for i, b := range bundles {
		checksum := b[:4]
		uintChecksum := binary.BigEndian.Uint32(checksum)
		rollingChecksumsToIndexes[uintChecksum] = i
		hashes[i] = b[4:]
	}
	return rollingChecksumsToIndexes, hashes
}
