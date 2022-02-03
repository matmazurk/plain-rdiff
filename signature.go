package main

import "encoding/binary"

func CalculateAndSendChecksums(bufferedReader bufferedReader, checksumsChan chan []byte) error {
	defer close(checksumsChan)

	for {
		anyReads, err := bufferedReader.ReadWindow()
		if err != nil {
			return err
		}
		if !anyReads {
			return nil
		}

		checksumsChan <- getChecksumsBundle(calculateChecksum(bufferedReader), bufferedReader.MD4())

		if bufferedReader.isEOF() {
			return nil
		}
	}
}

func calculateChecksum(br bufferedReader) uint32 {
	a := br.SumBufferBytes()

	var b uint32
	offsetedLen := br.Offset() + int64(br.Len())
	for ii, i := br.Offset(), 0; ii < int64(offsetedLen); i++ {
		b += uint32(offsetedLen-ii) * uint32(br.Get(i))
		ii++
	}
	return b<<16 | a
}

func getChecksumsBundle(rollingChecksum uint32, hash []byte) []byte {
	checksums := make([]byte, 4)
	binary.BigEndian.PutUint32(checksums, rollingChecksum)
	return append(checksums, hash...)
}
