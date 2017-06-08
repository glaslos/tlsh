package tlsh

import (
	"fmt"
	"io/ioutil"
	"math"
)

const (
	log1_5   = 0.4054651
	log1_3   = 0.26236426
	log1_1   = 0.095310180
	codeSize = 32
)

var (
	windowLength = 5
	effBuckets   = 128
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

func quartilePoints(buckets []uint) (q1, q2, q3 uint) {
	var spl, spr uint
	p1 := uint(effBuckets/4 - 1)
	p2 := uint(effBuckets/2 - 1)
	p3 := uint(effBuckets - effBuckets/4 - 1)
	end := uint(effBuckets - 1)

	bucketCopy := make([]uint, effBuckets)
	copy(bucketCopy, buckets[:effBuckets])

	shortCutLeft := make([]uint, effBuckets)
	shortCutRight := make([]uint, effBuckets)

	for l, r := uint(0), end; ; {
		ret := partition(&bucketCopy, l, r)
		if ret > p2 {
			r = ret - 1
			shortCutRight[spr] = ret
			spr++
		} else if ret < p2 {
			l = ret + 1
			shortCutLeft[spl] = ret
			spl++
		} else {
			q2 = bucketCopy[p2]
			break
		}
	}

	shortCutLeft[spl] = p2 - 1
	shortCutRight[spr] = p2 + 1

	for i, l := uint(0), uint(0); i <= spl; i++ {
		r := shortCutLeft[i]
		if r > p1 {
			for {
				ret := partition(&bucketCopy, l, r)
				if ret > p1 {
					r = ret - 1
				} else if ret < p1 {
					l = ret + 1
				} else {
					q1 = bucketCopy[p1]
					break
				}
			}
			break
		} else if r < p1 {
			l = r
		} else {
			q1 = bucketCopy[p1]
			break
		}
	}

	for i, r := uint(0), end; i <= spr; i++ {
		l := shortCutRight[i]
		if l < p3 {
			for {
				ret := partition(&bucketCopy, l, r)
				if ret > p3 {
					r = ret - 1
				} else if ret < p3 {
					l = ret + 1
				} else {
					q3 = bucketCopy[p3]
					break
				}
			}
			break
		} else if l > p3 {
			r = l
		} else {
			q3 = bucketCopy[p3]
			break
		}
	}

	return q1, q2, q3
}

func partition(buf *[]uint, left, right uint) uint {

	if left == right {
		return left
	}

	if left+1 == right {
		if (*buf)[left] > (*buf)[right] {
			(*buf)[right], (*buf)[left] = (*buf)[left], (*buf)[right]
		}
		return left
	}

	ret := left
	pivot := (left + right) >> 1
	val := (*buf)[pivot]

	(*buf)[pivot] = (*buf)[right]
	(*buf)[right] = val

	for i := left; i < right; i++ {
		if (*buf)[i] < val {
			(*buf)[i], (*buf)[ret] = (*buf)[ret], (*buf)[i]
			ret++
		}
	}

	(*buf)[right] = (*buf)[ret]
	(*buf)[ret] = val

	return ret
}

func lValue(length int) byte {
	var l byte

	if length <= 656 {
		l = byte(math.Floor(math.Log(float64(length)) / log1_5))
	} else if length <= 3199 {
		l = byte(math.Floor(math.Log(float64(length))/log1_3 - 8.72777))
	} else {
		l = byte(math.Floor(math.Log(float64(length))/log1_1 - 62.5472))
	}

	return l % 255
}

func swapByte(in byte) byte {
	var out byte

	out = ((in & 0xF0) >> 4) & 0x0F
	out |= ((in & 0x0F) << 4) & 0xF0

	return out
}

func bucketsBinaryRepresentation(buckets []uint, q1, q2, q3 uint) [codeSize]byte {
	var biHash [codeSize]byte

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
		// Prepend the new h to the hash
		biHash[(codeSize-1)-i] = h
	}
	return biHash
}

func hashTLSH(length int, buckets []uint, checksum byte, q1, q2, q3 uint) []byte {

	// binary representation of buckets
	biHash := bucketsBinaryRepresentation(buckets, q1, q2, q3)

	q1Ratio := byte(float32(q1)*100/float32(q3)) % 16
	q2Ratio := byte(float32(q2)*100/float32(q3)) % 16

	qRatio := ((q1Ratio & 0xF) << 4) | (q2Ratio & 0xF)

	// prepend header
	return append([]byte{swapByte(checksum), swapByte(lValue(length)), qRatio}, biHash[:]...)
}

func makeStringTLSH(biHash []byte) (hash string) {

	for i := 0; i < len(biHash); i++ {
		hash += fmt.Sprintf("%02X", biHash[i])
	}

	return
}

func fillBuckets(data []byte) ([]uint, byte) {
	chunk := make([]byte, windowLength)
	buckets := make([]uint, numBuckets)
	checksum := byte(0)
	salt := []byte{2, 3, 5, 7, 11, 13}
	sw := 0

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
	return buckets, checksum
}

//Hash calculates the TLSH for the input file
func Hash(filename string) (hash string, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	buckets, checksum := fillBuckets(data)
	q1, q2, q3 := quartilePoints(buckets)
	hash = makeStringTLSH(hashTLSH(len(data), buckets, checksum, q1, q2, q3))

	return hash, nil
}
