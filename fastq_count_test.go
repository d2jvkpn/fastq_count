package main

import (
	"fmt"
	"testing"
)

//go:generate go test -run TestFastqCount_t1
func TestFastqCount_t1(t *testing.T) {
	ct := NewCounter(33)
	var expected int64 = 37500

	go ReadBlocks("examples/Sample.fastq", ct, nil)
	ct.Counting()
	fmt.Printf("%#v\n", ct)

	if ct.BN != expected {
		t.Fatalf("Bases number not equals to %d\n", expected)
	}
}
