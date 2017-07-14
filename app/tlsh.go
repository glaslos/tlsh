package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/glaslos/tlsh"
)

var (
	// VERSION is set by the makefile
	VERSION = "v0.0.0"
	// BUILDDATE is set by the makefile
	BUILDDATE = ""
)

var file string
var raw bool
var version bool

func init() {
	flag.StringVar(&file, "f", "", "path to the `file` to be hashed")
	flag.BoolVar(&raw, "r", false, "set to get only the hash")
	flag.BoolVar(&version, "version", false, "print version")
	flag.Parse()
}

// Main contains the main code
func Main() {
	if version {
		fmt.Printf("%s %s\n", VERSION, BUILDDATE)
		return
	}
	if file == "" {
		fmt.Fprintf(os.Stderr, "Usage of %s [-f <file>]\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Println()
		return
	}
	hash, err := tlsh.HashFilename(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	if raw {
		fmt.Println(hash)
	} else {
		fmt.Printf("%s  %s\n", hash, file)
	}
}

func main() {
	Main()
}
