package main

import (
	"fmt"
	"os"
)

var (
	// Program version information.
	version string
	// Program build date.
	buildDate string
)

func main() {
	fmt.Println(os.Args[0], version, buildDate)
}
