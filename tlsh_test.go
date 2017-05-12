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
		{"tests/test_file_2", "B2319634F5C033244EB792AA3168A366E737553DA305A28440CE842D7B57A2CC63B6EC"},
		{"tests/test_file_3", "EA31834386C503B62A920319BA4F92D3BF6FC2B863384515A4EA5638450BC1E9376AE9"},
		{"tests/test_file_4", "5111421E72610B73189A13A055B8A8D9B22BB25B7AAF2A84146DF245232A06CD5FB854"},
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
