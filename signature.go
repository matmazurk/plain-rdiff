package main

import "encoding/binary"

func CalculateAndSendChecksums(
	bufferedReader bufferedReader,
	checksumsChan chan []byte,
	checksumCalculation func([]byte, int64) uint32,
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

		checksum := checksumCalculation(bufferedReader.Buf(), bufferedReader.Offset())
		checksumsChan <- getChecksumsBundle(checksum, bufferedReader.MD4())

		if bufferedReader.isEOF() {
			return nil
		}
	}
}

func getChecksumsBundle(rollingChecksum uint32, hash []byte) []byte {
	checksums := make([]byte, 4)
	binary.BigEndian.PutUint32(checksums, rollingChecksum)
	return append(checksums, hash...)
}
