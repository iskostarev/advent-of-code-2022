package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type ActorState int

const (
	StateThinking ActorState = iota
	StateHeadingToValve
	StateOpening
	StateIdling
)

type Actor struct {
	State    ActorState
	Location string
	Turns    int
}

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
	Vertices  []string
	Distances map[string]map[string]int
}

type State struct {
	Open      ValveSet
	Targeted  ValveSet
	Actors    []Actor
	Remaining int
	MaxTime   int
	Pressure  int
	Log       []string
}

type Parser struct {
	regex *regexp.Regexp
}

func (set *ValveSet) String() string {
	result := "{"
	first := true
	for item, _ := range set.contents {
		if !first {
			result += ","
		}
		first = false
		result += item
	}
	result += "}"
	return result
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
	result.Open = state.Open.Copy()
	result.Targeted = state.Targeted.Copy()

	result.Actors = make([]Actor, len(state.Actors))
	copy(result.Actors, state.Actors)

	if state.Log != nil {
		result.Log = make([]string, len(state.Log))
		copy(result.Log, state.Log)
	}

	result.Remaining = state.Remaining
	result.MaxTime = state.MaxTime
	result.Pressure = state.Pressure
	return
}

func (state *State) Decided() bool {
	for _, actor := range state.Actors {
		if actor.State == StateThinking {
			return false
		}
	}
	return true
}

func (state *State) String() string {
	actors := "["
	for i, actor := range state.Actors {
		if i != 0 {
			actors += ";"
		}
		switch actor.State {
		case StateThinking:
			actors += fmt.Sprintf("{THINKING AT %s}", actor.Location)
		case StateHeadingToValve:
			actors += fmt.Sprintf("{HEADING TO %s, %d}", actor.Location, actor.Turns)
		case StateOpening:
			actors += fmt.Sprintf("{OPENING %s}", actor.Location)
		case StateIdling:
			actors += fmt.Sprintf("{IDLING}")
		}
	}
	actors += "]"
	t := state.MaxTime - state.Remaining + 1
	return fmt.Sprintf("[T: %d, Prs: %d, Open: %s, Tar: %s] %s", t, state.Pressure, state.Open.String(), state.Targeted.String(), actors)
}

func (state *State) AddLog(msg string) {
	if state.Log == nil {
		return
	}

	line := fmt.Sprintf("%s %s", state.String(), msg)

	state.Log = append(state.Log, line)
}

func (cave *Cave) TotalFlowRate() (result int) {
	for _, valve := range cave.Graph {
		result += valve.FlowRate
	}
	return
}

func (cave *Cave) doIterateNextMoves(state State, index int, cb func(State)) {
	if index == len(state.Actors) {
		cb(state)
		return
	}

	if state.Actors[index].State != StateThinking {
		cave.doIterateNextMoves(state, index+1, cb)
		return
	}

	for _, target := range cave.Vertices {
		if !state.Open.Has(target) && !state.Targeted.Has(target) && cave.Graph[target].FlowRate > 0 {
			dist := cave.Distances[state.Actors[index].Location][target]
			if dist == -1 || dist >= state.Remaining {
				continue
			}

			ns := state.Copy()

			ns.Actors[index].State = StateHeadingToValve
			ns.Actors[index].Location = target
			ns.Actors[index].Turns = dist
			ns.Targeted.Add(target)

			cave.doIterateNextMoves(ns, index+1, cb)
		}
	}

	state.Actors[index].State = StateIdling
	cave.doIterateNextMoves(state, index+1, cb)
}

func (cave *Cave) iterateNextMoves(state State, cb func(State)) {
	idling := []State{}
	cave.doIterateNextMoves(state.Copy(), 0, func(ns State) {
		eager := true

		for _, actor := range ns.Actors {
			if actor.State == StateIdling {
				eager = false
				break
			}
		}

		if eager {
			idling = nil
			cb(ns)
		} else if idling != nil {
			idling = append(idling, ns)
		}
	})

	if idling != nil {
		for _, ns := range idling {
			cb(ns)
		}
	}
}

func (cave *Cave) doSearchForMaxPressure(state State, depth int, max *int) State {
	// for i := 0; i < depth; i++ {
	// 	fmt.Print(" ")
	// }
	// fmt.Println(state.String())

	state.AddLog("begin")

	if state.Remaining == 0 {
		state.AddLog("no time left")

		if *max < state.Pressure {
			*max = state.Pressure
		}
		return state
	}

	dPressure := 0
	state.Open.Traverse(func(valve string) {
		dPressure += cave.Graph[valve].FlowRate
	})

	for state.Decided() {
		state.Remaining--
		state.Pressure += dPressure

		for i, _ := range state.Actors {
			switch state.Actors[i].State {
			case StateThinking:
				panic("Impossible state")
			case StateHeadingToValve:
				if state.Actors[i].Turns <= 0 {
					panic("Impossible condition")
				}
				state.Actors[i].Turns--
				if state.Actors[i].Turns == 0 {
					state.Actors[i].State = StateOpening
				}
			case StateOpening:
				state.Actors[i].State = StateThinking
				if !state.Open.Has(state.Actors[i].Location) {
					state.Open.Add(state.Actors[i].Location)
					dPressure += cave.Graph[state.Actors[i].Location].FlowRate
				}
			}
		}

		state.AddLog("automatic")

		if state.Remaining == 0 {
			if *max < state.Pressure {
				*max = state.Pressure
			}
			return state
		}

	}

	basePressure := dPressure * state.Remaining

	maxCandidate := state.Copy()
	maxCandidate.Pressure += basePressure
	maxCandidate.AddLog(fmt.Sprintf("do nothing, get %d = %d*%d", basePressure, dPressure, state.Remaining))

	upperBound := state.Pressure + state.Remaining*cave.TotalFlowRate()
	if upperBound <= *max {
		return maxCandidate
	}

	cave.iterateNextMoves(state, func(ns State) {
		st := cave.doSearchForMaxPressure(ns, depth+1, max)
		if st.Pressure > maxCandidate.Pressure {
			maxCandidate = st
		}
	})

	if *max < maxCandidate.Pressure {
		*max = maxCandidate.Pressure
	}
	return maxCandidate
}

func (cave *Cave) MaxPressure(useElephant bool) int {
	minutes := 30
	const init = "AA"

	if cave.Graph[init].FlowRate > 0 {
		panic("Not implemented")
	}

	state := State{}
	state.Actors = []Actor{Actor{Location: init}}

	if useElephant {
		minutes -= 4
		state.Actors = append(state.Actors, Actor{Location: init})
	}

	state.Remaining = minutes
	state.MaxTime = minutes
	//state.Log = []string{}

	max := 0
	finalState := cave.doSearchForMaxPressure(state, 0, &max)
	if finalState.Log != nil {
		for _, line := range finalState.Log {
			fmt.Println(line)
		}
	}
	return finalState.Pressure
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
	result.Vertices = []string{}

	parser := MakeParser()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		valve, rate, edges := parser.ParseLine(line)
		result.Graph[valve] = Valve{FlowRate: rate, Edges: edges}
		result.Vertices = append(result.Vertices, valve)
	}

	slices.SortFunc(result.Vertices, func(lhs, rhs string) bool {
		return result.Graph[lhs].FlowRate < result.Graph[rhs].FlowRate
	})

	result.CalcDistances()
	return
}

func main() {
	mode2 := false
	if (len(os.Args) > 1) && (os.Args[1] == "2") {
		mode2 = true
	}

	scanner := bufio.NewScanner(os.Stdin)
	cave := ParseInput(scanner)
	fmt.Println(cave.MaxPressure(mode2))
}
