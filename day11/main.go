package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type WorryLevel int
type MonkeyOp func(WorryLevel) WorryLevel
type MonkeyTest func(WorryLevel) int
type Monkey struct {
	Items        []WorryLevel
	Operation    MonkeyOp
	Test         MonkeyTest
	InspectCount int
}

type MonkeyGroup struct {
	monkeys   map[int]*Monkey
	maxMonkey int
}

type MonkeyParser struct {
	header        *regexp.Regexp
	startingItems *regexp.Regexp
	operation     *regexp.Regexp
	test          *regexp.Regexp
	ifTestTrue    *regexp.Regexp
	ifTestFalse   *regexp.Regexp

	opBinary    *regexp.Regexp
	testDivBy   *regexp.Regexp
	throwAction *regexp.Regexp
}

type TopSelector struct {
	first, second int
}

func (monkey *Monkey) AppendItem(item WorryLevel) {
	monkey.Items = append(monkey.Items, item)
}

func MakeMonkeyGroup() (result MonkeyGroup) {
	result.monkeys = make(map[int]*Monkey)
	return
}

func (group *MonkeyGroup) AddMonkey(num int, monkey *Monkey) {
	_, exists := group.monkeys[num]
	if exists {
		panic("Monkey already exists")
	}
	group.monkeys[num] = monkey
	if group.maxMonkey < num {
		group.maxMonkey = num
	}
}

func (group *MonkeyGroup) Monkey(i int) (result *Monkey) {
	result, ok := group.monkeys[i]
	if !ok {
		panic("Missing monkey")
	}
	return
}

func (group *MonkeyGroup) NumMonkeys() int {
	return len(group.monkeys)
}

func MakeMonkeyParser() (result MonkeyParser) {
	result.header = regexp.MustCompile(`^Monkey (\d+):$`)
	result.startingItems = regexp.MustCompile(`^  Starting items:\s*(\d+(?:,\s*\d+)*)$`)
	result.operation = regexp.MustCompile(`^  Operation:\s*new\s*=\s*(.*)$`)
	result.test = regexp.MustCompile(`^  Test:\s*(.*)$`)
	result.ifTestTrue = regexp.MustCompile(`^    If true:\s*(.*)$`)
	result.ifTestFalse = regexp.MustCompile(`^    If false:\s(.*)$`)

	result.opBinary = regexp.MustCompile(`^((?:old)|(?:\d+))\s*([+*])\s*((?:old)|(?:\d+))$`)
	result.testDivBy = regexp.MustCompile(`^divisible by (\d+)$`)
	result.throwAction = regexp.MustCompile(`^throw to monkey (\d+)$`)
	return
}

func (parser *MonkeyParser) ParseStartingItems(str string) []WorryLevel {
	result := []WorryLevel{}

	for _, sub := range strings.Split(str, ",") {
		sub = strings.TrimSpace(sub)
		item, err := strconv.Atoi(sub)
		if err != nil {
			panic("ParseStartingItems: integer expected")
		}
		result = append(result, WorryLevel(item))
	}

	return result
}

func (parser *MonkeyParser) ParseOperation(str string) MonkeyOp {
	matches := parser.opBinary.FindStringSubmatch(str)
	if len(matches) != 4 {
		panic("Invalid operation")
	}

	parseArg := func(s string) (useOld bool, literal WorryLevel) {
		if s == "old" {
			useOld = true
			return
		}

		useOld = false
		val, err := strconv.Atoi(s)
		if err != nil {
			panic("Invalid operation argument")
		}

		literal = WorryLevel(val)

		return
	}

	lhsOld, lhsLit := parseArg(matches[1])
	rhsOld, rhsLit := parseArg(matches[3])
	var op func(WorryLevel, WorryLevel) WorryLevel
	switch matches[2] {
	case "+":
		op = func(lhs, rhs WorryLevel) WorryLevel { return lhs + rhs }
	case "*":
		op = func(lhs, rhs WorryLevel) WorryLevel { return lhs * rhs }
	default:
		panic("Invalid operation sign")
	}

	return func(old WorryLevel) WorryLevel {
		var lhs, rhs WorryLevel
		if lhsOld {
			lhs = old
		} else {
			lhs = lhsLit
		}

		if rhsOld {
			rhs = old
		} else {
			rhs = rhsLit
		}

		return op(lhs, rhs)
	}
}

