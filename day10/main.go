package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type InstructionType int

const (
	InstrNoop InstructionType = iota
	InstrAddx
)

type Instruction struct {
	Type  InstructionType
	Value int
}

type CpuState int

const (
	CSReady CpuState = iota
	CSAdding
)

type Cpu struct {
	cycle   int
	regX    int
	program []Instruction
	state   CpuState
	buffer  int
}

func MakeCpu(program []Instruction) (result Cpu) {
	result.cycle = 1
	result.regX = 1
	result.program = program
	return
}

func (cpu *Cpu) Cycle() int {
	return cpu.cycle
}

func (cpu *Cpu) X() int {
	return cpu.regX
}

func (cpu *Cpu) NextCycle() bool {
	if len(cpu.program) == 0 {
		return false
	}

	cpu.cycle++

	switch cpu.state {
	case CSReady:
		nextInstruction := cpu.program[0]
		cpu.program = cpu.program[1:]

		switch nextInstruction.Type {
		case InstrNoop:
		case InstrAddx:
			cpu.state = CSAdding
			cpu.buffer = nextInstruction.Value
		}
	case CSAdding:
		cpu.regX += cpu.buffer
		cpu.state = CSReady
	}

	return true
}

func ParseInstruction(str string) (result Instruction) {
	if str == "noop" {
		result.Type = InstrNoop
		return
	}

	fields := strings.Fields(str)
	if fields[0] == "addx" {
		result.Type = InstrAddx
		if len(fields) != 2 {
			panic("addx: exactly 1 argument expected")
		}
		var err error
		result.Value, err = strconv.Atoi(fields[1])
		if err != nil {
			panic(err)
		}
	} else {
		panic("Invalid instruction")
	}

	return
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	program := []Instruction{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			program = append(program, ParseInstruction(line))
		}
	}

	cpu := MakeCpu(program)

	const first = 20
	const interval = 40

	sigStrSum := 0

	for cpu.NextCycle() {
		if cpu.Cycle() < first {
			continue
		}

		if (cpu.Cycle()-first)%interval == 0 {
			sigStrSum += cpu.Cycle() * cpu.X()
			//fmt.Printf("%d: %d\n", cpu.Cycle(), cpu.X())
		}
	}
	fmt.Println(sigStrSum)
}
