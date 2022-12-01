package main

import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"strconv"
)

type Selector struct {
	size int
	topvals []uint64 // sorted from min to max
}

func (selector *Selector) findPivot(val uint64) int {
	for i := selector.size-1; i >= 0; i-- {
		if selector.topvals[i] < val {
			return i
		}
	}

	panic("no pivot was found")
}

func (selector *Selector) Init(size int) {
	selector.size = size
	selector.topvals = make([]uint64, size)
}

func (selector *Selector) Insert(val uint64) {
	if selector.topvals[0] >= val {
		return
	}

	pivot := selector.findPivot(val)

	for i := 0; i < pivot; i++ {
		selector.topvals[i] = selector.topvals[i+1]
	}
	selector.topvals[pivot] = val
}

func (selector *Selector) Select() (result uint64) {
	for _, val := range(selector.topvals) {
		result += val
	}
	return
}

func (selector *Selector) DebugPrint() () {
	line := ""
	for _, val := range(selector.topvals) {
		line += fmt.Sprintf("%d;", val)
	}
	fmt.Println(line)
}

func main() {
	size := 1

	if len(os.Args) > 1 {
		var err error
		size, err = strconv.Atoi(os.Args[1])
		if err != nil {
			panic(err)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)

	var selector Selector
	selector.Init(size)

	var cur uint64 = 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			selector.Insert(cur)
			cur = 0
		} else {
			calories, err := strconv.Atoi(line)
			if err != nil {
				panic(err)
			}
			cur += uint64(calories)
		}
	}

	selector.Insert(cur)
	fmt.Println(selector.Select())
}
