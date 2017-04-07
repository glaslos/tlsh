package tlsh

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
)

var windowLength = 5

// LSH holds the hash components
type LSH struct {
	Checksum byte
	Length   byte
}

func getTriplets(slice []byte) (triplets [][]byte) {
	triplets = [][]byte{
		{slice[0], slice[1], slice[2]},
		{slice[0], slice[1], slice[3]},
		{slice[0], slice[1], slice[4]},
		{slice[0], slice[2], slice[3]},
		{slice[0], slice[2], slice[4]},
		{slice[0], slice[3], slice[4]},
	}
	return triplets
}

func mappingTable() (table []byte) {
	table = make([]byte, 256)
	for i := 0; i <= 255; i++ {
		table[i] = byte(i)
	}
	for i := range table {
		j := rand.Intn(i + 1)
		table[i], table[j] = table[j], table[i]
	}
	return
}

func pearsonHash(keys []byte, table []byte) (h byte) {
	for _, c := range keys {
		h = table[h^c]
	}
	return
}

func quartilePoints(buckets []byte) (q1, q2, q3 byte) {
	sortedBuckets := SortByteArray(buckets)
	// 75%, 50% and 25%
	return sortedBuckets[64], sortedBuckets[128], sortedBuckets[192]
}

func makeHash(buckets []byte, q1, q2, q3 byte) string {
	var biHash string
	for i := 0; i < 31; i++ {
		var h uint
		for j := 0; j < 4; j++ {
			k := buckets[4*i+j]
			if q3 < k {
				h += 3 << (uint(j) * 2)
			} else if q2 < k {
				h += 2 << (uint(j) * 2)
			} else if q1 < k {
				h += 1 << (uint(j) * 2)
			}
		}
		//h = h % 255
		biHash += strings.ToUpper(strconv.FormatInt(int64(h), 16))
	}
	return biHash
}

//Hash calculates the TLSH for the input file
func Hash(filename string) (hash string, err error) {
	table := mappingTable()
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	buckets := make([]byte, 256)
	sw := 0
	for sw <= len(data)-windowLength {
		chunk := data[sw : sw+windowLength]
		sw += windowLength
		triplets := getTriplets(chunk)
		for _, triplet := range triplets {
			buckets[pearsonHash(triplet, table)]++
		}
	}
	buckets2 := make([]byte, 256)
	copy(buckets2, buckets)
	q1, q2, q3 := quartilePoints(buckets2)
	q1Ratio := (q1 * 100 / q3) % 16
	q2Ratio := (q2 * 100 / q3) % 16
	fmt.Println(q1Ratio, q2Ratio)
	//checksum := pearsonHash(triplet, table)

	strHash := makeHash(buckets, q1, q2, q3)
	fmt.Println(strHash)
	return
}
