package tlsh

import (
	"hash"
)

var salt = [6]byte{2, 3, 5, 7, 11, 13}

type chunkState struct {
	buckets    [numBuckets]uint
	chunk      [windowLength]byte
	chunkSlice []byte
	fileSize   int
	checksum   byte
	chunk3     *[3]byte
}

var _ hash.Hash = &TLSH{}

func (t *TLSH) Reset() {
	t.checksum = byte(0)
	t.lValue = 0
	t.q1Ratio = 0
	t.q2Ratio = 0
	t.qRatio = 0
	t.code = [codeSize]byte{}
	t.state = chunkState{
		buckets:    [numBuckets]uint{},
		chunk:      [windowLength]byte{},
		chunkSlice: []byte{},
		fileSize:   0,
		checksum:   byte(0),
		chunk3:     &[3]byte{},
	}
}

func (t *TLSH) BlockSize() int {
	return 1
}

func (t *TLSH) Size() int {
	return len(t.Binary())
}

func (t *TLSH) Sum(b []byte) []byte {
	q1, q2, q3 := quartilePoints(t.state.buckets)
	q1Ratio := byte(float32(q1)*100/float32(q3)) % 16
	q2Ratio := byte(float32(q2)*100/float32(q3)) % 16
	qRatio := ((q1Ratio & 0xF) << 4) | (q2Ratio & 0xF)

	biHash := bucketsBinaryRepresentation(t.state.buckets, q1, q2, q3)

	*t = *new(t.state.checksum, lValue(t.state.fileSize), q1Ratio, q2Ratio, qRatio, biHash, t.state)
	return t.Binary()
}

func (s *chunkState) process() {
	s.chunk3[0] = s.chunk[0]
	s.chunk3[1] = s.chunk[1]
	s.chunk3[2] = s.checksum
	s.checksum = pearsonHash(0, s.chunk3)

	s.chunk3[2] = s.chunk[2]
	s.buckets[pearsonHash(salt[0], s.chunk3)]++

	s.chunk3[2] = s.chunk[3]
	s.buckets[pearsonHash(salt[1], s.chunk3)]++

	s.chunk3[1] = s.chunk[2]
	s.buckets[pearsonHash(salt[2], s.chunk3)]++

	s.chunk3[2] = s.chunk[4]
	s.buckets[pearsonHash(salt[3], s.chunk3)]++

	s.chunk3[1] = s.chunk[1]
	s.buckets[pearsonHash(salt[4], s.chunk3)]++

	s.chunk3[1] = s.chunk[3]
	s.buckets[pearsonHash(salt[5], s.chunk3)]++

	copy(s.chunk[1:], s.chunk[0:4])
}

func (t *TLSH) Write(p []byte) (int, error) {
	t.state.fileSize += len(p)
	if len(t.state.chunkSlice) < windowLength {
		missing := windowLength - len(t.state.chunkSlice)
		switch {
		case len(p) < missing:
			t.state.chunkSlice = append(t.state.chunkSlice, p...)
			return len(p), nil
		default:
			t.state.chunkSlice = append(t.state.chunkSlice, p[0:missing]...)
			p = p[missing:]
			copy(t.state.chunk[:], t.state.chunkSlice[0:5])
			t.state.chunk = reverse(t.state.chunk)
			t.state.process()
		}
	}
	for _, b := range p {
		t.state.chunk[0] = b
		t.state.process()
	}

	return len(p), nil
}
