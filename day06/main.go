package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Device struct {
	counter uint
	stream  io.Reader
	window  RingBuffer
}

func NewDevice(stream io.Reader, windowsize int) (result *Device) {
	result = new(Device)
	result.stream = stream
	result.window = MakeRingBuffer(windowsize)
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
	device.window.Push(result)
	device.counter++
	return
}

func (device *Device) Pos() uint {
	return device.counter
}

func (device *Device) WinSize() int {
	return device.window.Size()
}

func (device *Device) IsWindowUniq() bool {
	if device.window.Length() != device.WinSize() {
		return false
	}

	seen := make([]byte, device.WinSize())
	for i := 0; i < device.WinSize(); i++ {
		c, err := device.window.Get(i)
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
	winsize := 4

	if len(os.Args) > 1 {
		var err error
		winsize, err = strconv.Atoi(os.Args[1])
		if err != nil {
			panic(err)
		}
	}

	reader := bufio.NewReader(os.Stdin)
	device := NewDevice(reader, winsize)
	for {
		_, err := device.ReadChar()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		if device.IsWindowUniq() {
			fmt.Println(device.Pos())
			break
		}
	}
}
