/* Copyright 2017 Lukas Rist

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

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

// TLSH holds hash components
type TLSH struct {
	checksum byte
	lValue   byte
	q1Ratio  byte
	q2Ratio  byte
	qRatio   byte
	code     [codeSize]byte
	state    chunkState
}

// New represents type factory for Tlsh
func New(checksum, lValue, q1Ratio, q2Ratio, qRatio byte, code [codeSize]byte) *TLSH {
	return &TLSH{
		checksum: checksum,
		lValue:   lValue,
		q1Ratio:  q1Ratio,
		q2Ratio:  q2Ratio,
		qRatio:   qRatio,
		code:     code,
	}
}

// Binary returns the binary representation of the hash
func (t *TLSH) Binary() []byte {
	return append([]byte{swapByte(t.checksum), swapByte(t.lValue), t.qRatio}, t.code[:]...)
}

// String returns the string representation of the hash`
func (t *TLSH) String() string {
	return hex.EncodeToString(t.Binary())
}

// Parsing the hash of the string type
func ParseStringToTlsh(hashString string) (*TLSH, error) {
	var code [codeSize]byte
	hashByte, err := hex.DecodeString(hashString)
	if err != nil {
		return &TLSH{}, err
	}
	chechsum := swapByte(hashByte[0])
	lValue := swapByte(hashByte[1])
	qRatio := hashByte[2]
	q1Ratio := (qRatio >> 4) & 0xF
	q2Ratio := qRatio & 0xF
	copy(code[:], hashByte[3:])
	return New(chechsum, lValue, q1Ratio, q2Ratio, qRatio, code), nil
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
		ret := partition(bucketCopy, l, r)
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
				ret := partition(bucketCopy, l, r)
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
				ret := partition(bucketCopy, l, r)
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

func partition(buf []uint, left, right uint) uint {
	if left == right {
		return left
	}

	if left+1 == right {
		if buf[left] > buf[right] {
			buf[right], buf[left] = buf[left], buf[right]
		}
		return left
	}

	ret := left
	pivot := (left + right) >> 1
	val := buf[pivot]

	buf[pivot] = buf[right]
	buf[right] = val

	for i := left; i < right; i++ {
		if buf[i] < val {
			buf[i], buf[ret] = buf[ret], buf[i]
			ret++
		}
	}

	buf[right] = buf[ret]
	buf[ret] = val

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

func reverse(s [5]byte) [5]byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func fillBuckets(r FuzzyReader) ([numBuckets]uint, byte, int, error) {
	buckets := [numBuckets]uint{}
	chunkSlice := make([]byte, windowLength)
	chunk := [windowLength]byte{}
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

// hashCalculate calculate TLSH
func hashCalculate(r FuzzyReader) (*TLSH, error) {
	buckets, checksum, fileSize, err := fillBuckets(r)
	if err != nil {
		return &TLSH{}, err
	}

	q1, q2, q3 := quartilePoints(buckets)
	q1Ratio := byte(float32(q1)*100/float32(q3)) % 16
	q2Ratio := byte(float32(q2)*100/float32(q3)) % 16
	qRatio := ((q1Ratio & 0xF) << 4) | (q2Ratio & 0xF)

	biHash := bucketsBinaryRepresentation(buckets, q1, q2, q3)

	t := New(checksum, lValue(fileSize), q1Ratio, q2Ratio, qRatio, biHash)
	t.state = chunkState{
		buckets:  buckets,
		fileSize: fileSize,
		checksum: checksum,
	}
	return t, nil
}

// FuzzyReader interface
type FuzzyReader interface {
	io.Reader
	io.ByteReader
}

//HashReader calculates the TLSH for the input reader
func HashReader(r FuzzyReader) (*TLSH, error) {
	t, err := hashCalculate(r)
	if err != nil {
		return &TLSH{}, err
	}
	return t, err
}

//HashBytes calculates the TLSH for the input byte slice
func HashBytes(blob []byte) (*TLSH, error) {
	r := bytes.NewReader(blob)
	return HashReader(r)
}

//HashFilename calculates the TLSH for the input file
func HashFilename(filename string) (*TLSH, error) {
	f, err := os.Open(filename)
	if err != nil {
		return &TLSH{}, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	return HashReader(r)
}

// Diff current hash with other hash
func (t *TLSH) Diff(t2 *TLSH) int {
	return diffTotal(t, t2, true)
}

// DiffFilenames calculate distance between two files
func DiffFilenames(filenameA, filenameB string) (int, error) {
	f, err := os.Open(filenameA)
	if err != nil {
		return -1, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	tlshA, err := hashCalculate(r)
	if err != nil {
		return -1, err
	}

	f, err = os.Open(filenameB)
	if err != nil {
		return -1, err
	}
	defer f.Close()

	r = bufio.NewReader(f)
	tlshB, err := hashCalculate(r)
	if err != nil {
		return -1, err
	}

	return tlshA.Diff(tlshB), nil
}
