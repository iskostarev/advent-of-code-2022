package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ParseSnafuDigit(c byte) int {
	switch c {
	case '2':
		return 2
	case '1':
		return 1
	case '0':
		return 0
	case '-':
		return -1
	case '=':
		return -2
	}
	panic("Unexpected digit")
}

func ParseSnafu(str string) (result int) {
	n := 1
	for i := len(str) - 1; i >= 0; i-- {
		result += n * ParseSnafuDigit(str[i])
		n *= 5
	}
	return
}

func IntToSnafu(num int) (result string) {
	for {
		rem := num % 5
		if rem <= 2 {
			result = fmt.Sprintf("%d", rem) + result
		} else {
			rem = rem - 5
			num -= rem
			switch rem {
			case -1:
				result = "-" + result
			case -2:
				result = "=" + result
			default:
				panic("Impossible")
			}
		}

		num /= 5
		if num == 0 {
			return result
		}
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	sum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		sum += ParseSnafu(line)
	}

	fmt.Println(IntToSnafu(sum))
}
