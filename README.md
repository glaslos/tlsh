![Workflow](https://github.com/glaslos/ssdeep/actions/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/glaslos/tlsh)](https://goreportcard.com/report/github.com/glaslos/tlsh)
[![Go Reference](https://pkg.go.dev/badge/badge/glaslos/tlsh.svg)](https://pkg.go.dev/github.com/glaslos/tlsh)

# TLSH
Trend Micro Locality Sensitive Hash lib in Golang

Based on https://github.com/trendmicro/tlsh

See paper here: https://github.com/trendmicro/tlsh/blob/master/TLSH_CTC_final.pdf

TLSH is a fuzzy matching library. Given a byte stream with a minimum length of 256 bytes, TLSH generates a hash value which can be used for similarity comparisons. Similar objects will have similar hash values which allows for the detection of similar objects by comparing their hash values. Note that the byte stream should have a sufficient amount of complexity. For example, a byte stream of identical bytes will not generate a hash value.

The computed hash is 35 bytes long (output as 70 hexidecimal charactes). The first 3 bytes are used to capture the information about the file as a whole (length, ...), while the last 32 bytes are used to capture information about incremental parts of the file.
