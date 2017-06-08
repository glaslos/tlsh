package tlsh

import (
	"fmt"
	"io/ioutil"
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
		{"tests/test_file_5", "E1D1B7337E4E03044FE22379D7C9C95ED66CE42426C39759CCEA9A2AF516838E723364"},
		{"tests/test_file_6", "2FE1A7723E8603145BF222F9979ACC7EF74CE4242BD3A7D49899F919F146814C3233A8"},
		{"tests/test_file_7_lena.jpg", "85C2F1CE3D989428683106EBE5EAAAC924F2D5020B38B1550DA8E5F0DD8C65DECF7037"},
		{"tests/test_file_8_lena.png", "F7A433B5648BCC69DD48E1DDF1A1876C56E08C0BB264438FAB412C4686FA3F3DB05E36"},
		{"tests/test_file_9_tinyssl.exe", "67A3AD97F601C873E11A0AF49D83D2D6BC7F7F709E522C9B74990B0E8D796822D1D48A"},
	}
)

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

func BenchmarkPearson(b *testing.B) {
	var salt = byte(0)
	var keys = []byte{1, 3, 7}
	for n := 0; n < b.N; n++ {
		pearsonHash(salt, keys)
	}
}

func BenchmarkFillBuckets(b *testing.B) {
	data, err := ioutil.ReadFile("tests/test_file_1")
	if err != nil {
		b.Error(err)
	}
	for n := 0; n < b.N; n++ {
		fillBuckets(data)
	}
}

func BenchmarkQuartilePoints(b *testing.B) {
	data, err := ioutil.ReadFile("tests/test_file_1")
	if err != nil {
		b.Error(err)
	}
	buckets, _ := fillBuckets(data)
	for n := 0; n < b.N; n++ {
		quartilePoints(buckets)
	}
}
