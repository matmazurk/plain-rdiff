package main

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const BUNDLE_SIZE = 20

func GetFileReader(fileName string) (*os.File, error) {
	f, err := os.Open(fileName)
	return f, err
}

func CreateAndFillFile(filePath string, c chan []byte) error {
	newFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		err = newFile.Close()
	}()

	for bytes := range c {
		writ, err := newFile.Write(bytes)
		if err != nil {
			println(writ)
			return err
		}
	}
	return nil
}

func CreateAndFillDeltaFile(filePath string, c chan DeltaChunk) error {
	newFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	for deltaChunk := range c {
		newFile.Write(deltaChunk.ToBytes())
		if !deltaChunk.r.empty() {
			deltaChunk.r.clear()
		}
	}
	return newFile.Close()
}

func ReadSignatureFile(filePath string) ([][]byte, error) {
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	slices := make([][]byte, len(contents)/BUNDLE_SIZE)
	for i := range slices {
		slices[i] = contents[(i * BUNDLE_SIZE) : (i+1)*BUNDLE_SIZE]
	}

	return slices, nil
}

func DeltaReader(delta *os.File, c chan DeltaChunk) error {
	defer close(c)
	for {
		b := make([]byte, 1)
		_, err := delta.Read(b)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		switch b[0] {
		case 0:
			blockLenBytes := make([]byte, 8)
			_, err := delta.Read(blockLenBytes)
			if err != nil {
				return err
			}
			blockLen := binary.BigEndian.Uint64(blockLenBytes)
			rawData := make([]byte, blockLen)
			_, err = delta.Read(rawData)
			if err != nil {
				return err
			}
			c <- NewDeltaChunkWithRawData(rawData)
		case 1:
			fromBytes := make([]byte, 8)
			_, err = delta.Read(fromBytes)
			if err != nil {
				return err
			}
			from := binary.BigEndian.Uint64(fromBytes)
			toBytes := make([]byte, 8)
			_, err = delta.Read(toBytes)
			if err != nil {
				return err
			}
			to := binary.BigEndian.Uint64(toBytes)
			r := Range{&from, &to}
			c <- NewDeltaChunkWithRange(r)
		default:
			return nil
		}
	}
}
