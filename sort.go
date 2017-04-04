package tlsh

import (
	"sort"
)

type sortByteArray []byte

func (b sortByteArray) Len() int {
	return len(b)
}

func (b sortByteArray) Less(i, j int) bool {
	if b[i] < b[j] {
		return true
	}
	return false
}

func (b sortByteArray) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// SortByteArray sorts an array of bytes
func SortByteArray(src []byte) []byte {
	sorted := sortByteArray(src)
	sort.Sort(sorted)
	return sorted
}
