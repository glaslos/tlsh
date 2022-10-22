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

var (
	file    string
	compare string
	raw     bool
	version bool
)

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
	if compare != "" {
		hashCompare, err := tlsh.HashFilename(compare)
		if err != nil {
			fmt.Println(err)
			return
		}
		distance := hash.Diff(hashCompare)

		fmt.Printf("%d  %s  %s - %s  %s\n", distance, hash, file, hashCompare, compare)
	} else {
		if raw {
			fmt.Println(hash)
		} else {
			fmt.Printf("%s  %s\n", hash, file)
		}
	}
}

func main() {
	flag.StringVar(&file, "f", "", "path to the `file` to be hashed")
	flag.StringVar(&compare, "c", "", "specifies a `filename` or `digest` whose TLSH value will be compared to a filename specified (-f)")
	flag.BoolVar(&raw, "r", false, "set to get only the hash")
	flag.BoolVar(&version, "version", false, "print version")
	flag.Parse()
	Main()
}
