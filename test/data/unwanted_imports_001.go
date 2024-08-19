package foo

import (
	"fmt"
	// JUSTIFY(unwanted-imports): unwanted_imports_001.go/001
	"io/ioutil"
)

func main() {
	ioutil.ReadDir("foo")
	fmt.Printf("foo")
}
