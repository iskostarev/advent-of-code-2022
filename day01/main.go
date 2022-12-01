package main

import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"strconv"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	var max uint64 = 0
	var cur uint64 = 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if cur > max {
				max = cur
			}
			cur = 0
		} else {
			calories, err := strconv.Atoi(line)
			if err != nil {
				panic(err)
			}
			cur += uint64(calories)
		}
	}

	if cur > max {
		max = cur
	}

	fmt.Println(max)
}
