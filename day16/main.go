package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Valve struct {
	FlowRate int
	Edges    []string
}

type ValveSet struct {
	contents map[string]bool
	cow      bool
}

type Cave struct {
	Graph     map[string]Valve
	Distances map[string]map[string]int
}

type State struct {
	CurValve string
	Open     ValveSet
}

type Parser struct {
	regex *regexp.Regexp
}

func (set *ValveSet) Size() int {
	return len(set.contents)
}

func (set *ValveSet) Has(valve string) bool {
	if set.contents == nil {
		return false
	}
	return set.contents[valve]
}

func (set *ValveSet) Add(valve string) {
	if set.contents == nil {
		set.contents = map[string]bool{}
		set.cow = false
	}
	if set.cow {
		newContents := map[string]bool{}
		for k, v := range set.contents {
			newContents[k] = v
		}
		set.contents = newContents
		set.cow = false
	}
	set.contents[valve] = true
}

func (set *ValveSet) Traverse(cb func(string)) {
	if set.contents == nil {
		return
	}
	for valve, _ := range set.contents {
		cb(valve)
	}
}

func (set *ValveSet) Clear() {
	set.contents = nil
	set.cow = false
}

func (set *ValveSet) Copy() (result ValveSet) {
	result.contents = set.contents
	result.cow = true
	return
}

func (state *State) Copy() (result State) {
	result.CurValve = state.CurValve
	result.Open = state.Open.Copy()
	return
}

func (cave *Cave) TotalFlowRate() (result int) {
	for _, valve := range cave.Graph {
		result += valve.FlowRate
	}
	return
}

func (cave *Cave) doSearchForMaxPressure(state State, pressure, remaining, depth int) int {
	printLine := func(line string) {
		/*
			for i := 0; i < depth; i++ {
				fmt.Print(" ")
			}
			fmt.Printf("[-%d] Pressure=%d, State=%v: %s\n", remaining, pressure, state, line)
		*/
	}

	printLine("started")

	if remaining == 0 {
		printLine("no more time")
		return pressure
	}

	if state.Open.Size() == len(cave.Graph) {
		printLine("no more actions")
		return cave.TotalFlowRate() * remaining
	}

	dPressure := 0
	state.Open.Traverse(func(valve string) {
		dPressure += cave.Graph[valve].FlowRate
	})

	maxCandidate := pressure + dPressure*remaining

	if !state.Open.Has(state.CurValve) && cave.Graph[state.CurValve].FlowRate > 0 {
		ns := state.Copy()
		ns.Open.Add(state.CurValve)
		printLine("considering cand open")
		maxCandidate = cave.doSearchForMaxPressure(ns, pressure+dPressure, remaining-1, depth+1)
		printLine(fmt.Sprintf("cand open: %d", maxCandidate))
	}

	for target, _ := range cave.Graph {
		if state.Open.Has(target) || cave.Graph[target].FlowRate <= 0 {
			continue
		}

		dist := cave.Distances[state.CurValve][target]
		if dist == -1 || dist >= remaining {
			continue
		}

		minutes := dist + 1
		ns := state.Copy()
		ns.CurValve = target
		ns.Open.Add(target)

		printLine(fmt.Sprintf("considering cand go to %s", target))
		cand := cave.doSearchForMaxPressure(ns, pressure+dPressure*minutes, remaining-minutes, depth+1)
		printLine(fmt.Sprintf("cand go to %s: %d", target, maxCandidate))
		if cand > maxCandidate {
			maxCandidate = cand
		}
	}

	return maxCandidate
}

func (cave *Cave) MaxPressure() int {
	const minutes = 30
	state := State{CurValve: "AA"}
	return cave.doSearchForMaxPressure(state, 0, minutes, 0)
}

func (cave *Cave) CalcDistance(v1, v2 string, visited ValveSet) (result int) {
	result, ok := cave.Distances[v1][v2]
	if ok {
		return
	}

	result = -1
	for _, edge := range cave.Graph[v1].Edges {
		if edge == v2 {
			result = 1
			break
		}
	}

	if result == -1 {
		visited = visited.Copy()
		visited.Add(v1)
		for _, edge := range cave.Graph[v1].Edges {
			if visited.Has(edge) {
				continue
			}
			dist := cave.CalcDistance(edge, v2, visited)
			if dist != -1 {
				dist++
				if result == -1 || dist < result {
					result = dist
				}
			}
		}
	}

	if result != -1 {
		cave.Distances[v1][v2] = result
	}
	return
}

func (cave *Cave) CalcDistances() {
	cave.Distances = map[string]map[string]int{}
	for v, _ := range cave.Graph {
		cave.Distances[v] = map[string]int{}
		cave.Distances[v][v] = 0
	}

	for v1, _ := range cave.Graph {
		for v2, _ := range cave.Graph {
			cave.CalcDistance(v1, v2, ValveSet{})
		}
	}
}

func MakeParser() (result Parser) {
	result.regex = regexp.MustCompile(`^Valve ([A-Z]+) has flow rate=(\d+); tunnels? leads? to valves? ([A-Z]+(?:, [A-Z]+)*)$`)
	return
}

func ParseInt(s string) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return result
}

func ParseList(list string) (result []string) {
	result = []string{}
	for _, item := range strings.Split(list, ",") {
		item = strings.TrimSpace(item)
		result = append(result, item)
	}
	return result
}

func (parser *Parser) ParseLine(line string) (valve string, rate int, edges []string) {
	matches := parser.regex.FindStringSubmatch(line)
	if len(matches) != 4 {
		panic("Invalid line")
	}
	valve = matches[1]
	rate = ParseInt(matches[2])
	edges = ParseList(matches[3])
	return
}

func ParseInput(scanner *bufio.Scanner) (result Cave) {
	result.Graph = map[string]Valve{}

	parser := MakeParser()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		valve, rate, edges := parser.ParseLine(line)
		result.Graph[valve] = Valve{FlowRate: rate, Edges: edges}
	}

	result.CalcDistances()
	return
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	cave := ParseInput(scanner)
	fmt.Println(cave.MaxPressure())
}
