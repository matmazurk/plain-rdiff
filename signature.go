package main

import "encoding/binary"

func CalculateAndSendChecksums(bufferedReader bufferedReader, checksumsChan chan []byte) error {
	for {
		err := bufferedReader.ReadWindow()
		if err != nil {
			return err
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
	for ii, i := br.Offset(), 0; i < int(offsetedLen); i++ {
		b += uint32(offsetedLen-ii+1) * uint32(br.Get(i))
	}
	return b<<16 | a
}

func getChecksumsBundle(rollingChecksum uint32, hash []byte) []byte {
	var rollingChecksumBytes []byte
	binary.BigEndian.PutUint32(rollingChecksumBytes, rollingChecksum)
	return append(rollingChecksumBytes, hash...)
}
