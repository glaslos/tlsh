package tlsh

import (
	"fmt"
	"io/ioutil"
)

var windowLength = 5
var effBuckets = 256
var codeSize = 32
var numBuckets = 256

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
	buckets2 := make([]byte, effBuckets)
	copy(buckets2, buckets[:effBuckets])
	sortedBuckets := SortByteArray(buckets2)
	// 25%, 50% and 75%
	return sortedBuckets[(effBuckets/4)-1], sortedBuckets[(effBuckets/2)-1], sortedBuckets[effBuckets-(effBuckets/4)-1]
}

func makeHash(buckets []byte, q1, q2, q3 byte) []uint {
	var biHash []uint

	for i := 0; i < codeSize; i++ {
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
		biHash = append([]uint{h}, biHash...)
	}

	return biHash
}

//Hash calculates the TLSH for the input file
func Hash(filename string) (hash string, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	buckets := make([]byte, numBuckets)
	chunk := make([]byte, windowLength)
	sw := 0
	for sw <= len(data)-windowLength {

		for j, x := sw+windowLength-1, 0; j >= sw; j, x = j-1, x+1 {
			chunk[x] = data[j]
		}

		sw++
		triplets := getTriplets(chunk)
		salt := []byte{2, 3, 5, 7, 11, 13}
		for i, triplet := range triplets {
			buckets[pearsonHash(salt[i], triplet)]++
		}
	}
	q1, q2, q3 := quartilePoints(buckets)
	fmt.Println(q1, q2, q3)
	q1Ratio := (q1 * 100 / q3) % 16
	q2Ratio := (q2 * 100 / q3) % 16
	fmt.Println(q1Ratio, q2Ratio)
	//checksum := pearsonHash(0, triplet)
	//fmt.Println(checksum)

	biHash := makeHash(buckets, q1, q2, q3)

	for i := 0; i < len(biHash); i++ {
		hash += fmt.Sprintf("%02X", biHash[i])
	}

	return hash, nil
}
