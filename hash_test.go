package tlsh

import (
	"fmt"
	"hash"
	"testing"
)

func TestHashinterface(t *testing.T) {
	h := New()
	var _ hash.Hash = h
	t.Log(h.BlockSize())
	h.Reset()
	n, err := h.Write([]byte("hello"))
	if err != nil {
		t.Error(err)
	}
	t.Log(n)
	t.Log(h.Size())
	hash := h.Sum(nil)
	t.Log(hash)
	t.Log(fmt.Sprintf("%x", hash[:]))
}

func TestHashWrite(t *testing.T) {
	// hash using the hash.Hash interface methods
	h1 := New()
	h1.Write([]byte("1234"))
	h1.Write([]byte("11"))
	h1.Write([]byte("1111111"))
	t.Log(fmt.Sprintf("h1: %x", h1.Sum(nil)))
	t.Logf("checksum h1: %d, %x", h1.state.checksum, h1.checksum)

	// hash from read
	h2, err := HashBytes([]byte("1234111111111"))
	if err != nil {
		t.Error(err)
	}
	t.Logf("checksum h2: %d, %x", h2.state.checksum, h2.checksum)
	t.Log(fmt.Sprintf("h2: %x", h2.Binary()))

	// compare hashes
	if h1.state.fileSize != h2.state.fileSize {
		t.Errorf("file size mismatch: %d != %d", h1.state.fileSize, h2.state.fileSize)
	}
	if h1.checksum != h2.checksum {
		t.Errorf("checksum mismatch: %x != %x", h1.checksum, h2.checksum)
	}
	diff := h1.Diff(h2)
	if diff != 0 {
		t.Errorf("hashes differ by: %d", diff)
	}
}
