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
	var spl, spr byte
	var p1 byte = byte(effBuckets/4 - 1)
	var p2 byte = byte(effBuckets/2 - 1)
	var p3 byte = byte(effBuckets - effBuckets/4 - 1)
	var end byte = byte(effBuckets - 1)

	bucket_copy := make([]byte, effBuckets)
	copy(bucket_copy, buckets[:effBuckets])

	short_cut_left := make([]byte, effBuckets)
	short_cut_right := make([]byte, effBuckets)

	for l, r := byte(0), end; ; {
		ret := partition(&bucket_copy, l, r)
		if ret > p2 {
			r = ret - 1
			short_cut_right[spr] = ret
			spr++
		} else if ret < p2 {
			l = ret + 1
			short_cut_left[spl] = ret
			spl++
		} else {
			q2 = bucket_copy[p2]
			break
		}
	}

	short_cut_left[spl] = p2 - 1
	short_cut_right[spr] = p2 + 1

	for i, l := byte(0), byte(0); i <= spl; i++ {
		r := short_cut_left[i]
		if r > p1 {
			for ; ; {
				ret := partition(&bucket_copy, l, r)
				if ret > p1 {
					r = ret - 1
				} else if ret < p1 {
					l = ret + 1
				} else {
					q1 = bucket_copy[p1]
					break
				}
			}
			break
		} else if r < p1 {
			l = r
		} else {
			q1 = bucket_copy[p1]
			break
		}
	}

	for i, r := byte(0), end; i <= spr; i++ {
		l := short_cut_right[i]
		if l < p3 {
			for ; ; {
				ret := partition(&bucket_copy, l, r)
				if ret > p3 {
					r = ret - 1
				} else if ret < p3 {
					l = ret + 1
				} else {
					q3 = bucket_copy[p3]
					break
				}
			}
			break
		} else if l > p3 {
			r = l
		} else {
			q3 = bucket_copy[p3]
			break
		}
	}

	return q1, q2, q3
}

func partition(buf *[]byte, left, right byte) byte {

	if left == right {
		return left
	}

	if left + 1 == right {
		if (*buf)[left] > (*buf)[right] {
			(*buf)[right], (*buf)[left] = (*buf)[left], (*buf)[right]
		}
		return left
	}

	var ret byte = left
	var pivot byte = (left + right) >> 1
	var val byte = (*buf)[pivot]

	(*buf)[pivot] = (*buf)[right]
	(*buf)[right] = val

	for i := left; i < right; i++ {
		if ((*buf))[i] < val {
			(*buf)[i], (*buf)[ret] = (*buf)[ret], (*buf)[i]
			ret++
		}
	}

	(*buf)[right] = (*buf)[ret]
	(*buf)[ret] = val

	return ret;
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
