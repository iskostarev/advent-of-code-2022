package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type MonkeyValue int
type Monkey interface {
	Calculate(group *MonkeyGroup, depth int) MonkeyValue
}

type MonkeyGroup map[string]Monkey

type MonkeyNumber struct {
	Value MonkeyValue
}

type MonkeyOp struct {
	Operation func(MonkeyValue, MonkeyValue) MonkeyValue
	Lhs, Rhs  string
	cached    bool
	cachedVal MonkeyValue
}

type Parser struct {
	lineNum *regexp.Regexp
	lineOp  *regexp.Regexp
}

func debugPrint(depth int, str string) {
	// line := strings.Repeat(" ", depth)
	// line += str
	// fmt.Println(line)
}

func (group *MonkeyGroup) Get(id string) (result Monkey) {
	result, ok := (*group)[id]
	if !ok {
		panic("Invalid monkey ID")
	}
	return
}

func (m *MonkeyNumber) Calculate(group *MonkeyGroup, depth int) MonkeyValue {
	debugPrint(depth, fmt.Sprintf("value: %v", m.Value))
	return m.Value
}

func (m *MonkeyOp) Calculate(group *MonkeyGroup, depth int) MonkeyValue {
	if m.cached {
		debugPrint(depth, fmt.Sprintf("cached: %v", m.cachedVal))
		return m.cachedVal
	}

	debugPrint(depth, "operation...")
	lhs := group.Get(m.Lhs).Calculate(group, depth+1)
	rhs := group.Get(m.Rhs).Calculate(group, depth+1)
	m.cachedVal = m.Operation(lhs, rhs)
	m.cached = true
	debugPrint(depth, fmt.Sprintf("operation got: %v", m.cachedVal))
	return m.cachedVal
}

func MakeParser() (result Parser) {
	result.lineNum = regexp.MustCompile(`^([a-z]{4}):\s*(\d+)\s*$`)
	result.lineOp = regexp.MustCompile(`^([a-z]{4}):\s*([a-z]{4})\s*([+-/*])\s*([a-z]{4})\s*$`)
	return
}

func ParseInt(s string) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return result
}

func ParseOp(op string) func(MonkeyValue, MonkeyValue) MonkeyValue {
	switch op {
	case "+":
		return func(lhs, rhs MonkeyValue) MonkeyValue {
			return lhs + rhs
		}
	case "-":
		return func(lhs, rhs MonkeyValue) MonkeyValue {
			return lhs - rhs
		}
	case "*":
		return func(lhs, rhs MonkeyValue) MonkeyValue {
			return lhs * rhs
		}
	case "/":
		return func(lhs, rhs MonkeyValue) MonkeyValue {
			return lhs / rhs
		}
	}
	panic("Invalid operation")
}

func (parser *Parser) ParseLine(line string) (id string, monkey Monkey) {
	matches := parser.lineNum.FindStringSubmatch(line)
	if len(matches) == 3 {
		id = matches[1]
		num := MonkeyValue(ParseInt(matches[2]))
		monkey = &MonkeyNumber{Value: num}
		return
	}

	matches = parser.lineOp.FindStringSubmatch(line)
	if len(matches) == 5 {
		id = matches[1]
		monkey = &MonkeyOp{
			Lhs:       matches[2],
			Rhs:       matches[4],
			Operation: ParseOp(matches[3]),
		}
		return
	}

	panic("Invalid input")
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	parser := MakeParser()
	group := MonkeyGroup{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		id, monkey := parser.ParseLine(line)
		group[id] = monkey
	}
	rootVal := group.Get("root").Calculate(&group, 0)
	fmt.Println(rootVal)
}
