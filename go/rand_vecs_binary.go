package main

import (
	"os"
	"unsafe"
)

func main() {
	// Example slice of float32
	slice := []float32{1.23, 4.56, 7.89}

	// Convert the []float32 to []byte
	byteSlice := float32SliceToBytes(slice)

	// Write the raw bytes to stdout
	os.Stdout.Write(byteSlice)
}

// float32SliceToBytes converts a []float32 to a []byte without copying the data.
func float32SliceToBytes(slice []float32) []byte {
	// Calculate the size of the byte slice
	size := len(slice) * int(unsafe.Sizeof(slice[0]))

	// Use unsafe to cast the []float32 to a []byte
	return (*[1 << 30]byte)(unsafe.Pointer(&slice[0]))[:size:size]
}

