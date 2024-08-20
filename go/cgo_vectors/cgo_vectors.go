// main.go
package main

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lmylib
#include "mylib.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func main() {
	a := []float32{1.1, 2.2, 3.3}
	b := []float32{4.4, 5.5, 6.6}
	result := make([]float32, len(a))

	// Call the C function
	C.add_vectors(
		(*C.float)(unsafe.Pointer(&a[0])),
		(*C.float)(unsafe.Pointer(&b[0])),
		(*C.float)(unsafe.Pointer(&result[0])),
		C.int(len(a)),
	)

	fmt.Printf("Result: %v\n", result)
}
