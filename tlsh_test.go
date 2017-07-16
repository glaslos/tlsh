package tlsh

import (
	"bufio"
	"os"
	"testing"
)

var (
	hashTestCases = []struct {
		filename string
		hash     string
	}{
		{"tests/test_file_1", "8ed02202fc30802303a002b03b33300fc30a82f83008c2fa000a0080b8ba0e02cca0c3"},
		{"tests/test_file_2", "b2319634f5c033244eb792aa3168a366e737553da305a28440ce842d7b57a2cc63b6ec"},
		{"tests/test_file_3", "ea31834386c503b62a920319ba4f92d3bf6fc2b863384515a4ea5638450bc1e9376ae9"},
		{"tests/test_file_4", "5111421e72610b73189a13a055b8a8d9b22bb25b7aaf2a84146df245232a06cd5fb854"},
		{"tests/test_file_5", "e1d1b7337e4e03044fe22379d7c9c95ed66ce42426c39759ccea9a2af516838e723364"},
		{"tests/test_file_6", "2fe1a7723e8603145bf222f9979acc7ef74ce4242bd3a7d49899f919f146814c3233a8"},
		{"tests/test_file_7_lena.jpg", "85c2f1ce3d989428683106ebe5eaaac924f2d5020b38b1550da8e5f0dd8c65decf7037"},
		{"tests/test_file_8_lena.png", "f7a433b5648bcc69dd48e1ddf1a1876c56e08c0bb264438fab412c4686fa3f3db05e36"},
		{"tests/test_file_9_tinyssl.exe", "67a3ad97f601c873e11a0af49d83d2d6bc7f7f709e522c9b74990b0e8d796822d1d48a"},
	}

	diffTestCases = []struct {
		filenameA string
		filenameB string
		diff      int
	}{
		{"tests/test_file_1", "tests/test_file_2", 418},
		{"tests/test_file_1", "tests/test_file_8_lena.png", 1014},
		{"tests/test_file_3", "tests/test_file_1", 374},
		{"tests/test_file_3", "tests/test_file_8_lena.png", 967},
		{"tests/test_file_7_lena.jpg", "tests/test_file_8_lena.png", 619},
	}
)

func TestHash(t *testing.T) {
	for _, tc := range hashTestCases {
		if hash, err := Hash(tc.filename); hash != tc.hash {
			if err != nil {
				t.Error(err)
			}
			t.Errorf("\nfilename: %s\n%s\n%s - doesn't match real hash\n", tc.filename, tc.hash, hash)
		}
	}
}

func TestDiff(t *testing.T) {
	for _, tc := range diffTestCases {
		if diff, err := Diff(tc.filenameA, tc.filenameB); diff != tc.diff {
			if err != nil {
				t.Error(err)
			}
			t.Errorf("\nfilename: %s and %s have wrong distance %d vs. %d\n", tc.filenameA, tc.filenameB, tc.diff, diff)
		}
	}
}

func BenchmarkPearson(b *testing.B) {
	var salt = byte(0)
	var keys = [3]byte{1, 3, 7}
	for n := 0; n < b.N; n++ {
		pearsonHash(salt, &keys)
	}
}

func BenchmarkFillBuckets(b *testing.B) {
	f, err := os.Open("tests/test_file_1")
	defer f.Close()
	if err != nil {
		b.Error(err)
	}
	f.Seek(0, 0)
	for n := 0; n < b.N; n++ {
		r := bufio.NewReader(f)
		fillBuckets(r)
		f.Seek(0, 0)
	}
}

func BenchmarkQuartilePoints(b *testing.B) {
	f, err := os.Open("tests/test_file_1")
	defer f.Close()
	if err != nil {
		b.Error(err)
	}
	r := bufio.NewReader(f)
	buckets, _, _, err := fillBuckets(r)
	if err != nil {
		b.Error(err)
	}
	for n := 0; n < b.N; n++ {
		quartilePoints(buckets)
	}
}

func BenchmarkHash(b *testing.B) {
	f := "tests/test_file_1"
	for n := 0; n < b.N; n++ {
		Hash(f)
	}
}
