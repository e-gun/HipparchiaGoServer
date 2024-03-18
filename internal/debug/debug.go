package debug

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/m"
)

var Msg = m.NewMessageMaker()

//
// DEBUGGING
//

// stringkeyprinter - print out the keys of a map
func stringkeyprinter[T any](n string, m map[string]T) {
	Msg.WARN(n)
	counter := 0
	for k, _ := range m {
		fmt.Printf("[%d] %s\n", counter, k)
		counter += 1
	}
}

// stringmapprinter - print out the k/v pairs of a map
func stringmapprinter[T any](n string, m map[string]T) {
	Msg.WARN(n)
	counter := 0
	for k, v := range m {
		fmt.Printf("[%d] %s\t", counter, k)
		fmt.Println(v)
		counter += 1
	}
}

// sliceprinter - print out the members of a slice
func sliceprinter[T any](n string, s []T) {
	Msg.WARN(n)
	for i, v := range s {
		fmt.Printf("[%d]\t", i)
		fmt.Println(v)
	}
}
