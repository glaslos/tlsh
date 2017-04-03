package tlsh

import (
	"fmt"
	"io/ioutil"
	"math/rand"
)

var windowLength = 5

func getTriplets(slice []byte) (res [][]byte) {
	res = [][]byte{
		{slice[0], slice[1], slice[2]},
		{slice[0], slice[1], slice[3]},
		{slice[0], slice[1], slice[4]},
		{slice[0], slice[2], slice[3]},
		{slice[0], slice[2], slice[4]},
		{slice[0], slice[3], slice[4]},
	}
	return res
}

func phashOriginal(keys []byte) (h byte) {
	table := make([]byte, 256)
	for i := 0; i <= 255; i++ {
		table[i] = byte(i)
	}
	for i := range table {
	    j := rand.Intn(i + 1)
	    table[i], table[j] = table[j], table[i]
	}
	for _, c := range keys {
		h = table[h^c]
	}
	return
}

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
		triplets := getTriplets(chunk)
		fmt.Println(phashOriginal(triplets[1]))
		break
	}
	return
}