func (parser *MonkeyParser) ParseThrowAction(str string) (result int) {
	matches := parser.throwAction.FindStringSubmatch(str)
	if len(matches) != 2 {
		panic("Invalid throw action")
	}

	var err error
	result, err = strconv.Atoi(matches[1])
	if err != nil {
		panic("Invalid throw target")
	}

	return
}

func (parser *MonkeyParser) ParseTest(test, ifTrue, ifFalse string) MonkeyTest {
	matches := parser.testDivBy.FindStringSubmatch(test)
	if len(matches) != 2 {
		panic("Invalid test")
	}

	divBy, err := strconv.Atoi(matches[1])
	if err != nil {
		panic("Invalid test expression")
	}

	trueTarget := parser.ParseThrowAction(ifTrue)
	falseTarget := parser.ParseThrowAction(ifFalse)

	return func(level WorryLevel) int {
		if int(level)%divBy == 0 {
			return trueTarget
		} else {
			return falseTarget
		}
	}
}

func (parser *MonkeyParser) ParseMonkey(scanner *bufio.Scanner) (success bool, num int, result Monkey) {
	parseLine := func(line string, regex *regexp.Regexp) (bool, []string) {
		matches := regex.FindStringSubmatch(line)
		if len(matches) != regex.NumSubexp()+1 {
			return false, nil
		}
		return true, matches
	}

	getLine := func(regex *regexp.Regexp) (bool, []string) {
		if !scanner.Scan() {
			return false, nil
		}
		return parseLine(scanner.Text(), regex)
	}

	var ok bool
	var matches []string
	var err error

	for {
		if !scanner.Scan() {
			success = false
			return
		}
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		ok, matches = parseLine(line, parser.header)
		if !ok {
			panic("Invalid header")
		}
		num, err = strconv.Atoi(matches[1])
		if err != nil {
			panic("Invalid monkey number")
		}
		break
	}

	ok, matches = getLine(parser.startingItems)
	if !ok {
		panic("Invalid starting items")
	}
	result.Items = parser.ParseStartingItems(matches[1])

	ok, matches = getLine(parser.operation)
	if !ok {
		panic("Invalid operation")
	}
	result.Operation = parser.ParseOperation(matches[1])

	ok, matches = getLine(parser.test)
	if !ok {
		panic("Invalid test line")
	}
	test := matches[1]

	ok, matches = getLine(parser.ifTestTrue)
	if !ok {
		panic("Invalid if_test_true expression")
	}
	ifTestTrue := matches[1]

	ok, matches = getLine(parser.ifTestFalse)
	if !ok {
		panic("Invalid if_test_false expression")
	}
	ifTestFalse := matches[1]

	result.Test = parser.ParseTest(test, ifTestTrue, ifTestFalse)

	success = true
	return
}

func ParseMonkeyGroup(scanner *bufio.Scanner) (result MonkeyGroup) {
	parser := MakeMonkeyParser()
	result = MakeMonkeyGroup()

	for {
		success, num, monkey := parser.ParseMonkey(scanner)
		if !success {
			break
		}
		result.AddMonkey(num, &monkey)
	}

	if scanner.Scan() {
		panic("Extra lines at the end")
	}

	return
}

func Turn(group *MonkeyGroup, cur int) {
	monkey := group.Monkey(cur)
	for _, item := range monkey.Items {
		wl := monkey.Operation(item)
		wl /= 3
		target := monkey.Test(wl)
		if target == cur {
			panic("Can't throw an item to itself")
		}
		//fmt.Printf("%d: %d -> %d, target=%d\n", cur, item, wl, target)
		group.Monkey(target).AppendItem(wl)
		monkey.InspectCount++
	}
	monkey.Items = monkey.Items[:0]
}

func Round(group *MonkeyGroup) {
	for i := 0; i < group.NumMonkeys(); i++ {
		Turn(group, i)
	}
}

func (selector *TopSelector) Insert(val int) {
	if val >= selector.first {
		selector.second = selector.first
		selector.first = val
	} else if val >= selector.second {
		selector.second = val
	}
}

func (selector *TopSelector) MonkeyBusiness() int {
	return selector.first * selector.second
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	monkeyGroup := ParseMonkeyGroup(scanner)

	for i := 0; i < 20; i++ {
		Round(&monkeyGroup)
	}

	selector := TopSelector{}
	for i := 0; i < monkeyGroup.NumMonkeys(); i++ {
		selector.Insert(monkeyGroup.Monkey(i).InspectCount)
	}

	fmt.Println(selector.MonkeyBusiness())
}
