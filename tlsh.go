package tlsh

import (
	"fmt"
	"io/ioutil"
	"math"
)

const (
	LOG_1_5 = 0.4054651
	LOG_1_3 = 0.26236426
	LOG_1_1 = 0.095310180
)

var (
	windowLength = 5
	effBuckets   = 256
	codeSize     = 32
	numBuckets   = 256
)

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

func lValue(length int) byte {
	var l byte

	if length <= 656 {
		l = byte(math.Floor(math.Log(float64(length)) / LOG_1_5))
	} else if length <= 3199 {
		l = byte(math.Floor(math.Log(float64(length))/LOG_1_3 - 8.72777))
	} else {
		l = byte(math.Floor(math.Log(float64(length))/LOG_1_1 - 62.5472))
	}

	return l % 255
}

func swapByte(in byte) byte {
	var out byte

	out = ((in & 0xF0) >> 4) & 0x0F
	out |= ((in & 0x0F) << 4) & 0xF0

	return out
}

func makeHash(buckets []byte, q1, q2, q3 byte) []byte {
	var biHash []byte

	for i := 0; i < codeSize; i++ {
		var h byte
		for j := 0; j < 4; j++ {
			k := buckets[4*i+j]
			if q3 < k {
				h += 3 << (byte(j) * 2)
			} else if q2 < k {
				h += 2 << (byte(j) * 2)
			} else if q1 < k {
				h += 1 << (byte(j) * 2)
			}
		}
		biHash = append([]byte{h}, biHash...)
	}

	return biHash
}

//Hash calculates the TLSH for the input file
func Hash(filename string) (hash string, err error) {

	buckets := make([]byte, numBuckets)
	chunk := make([]byte, windowLength)
	salt := []byte{2, 3, 5, 7, 11, 13}
	sw := 0
	checksum := byte(0)

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	for sw <= len(data)-windowLength {

		for j, x := sw+windowLength-1, 0; j >= sw; j, x = j-1, x+1 {
			chunk[x] = data[j]
		}

		sw++
		triplets := getTriplets(chunk)

		checksumTriplet := []byte{chunk[0], chunk[1], checksum}
		checksum = pearsonHash(0, checksumTriplet)

		for i, triplet := range triplets {
			buckets[pearsonHash(salt[i], triplet)]++
		}
	}
	q1, q2, q3 := quartilePoints(buckets)
	q1Ratio := (q1 * 100 / q3) % 16
	q2Ratio := (q2 * 100 / q3) % 16
	lValue := lValue(len(data))
	biHash := makeHash(buckets, q1, q2, q3)

	fmt.Printf("checksum=%02X\n", swapByte(checksum))
	fmt.Printf("L=%02X\n", swapByte(lValue))
	fmt.Printf("q1Ratio=%X, q2Ratio=%x\n", q1Ratio, q2Ratio)

	for i := 0; i < len(biHash); i++ {
		hash += fmt.Sprintf("%02X", biHash[i])
	}

	return hash, nil
}
