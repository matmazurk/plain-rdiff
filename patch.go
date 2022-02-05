package main

type ReaderAt interface {
	ReadAt(b []byte, off int64) (n int, err error)
}

func Patch(deltaChunksChan chan DeltaChunk, oldFileReader ReaderAt, newFileWriterChan chan []byte) error {
	defer close(newFileWriterChan)
	for ss := range deltaChunksChan {
		if ss.r != nil {
			from := *ss.r.from
			to := *ss.r.to
			blockLen := to - from
			data := make([]byte, blockLen)
			_, err := oldFileReader.ReadAt(data, int64(from))
			if err != nil {
				return err
			}
			newFileWriterChan <- data
		} else {
			newFileWriterChan <- *ss.d
		}
	}
	return nil
}
