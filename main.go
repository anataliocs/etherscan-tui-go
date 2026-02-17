// Deprecated: this file has been superseded by cmd/ethereum-explorer/main.go
// and internal packages. It remains here only as a reference or for backwards compatibility.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("This entrypoint is deprecated. Please use: go run ./cmd/ethereum-explorer")
	os.Exit(1)
}
