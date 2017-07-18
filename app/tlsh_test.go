package main

import "testing"

func TestMainVersion(t *testing.T) {
	version = true
	Main()
}

func TestMainRaw(t *testing.T) {
	version = false
	file = "../tests/test_file_1"
	raw = true
	Main()
}

func TestMainHash(t *testing.T) {
	version = false
	file = "../tests/test_file_1"
	raw = false
	Main()
}

func TestMainHashError(t *testing.T) {
	version = false
	file = "../tests/test_file_666"
	raw = false
	Main()
}

func TestMainCompare(t *testing.T) {
	version = false
	file = "../tests/test_file_1"
	compare = "../tests/test_file_2"
	raw = false
	Main()
}

func TestMainHelp(t *testing.T) {
	version = false
	file = ""
	Main()
}
