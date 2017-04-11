package tlsh

import (
	"fmt"
	"io/ioutil"
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
		{slice[0], slice[2], slice[3]},
		{slice[0], slice[2], slice[4]},
		{slice[0], slice[1], slice[4]},
		{slice[0], slice[3], slice[4]},
	}
	return triplets
}

func quartilePoints(buckets []byte) (q1, q2, q3 byte) {
	buckets2 := make([]byte, 128)
	copy(buckets2, buckets[:128])
	sortedBuckets := SortByteArray(buckets2)
	// 25%, 50% and 75%
	return sortedBuckets[(128/4)-1], sortedBuckets[(128/2)-1], sortedBuckets[128-(128/4)-1]
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
		biHash += strings.ToUpper(strconv.FormatInt(int64(h), 16))
	}
	return biHash
}

//Hash calculates the TLSH for the input file
func Hash(filename string) (hash string, err error) {
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
		salt := []byte{2, 3, 5, 7, 11, 13}
		for i, triplet := range triplets {
			buckets[pearsonHash(salt[i], triplet)]++
		}
	}
	q1, q2, q3 := quartilePoints(buckets[:128])
	q1Ratio := (q1 * 100 / q3) % 16
	q2Ratio := (q2 * 100 / q3) % 16
	fmt.Println(q1Ratio, q2Ratio)
	//checksum := pearsonHash(0, triplet)
	//fmt.Println(checksum)

	strHash := makeHash(buckets, q1, q2, q3)
	fmt.Println(strHash)
	return
}
