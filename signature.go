package main

import (
	"encoding/binary"

	"golang.org/x/crypto/md4"
)

func CalculateAndSendChecksums(
	bufferedReader bufferedReader,
	checksumsChan chan []byte,
	checksumCalculation func([]byte) uint32,
) error {
	defer close(checksumsChan)

	for {
		anyReads, err := bufferedReader.ReadWindow()
		if err != nil {
			return err
		}
		if !anyReads {
			return nil
		}

		checksum := checksumCalculation(bufferedReader.Buf())
		checksumsChan <- getBundle(checksum, bufferedReader.GetHash(calculateMD4))

		if bufferedReader.isEOF() {
			return nil
		}
	}
}

func getBundle(rollingChecksum uint32, hash []byte) []byte {
	checksum := make([]byte, 4)
	binary.BigEndian.PutUint32(checksum, rollingChecksum)
	return append(checksum, hash...)
}

func calculateMD4(data []byte) []byte {
	h := md4.New()
	h.Write(data)
	return h.Sum(nil)
}
