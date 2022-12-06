package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

type Device struct {
	counter uint
	stream  io.Reader
	sop     RingBuffer
}

const SOP_LENGTH = 4

func NewDevice(stream io.Reader) (result *Device) {
	result = new(Device)
	result.stream = stream
	result.sop = MakeRingBuffer(SOP_LENGTH)
	return
}

func (device *Device) ReadChar() (result byte, err error) {
	out := [1]byte{}
	n, err := device.stream.Read(out[:1])
	if err != nil {
		return
	}
	if n != 1 {
		err = errors.New("Unexpected read count")
		return
	}
	result = out[0]
	device.sop.Push(result)
	device.counter++
	return
}

func (device *Device) Pos() uint {
	return device.counter
}

func (device *Device) IsSOP() bool {
	if device.sop.Length() != SOP_LENGTH {
		return false
	}

	var seen [SOP_LENGTH]byte
	for i := 0; i < SOP_LENGTH; i++ {
		c, err := device.sop.Get(i)
		if err != nil {
			panic(err)
		}
		seen[i] = c
		for j := 0; j < i; j++ {
			if seen[j] == c {
				return false
			}
		}
	}

	return true
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	device := NewDevice(reader)
	for {
		_, err := device.ReadChar()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		if device.IsSOP() {
			fmt.Println(device.Pos())
			break
		}
	}
}
