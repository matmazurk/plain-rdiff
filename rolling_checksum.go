package main

func CalculateChecksumWithoutPreviousCompounds(data []byte) (uint32, *uint32, *uint32) {
	var a uint32
	var b uint32
	for index, singleByte := range data {
		uintByte := uint32(singleByte)
		a += uintByte
		b += uint32(len(data)-index) * uint32(data[index])
	}
	return b<<16 | a, &a, &b
}

func CalculateChecksumUsingPreviousCompounds(
	data []byte,
	previous byte,
	length int,
	a, b uint32,
) (uint32, *uint32, *uint32) {
	prev := uint32(previous)
	dataLen := len(data)

	if len(data) < length {
		A := a - prev
		B := b - uint32(dataLen+1)*prev
		return B<<16 | A, &A, &B
	}

	A := a - prev + uint32(data[dataLen-1])
	B := b - uint32(dataLen)*prev + A

	return B<<16 | A, &A, &B
}
