package main

import (
	"errors"
)

type RingBuffer struct {
	size   int
	head   int
	length int
	buffer []byte
}

func MakeRingBuffer(size int) (result RingBuffer) {
	result.size = size
	result.buffer = make([]byte, size)
	return
}

func (rb *RingBuffer) Push(c byte) {
	var index int
	if rb.length < rb.size {
		index = (rb.head + rb.length) % rb.size
		rb.length++
	} else {
		index = rb.head
		rb.head = (rb.head + 1) % rb.size
	}
	rb.buffer[index] = c
}

func (rb *RingBuffer) Pop() (result byte, err error) {
	if rb.length == 0 {
		err = errors.New("Popping from empty ring buffer")
		return
	}
	result = rb.buffer[rb.head]
	rb.head = (rb.head + 1) % rb.size
	rb.length--
	return
}

func (rb *RingBuffer) Length() int {
	return rb.length
}

func (rb *RingBuffer) Get(index int) (result byte, err error) {
	if index >= rb.length {
		err = errors.New("Index out of bounds")
		return
	}

	result = rb.buffer[(rb.head+index)%rb.size]
	return
}
