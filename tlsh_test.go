package tlsh

import (
	"fmt"
	"testing"
)

var (
	testCases = []struct {
		filename string
		hash     string
	}{
		{"tests/test_file_1", "8ED02202FC30802303A002B03B33300FC30A82F83008C2FA000A0080B8BA0E02CCA0C3"},
	}
)

func TestSomething(t *testing.T) {
}

func TestReal(t *testing.T) {
	for _, tc := range testCases {
		if bar, err := Hash(tc.filename); bar != tc.hash {
			fmt.Printf("%s\n", bar)
			if err != nil {
				t.Error(err)
			}
			t.Errorf("\nfilename: %s\n%s\n%s - doesn't match real hash\n", tc.filename, tc.hash, bar)
		}
	}
}
