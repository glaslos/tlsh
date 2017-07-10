package tlsh

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"io"
	"math"
	"os"
)

const (
	log1_5       = 0.4054651
	log1_3       = 0.26236426
	log1_1       = 0.095310180
	codeSize     = 32
	windowLength = 5
	effBuckets   = 128
	numBuckets   = 256
)

// LSH holds the hash components
type LSH struct {
	Checksum byte
	Length   byte
}

func quartilePoints(buckets [numBuckets]uint) (q1, q2, q3 uint) {
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

func partition(pbuf *[]uint, left, right uint) uint {
	if left == right {
		return left
	}

	rbuf := (*pbuf)

	if left+1 == right {
		if rbuf[left] > rbuf[right] {
			rbuf[right], rbuf[left] = rbuf[left], rbuf[right]
		}
		return left
	}

	ret := left
	pivot := (left + right) >> 1
	val := rbuf[pivot]

	rbuf[pivot] = rbuf[right]
	rbuf[right] = val

	for i := left; i < right; i++ {
		if rbuf[i] < val {
			rbuf[i], rbuf[ret] = rbuf[ret], rbuf[i]
			ret++
		}
	}

	rbuf[right] = rbuf[ret]
	rbuf[ret] = val

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

func bucketsBinaryRepresentation(buckets [numBuckets]uint, q1, q2, q3 uint) [codeSize]byte {
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

func hashTLSH(length int, buckets [numBuckets]uint, checksum byte, q1, q2, q3 uint) []byte {

	// binary representation of buckets
	biHash := bucketsBinaryRepresentation(buckets, q1, q2, q3)

	q1Ratio := byte(float32(q1)*100/float32(q3)) % 16
	q2Ratio := byte(float32(q2)*100/float32(q3)) % 16

	qRatio := ((q1Ratio & 0xF) << 4) | (q2Ratio & 0xF)

	// prepend header
	return append([]byte{swapByte(checksum), swapByte(lValue(length)), qRatio}, biHash[:]...)
}

func reverse(s [5]byte) [5]byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func fillBuckets(r fuzzyReader) ([numBuckets]uint, byte, int, error) {
	buckets := [numBuckets]uint{}
	chunkSlice := make([]byte, windowLength)
	chunk := [windowLength]byte{}
	salt := [6]byte{2, 3, 5, 7, 11, 13}
	fileSize := 0
	checksum := byte(0)

	n, err := r.Read(chunkSlice)
	if err != nil {
		return [numBuckets]uint{}, 0, 0, err
	}
	copy(chunk[:], chunkSlice[0:5])
	chunk = reverse(chunk)
	fileSize += n

	chunk3 := &[3]byte{}

	for {
		chunk3[0] = chunk[0]
		chunk3[1] = chunk[1]
		chunk3[2] = checksum
		checksum = pearsonHash(0, chunk3)

		chunk3[2] = chunk[2]
		buckets[pearsonHash(salt[0], chunk3)]++

		chunk3[2] = chunk[3]
		buckets[pearsonHash(salt[1], chunk3)]++

		chunk3[1] = chunk[2]
		buckets[pearsonHash(salt[2], chunk3)]++

		chunk3[2] = chunk[4]
		buckets[pearsonHash(salt[3], chunk3)]++

		chunk3[1] = chunk[1]
		buckets[pearsonHash(salt[4], chunk3)]++

		chunk3[1] = chunk[3]
		buckets[pearsonHash(salt[5], chunk3)]++

		copy(chunk[1:], chunk[0:4])
		chunk[0], err = r.ReadByte()
		if err != nil {
			if err != io.EOF {
				return [numBuckets]uint{}, 0, 0, err
			}
			break
		}
		fileSize++
	}
	return buckets, checksum, fileSize, nil
}

type fuzzyReader interface {
	Read([]byte) (int, error)
	ReadByte() (byte, error)
}

//HashReader calculates the TLSH for the input reader
func HashReader(r fuzzyReader) (hash string, err error) {
	buckets, checksum, fileSize, err := fillBuckets(r)
	if err != nil {
		return
	}
	q1, q2, q3 := quartilePoints(buckets)
	hash = hex.EncodeToString(hashTLSH(fileSize, buckets, checksum, q1, q2, q3))

	return hash, nil
}

//HashBytes calculates the TLSH for the input byte slice
func HashBytes(blob []byte) (hash string, err error) {
	r := bytes.NewReader(blob)
	return HashReader(r)
}

//Hash calculates the TLSH for the input file
func Hash(filename string) (hash string, err error) {
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		return
	}
	r := bufio.NewReader(f)
	return HashReader(r)
}
