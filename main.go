package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
