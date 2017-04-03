package tlsh

import (
	"io/ioutil"
)

var windowLength = 5

//Hash calculates the TLSH for the input file
func Hash(filename string) (hash string, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	sw := 0
	for sw <= len(data)-windowLength {
		chunk := data[sw : sw+windowLength]
		sw += windowLength
	}
	return
}
