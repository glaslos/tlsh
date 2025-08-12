package tlsh

import (
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
	t.Logf("%x", hash[:])
}

func TestHashWrite(t *testing.T) {
	// hash using the hash.Hash interface methods
	h1 := New()
	h1.Write([]byte("1234"))
	h1.Write([]byte("11"))
	h1.Write([]byte("1111111"))
	t.Logf("h1: %x", h1.Sum(nil))
	t.Logf("checksum h1: %d, %x", h1.state.checksum, h1.checksum)

	// hash from read
	h2, err := HashBytes([]byte("1234111111111"))
	if err == nil {
		t.Error("Missing error of less than 50 bytes")
	}
	t.Logf("checksum h2: %d, %x", h2.state.checksum, h2.checksum)
	t.Logf("h2: %x", h2.Binary())

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

	h1.Write([]byte("1234567890"))
	h1.Write([]byte("1234567890"))
	h1.Write([]byte("1234567890"))
	h1.Write([]byte("1234567890"))
	t.Logf("h1: %x", h1.Sum(nil))
	t.Logf("checksum h1: %d, %x", h1.state.checksum, h1.checksum)
	s := "1234111111111" + "1234567890" + "1234567890" + "1234567890" + "1234567890"
	h3, err := HashBytes([]byte(s))
	if err != nil {
		t.Error(err)
	}
	t.Logf("checksum h3: %d, %x", h3.state.checksum, h3.checksum)
	t.Logf("h3: %x", h3.Binary())
	if h1.state.fileSize != h3.state.fileSize {
		t.Errorf("file size mismatch: %d != %d", h1.state.fileSize, h3.state.fileSize)
	}
	if h1.checksum != h3.checksum {
		t.Errorf("checksum mismatch: %x != %x", h1.checksum, h3.checksum)
	}
	diff = h1.Diff(h3)
	if diff != 0 {
		t.Errorf("hashes differ by: %d", diff)
	}
}
