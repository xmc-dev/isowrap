package main

import (
	"fmt"
	"unsafe"
)

func main() {
	p := unsafe.Pointer(uintptr(0xDEADBEEF))
	d := (*int)(p)

	*d = 2
	fmt.Println(d)
}
