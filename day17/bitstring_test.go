package main

import (
	"testing"
)

func AssertBitsSet(t *testing.T, bs Bitstring, bits []int) {
	for i := 0; i < bs.Size(); i++ {
		mustBeSet := false
		for _, b := range bits {
			if b == i {
				mustBeSet = true
				break
			}
		}

		if bs.At(i) != mustBeSet {
			t.Fatalf("Bit %d must be set to %v: %v", i, mustBeSet, bs)
		}
	}
}

func TestBasic(t *testing.T) {
	bs := MakeBitstring(200)

	if bs.Size() != 200 {
		t.Fatalf("Invalid size")
	}

	AssertBitsSet(t, bs, []int{})

	targets := []int{10, 20, 30, 60, 64, 65, 100, 127, 128, 190}
	for _, pos := range targets {
		bs.Set(pos, true)
	}
	AssertBitsSet(t, bs, targets)

	bs.Set(60, false)
	bs.Set(64, false)
	bs.Set(190, false)
	bs.Set(195, false)

	AssertBitsSet(t, bs, []int{10, 20, 30, 65, 100, 127, 128})
}
