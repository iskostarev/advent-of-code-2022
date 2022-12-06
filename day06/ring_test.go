package main

import (
	"testing"
)

func AssertGet(t *testing.T, ring *RingBuffer, index int, expected byte) {
	c, err := ring.Get(index)
	if err != nil {
		t.Fatalf("AssertGet(t, %v, %d, '%c'): err %v", ring, index, expected, err)
	}
	if c != expected {
		t.Fatalf("AssertGet(t, %v, %d, '%c'): unexpected '%c'", ring, index, expected, c)
	}
}

func AssertGetErr(t *testing.T, ring *RingBuffer, index int, expected string) {
	_, err := ring.Get(index)
	if err == nil {
		t.Fatalf("AssertGetErr(t, %v, %d, \"%s\"): no error", ring, index, expected)
	}
	if err.Error() != expected {
		t.Fatalf("AssertGetErr(t, %v, %d, \"%s\"): got %v instead", ring, index, expected, err)
	}
}

func AssertPop(t *testing.T, ring *RingBuffer, expected byte) {
	c, err := ring.Pop()
	if err != nil {
		t.Fatalf("AssertPop(t, %v, '%c'): err %v", ring, expected, err)
	}
	if c != expected {
		t.Fatalf("AssertPop(t, %v, '%c'): unexpected '%c'", ring, expected, c)
	}
}

func AssertPopErr(t *testing.T, ring *RingBuffer, expected string) {
	_, err := ring.Pop()
	if err == nil {
		t.Fatalf("AssertPopErr(t, %v, \"%s\"): no error", ring, expected)
	}
	if err.Error() != expected {
		t.Fatalf("AssertPopErr(t, %v, \"%s\"): got %v instead", ring, expected, err)
	}
}

func TestBasic(t *testing.T) {
	ring := MakeRingBuffer(4)
	ring.Push('A')
	ring.Push('B')
	AssertGet(t, &ring, 0, 'A')
	AssertGet(t, &ring, 1, 'B')
	AssertGetErr(t, &ring, 2, "Index out of bounds")
	AssertPop(t, &ring, 'A')
	AssertPop(t, &ring, 'B')
	AssertPopErr(t, &ring, "Popping from empty ring buffer")
	ring.Push('1')
	ring.Push('2')
	ring.Push('3')
	ring.Push('4')
	ring.Push('5')
	AssertGet(t, &ring, 0, '2')
	AssertGet(t, &ring, 1, '3')
	AssertGet(t, &ring, 2, '4')
	AssertGet(t, &ring, 3, '5')
	AssertGetErr(t, &ring, 4, "Index out of bounds")
}
