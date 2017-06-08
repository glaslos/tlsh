package main

import (
	"flag"
	"fmt"

	"github.com/glaslos/tlsh"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println("Please provide a file path: ./tlsh /tmp/file")
		return
	}
	fileName := flag.Args()[0]
	hash, err := tlsh.Hash(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s  %s\n", hash, fileName)
}
