package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Crate byte
type Stack []Crate

type Crates struct {
	Size   int
	Stacks []Stack
}

type MoveInstruction struct {
	Count, From, To int
}

type MoveInstructionParser struct {
	regex *regexp.Regexp
}

func (crate Crate) String() string {
	return string(byte(crate))
}

func (crates *Crates) validateStackIndex(index int) {
	if index < 0 || index >= crates.Size {
		panic("Invalid crate index")
	}
}

func (crates *Crates) Push(stack int, crate Crate) {
	stack--
	crates.validateStackIndex(stack)
	crates.Stacks[stack] = append(crates.Stacks[stack], crate)
}

func (crates *Crates) Pop(stack int) (crate Crate) {
	stack--
	crates.validateStackIndex(stack)

	stackLen := len(crates.Stacks[stack])
	if stackLen == 0 {
		panic("Popping empty stack")
	}
	crate = crates.Stacks[stack][stackLen-1]
	crates.Stacks[stack] = crates.Stacks[stack][0 : stackLen-1]
	return
}

func (crates *Crates) Top(stack int) (crate Crate) {
	stack--
	crates.validateStackIndex(stack)

	return crates.Stacks[stack][len(crates.Stacks[stack])-1]
}

func ParseCrates(scanner *bufio.Scanner) (result Crates) {
	crateLines := []string{}

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n")
		if line == "" {
			break
		}
		crateLines = append(crateLines, line)
	}

	crateNumbers := parseCrateLine(crateLines[len(crateLines)-1])
	result.Size = len(crateNumbers)
	result.Stacks = make([]Stack, result.Size)

	for i, numstr := range crateNumbers {
		if numstr != fmt.Sprintf(" %d ", i+1) {
			panic("Invalid crate order")
		}
	}

	done := make([]bool, result.Size)
	for i := len(crateLines) - 2; i >= 0; i-- {
		for crateIdx, crateStr := range parseCrateLine(crateLines[i]) {
			empty, crate := parseCrateString(crateStr)
			if empty {
				done[crateIdx] = true
			} else {
				if done[crateIdx] {
					panic("Unexpected space in crate stack")
				}
				result.Push(crateIdx+1, crate)
			}
		}
	}

	return
}

func parseCrateLine(line string) (result []string) {
	const COL_WIDTH = 4

	if (len(line)+1)%COL_WIDTH != 0 {
		panic("Invalid line length")
	}
	count := (len(line) + 1) / COL_WIDTH
	result = make([]string, count)

	for i := 0; i < count; i++ {
		if i != count-1 && line[i*COL_WIDTH+COL_WIDTH-1] != ' ' {
			panic("Invalid column separator")
		}
		result[i] = line[i*COL_WIDTH : (i+1)*COL_WIDTH-1]
	}

	return
}

func parseCrateString(str string) (empty bool, crate Crate) {
	if str == "   " {
		empty = true
	} else {
		if str[0] != '[' || str[2] != ']' {
			panic(fmt.Sprintf("Invalid crate string format: %s", str))
		}
		empty = false
		crate = Crate(str[1])
	}
	return
}

func MakeMoveInstructionParser() (result MoveInstructionParser) {
	result.regex = regexp.MustCompile(`^move (\d+) from (\d) to (\d)$`)
	return
}

func (parser *MoveInstructionParser) Parse(line string) (result MoveInstruction) {
	onFail := func() {
		panic(fmt.Sprintf("Failed to parse line: %s", line))
	}

	matches := parser.regex.FindStringSubmatch(line)
	if len(matches) != 4 {
		onFail()
	}

	parseInt := func(str string) (res int) {
		res, err := strconv.Atoi(str)
		if err != nil {
			onFail()
		}
		return
	}

	result.Count = parseInt(matches[1])
	result.From = parseInt(matches[2])
	result.To = parseInt(matches[3])
	return
}

func ApplyMoveInstruction(crates *Crates, ins MoveInstruction) {
	for i := 0; i < ins.Count; i++ {
		crates.Push(ins.To, crates.Pop(ins.From))
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	crates := ParseCrates(scanner)
	moveInsParser := MakeMoveInstructionParser()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		moveIns := moveInsParser.Parse(line)
		ApplyMoveInstruction(&crates, moveIns)
	}

	for i := 1; i <= crates.Size; i++ {
		fmt.Print(crates.Top(i))
	}
	fmt.Println()
}
