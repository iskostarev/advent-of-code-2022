package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Assignment struct {
	SectionMin, SectionMax int
}

type Parser struct {
	regex *regexp.Regexp
}

func MakeParser() (result Parser) {
	result.regex = regexp.MustCompile(`^(\d+)-(\d+),(\d+)-(\d+)$`)
	return
}

func (parser *Parser) ParseLine(line string) (first, second Assignment) {
	onFail := func() {
		panic(fmt.Sprintf("Failed to parse line: %s", line))
	}

	matches := parser.regex.FindStringSubmatch(line)
	if len(matches) != 5 {
		onFail()
	}

	parseInt := func(str string) (res int) {
		res, err := strconv.Atoi(str)
		if err != nil {
			onFail()
		}
		return
	}

	first.SectionMin = parseInt(matches[1])
	first.SectionMax = parseInt(matches[2])
	second.SectionMin = parseInt(matches[3])
	second.SectionMax = parseInt(matches[4])

	return
}

func (lhs *Assignment) Includes(rhs Assignment) bool {
	return lhs.SectionMin <= rhs.SectionMin && lhs.SectionMax >= rhs.SectionMax
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	parser := MakeParser()
	counter := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		ass1, ass2 := parser.ParseLine(line)
		if ass1.Includes(ass2) || ass2.Includes(ass1) {
			counter++
		}
	}

	fmt.Println(counter)
}
