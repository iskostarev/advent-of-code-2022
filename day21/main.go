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
	Calculate(group *MonkeyGroup, depth int) (bool, MonkeyValue)
	Solve(target MonkeyValue, group *MonkeyGroup, depth int) MonkeyValue
	SolveRoot(group *MonkeyGroup) MonkeyValue
}

type MonkeyGroup map[string]Monkey

type MonkeyNumber struct {
	Value MonkeyValue
	Human bool
}

type BinaryOp func(MonkeyValue, MonkeyValue) MonkeyValue

type ExtendedBinaryOp struct {
	Op, SolveLhs, SolveRhs BinaryOp
}

type MonkeyOp struct {
	Operation  ExtendedBinaryOp
	Lhs, Rhs   string
	cached     bool
	cachedVal  MonkeyValue
	cachedHumn bool
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

func (m *MonkeyNumber) Calculate(group *MonkeyGroup, depth int) (bool, MonkeyValue) {
	debugPrint(depth, fmt.Sprintf("value: %v", m.Value))
	return m.Human, m.Value
}

func (m *MonkeyOp) Calculate(group *MonkeyGroup, depth int) (bool, MonkeyValue) {
	if m.cached {
		debugPrint(depth, fmt.Sprintf("cached: %v", m.cachedVal))
		return m.cachedHumn, m.cachedVal
	}

	debugPrint(depth, "operation...")
	lhsHumn, lhs := group.Get(m.Lhs).Calculate(group, depth+1)
	rhsHumn, rhs := group.Get(m.Rhs).Calculate(group, depth+1)
	m.cachedVal = m.Operation.Op(lhs, rhs)
	m.cachedHumn = lhsHumn || rhsHumn
	m.cached = true
	debugPrint(depth, fmt.Sprintf("operation got: %v", m.cachedVal))
	return m.cachedHumn, m.cachedVal
}

func (m *MonkeyNumber) Solve(target MonkeyValue, group *MonkeyGroup, depth int) MonkeyValue {
	if m.Human {
		return target
	}
	panic("Can't solve a constant")
}

func (m *MonkeyOp) Solve(target MonkeyValue, group *MonkeyGroup, depth int) MonkeyValue {
	lhsHumn, lhs := group.Get(m.Lhs).Calculate(group, depth+1)
	rhsHumn, rhs := group.Get(m.Rhs).Calculate(group, depth+1)

	if lhsHumn && rhsHumn {
		panic("Both subtrees can't contain human")
	}

	if lhsHumn {
		target = m.Operation.SolveLhs(target, rhs)
		return group.Get(m.Lhs).Solve(target, group, depth+1)
	} else if rhsHumn {
		target = m.Operation.SolveRhs(target, lhs)
		return group.Get(m.Rhs).Solve(target, group, depth+1)
	}

	panic("No human in tree")
}

func (m *MonkeyNumber) SolveRoot(group *MonkeyGroup) MonkeyValue {
	panic("Can't solve a constant")
}

func (m *MonkeyOp) SolveRoot(group *MonkeyGroup) MonkeyValue {
	lhsHumn, lhs := group.Get(m.Lhs).Calculate(group, 0)
	rhsHumn, rhs := group.Get(m.Rhs).Calculate(group, 0)

	if lhsHumn && rhsHumn {
		panic("Both subtrees can't contain human")
	}

	if lhsHumn {
		return group.Get(m.Lhs).Solve(rhs, group, 0)
	} else if rhsHumn {
		return group.Get(m.Rhs).Solve(lhs, group, 0)
	}

	panic("No human in tree")
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

func ParseOp(symbol string) ExtendedBinaryOp {
	var op, solveLhs, solveRhs BinaryOp
	switch symbol {
	case "+":
		op = func(lhs, rhs MonkeyValue) MonkeyValue {
			return lhs + rhs
		}
		solveLhs = func(rslt, rhs MonkeyValue) MonkeyValue {
			return rslt - rhs
		}
		solveRhs = func(rslt, lhs MonkeyValue) MonkeyValue {
			return rslt - lhs
		}
	case "-":
		op = func(lhs, rhs MonkeyValue) MonkeyValue {
			return lhs - rhs
		}
		solveLhs = func(rslt, rhs MonkeyValue) MonkeyValue {
			return rslt + rhs
		}
		solveRhs = func(rslt, lhs MonkeyValue) MonkeyValue {
			return lhs - rslt
		}
	case "*":
		op = func(lhs, rhs MonkeyValue) MonkeyValue {
			return lhs * rhs
		}
		solveLhs = func(rslt, rhs MonkeyValue) MonkeyValue {
			return rslt / rhs
		}
		solveRhs = func(rslt, lhs MonkeyValue) MonkeyValue {
			return rslt / lhs
		}
	case "/":
		op = func(lhs, rhs MonkeyValue) MonkeyValue {
			return lhs / rhs
		}
		solveLhs = func(rslt, rhs MonkeyValue) MonkeyValue {
			return rslt * rhs
		}
		solveRhs = func(rslt, lhs MonkeyValue) MonkeyValue {
			return lhs / rslt
		}
	default:
		panic("Invalid operation")
	}
	return ExtendedBinaryOp{
		Op:       op,
		SolveLhs: solveLhs,
		SolveRhs: solveRhs,
	}
}

func (parser *Parser) ParseLine(line string) (id string, monkey Monkey) {
	matches := parser.lineNum.FindStringSubmatch(line)
	if len(matches) == 3 {
		id = matches[1]
		num := MonkeyValue(ParseInt(matches[2]))
		human := id == "humn"
		monkey = &MonkeyNumber{Value: num, Human: human}
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
		if id == "humn" {
			panic("humn must be a number")
		}
		return
	}

	panic("Invalid input")
}

func main() {
	mode2 := false
	if (len(os.Args) > 1) && (os.Args[1] == "2") {
		mode2 = true
	}

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

	if !mode2 {
		_, rootVal := group.Get("root").Calculate(&group, 0)
		fmt.Println(rootVal)
	} else {
		x := group.Get("root").SolveRoot(&group)
		fmt.Println(x)
	}
}
