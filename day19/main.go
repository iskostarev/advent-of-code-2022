package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Resource int

const (
	ResOre Resource = iota
	ResClay
	ResObsidian
	ResGeode

	ResCount
)

type Resources [ResCount]int
type Robots Resources
type Blueprint [ResCount]Resources

type State struct {
	Blueprint     *Blueprint
	Time, MaxTime int
	Robots        Robots
	Collected     Resources
}

type Parser struct {
	blueprint *regexp.Regexp
	robot     *regexp.Regexp
}

func ResourceFromString(s string) Resource {
	switch s {
	case "ore":
		return ResOre
	case "clay":
		return ResClay
	case "obsidian":
		return ResObsidian
	case "geode":
		return ResGeode
	}
	panic("Invalid resource name")
}

func (res Resource) String() string {
	switch res {
	case ResOre:
		return "ore"
	case ResClay:
		return "clay"
	case ResObsidian:
		return "obsidian"
	case ResGeode:
		return "geode"
	}
	panic("Invalid resource value")
}

func (resources Resources) String() (result string) {
	for r := 0; r < int(ResCount); r++ {
		if r != 0 {
			result += ","
		}
		result += fmt.Sprintf("%d %s", resources[r], Resource(r))
	}
	return
}

func (state State) String() string {
	return fmt.Sprintf("[T:%d/%d; Res: %s; Rob: %s]", state.Time, state.MaxTime, state.Collected, Resources(state.Robots))
}

func (state State) haveResources(req Resources) bool {
	for r := 0; r < int(ResCount); r++ {
		if state.Collected[r] < req[r] {
			return false
		}
	}
	return true
}

func (state *State) collect() {
	for r := 0; r < int(ResCount); r++ {
		state.Collected[r] += state.Robots[r]
	}
}

func (state State) tryBuild(rtype Resource) (bool, State) {
	if state.Time > state.MaxTime {
		return false, state
	}
	if !state.haveResources((*state.Blueprint)[rtype]) {
		return false, state
	}

	for r := 0; r < int(ResCount); r++ {
		req := (*state.Blueprint)[rtype][r]
		state.Collected[r] -= req
	}
	state.collect()
	state.Robots[rtype]++
	state.Time++
	return true, state
}

func (state State) decisions() []State {
	result := []State{}
	var build [ResCount]bool

	for ; state.Time < state.MaxTime; state.Time++ {
		for r := 0; r < int(ResCount); r++ {
			if build[r] {
				continue
			}
			ok, ns := state.tryBuild(Resource(r))
			if ok {
				build[r] = true
				result = append(result, ns)
			}
		}
		state.collect()
	}
	return append(result, state)
}

func (state State) upperBound() (result State) {
	result = state
	for ; result.Time < result.MaxTime; result.Time++ {
		for r := 0; r < int(ResCount); r++ {
			result.Collected[r] += result.Robots[r]
			result.Robots[r]++
		}
	}
	return
}

func (state State) maximize(cmp func(State, State) bool, maxSoFar *State, depth int, debug bool) State {
	indent := ""
	if debug {
		indent = strings.Repeat(" ", depth)
	}

	if state.Time == state.MaxTime {
		if cmp(*maxSoFar, state) {
			*maxSoFar = state
		}
		return state
	}

	max := state
	decisions := state.decisions()

	if debug {
		for _, decision := range decisions {
			fmt.Printf("%s+Decision: %v\n", indent, decision)
		}
	}

	for _, decision := range decisions {
		if !cmp(*maxSoFar, decision.upperBound()) {
			if debug {
				fmt.Printf("%s-Decision: %v; upperBound <= maxSoFar = %v\n", indent, decision, *maxSoFar)
			}
			continue
		}

		if debug {
			fmt.Printf("%s-Decision: %v\n", indent, decision)
		}
		ns := decision.maximize(cmp, maxSoFar, depth+1, debug)
		if cmp(max, ns) {
			max = ns
		}
	}

	if cmp(*maxSoFar, max) {
		*maxSoFar = max
	}
	return max
}

func (blueprint Blueprint) Maximize(minutes int, cmp func(State, State) bool) (result State) {
	state := State{}
	state.Blueprint = &blueprint
	state.MaxTime = minutes
	state.Robots[ResOre] = 1

	maxSoFar := State{}

	const debug = false
	return state.maximize(cmp, &maxSoFar, 0, debug)
}

func (blueprint Blueprint) MaxGeodes(minutes int) int {
	return blueprint.Maximize(minutes, func(lhs, rhs State) bool {
		return lhs.Collected[ResGeode] < rhs.Collected[ResGeode]
	}).Collected[ResGeode]
}

func MakeParser() (result Parser) {
	result.blueprint = regexp.MustCompile(`^Blueprint\s+(\d+):(.*)$`)
	result.robot = regexp.MustCompile(`^Each\s+(ore|clay|obsidian|geode)\s+robot\s+costs\s+(.*)$`)
	return
}

func ParseInt(s string) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return result
}

func (parser *Parser) ParseCost(cost string) (result Resources) {
	for _, part := range strings.Split(cost, " and ") {
		fields := strings.Fields(part)
		if len(fields) != 2 {
			panic("Invalid cost")
		}
		cost := ParseInt(fields[0])
		res := ResourceFromString(fields[1])
		result[res] = cost
	}
	return
}

func (parser *Parser) ParseLine(line string) (id int, result Blueprint) {
	matches := parser.blueprint.FindStringSubmatch(line)
	if len(matches) != 3 {
		panic("Invalid line")
	}
	id = ParseInt(matches[1])
	for _, statement := range strings.Split(matches[2], ".") {
		statement := strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		robotMatches := parser.robot.FindStringSubmatch(statement)
		robotType := ResourceFromString(robotMatches[1])
		result[robotType] = parser.ParseCost(robotMatches[2])
	}

	for _, cost := range result {
		var empty Resources
		if cost == empty {
			panic("Empty cost")
		}
	}
	return
}

func mode1() {
	scanner := bufio.NewScanner(os.Stdin)
	parser := MakeParser()

	sum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		id, blueprint := parser.ParseLine(line)
		max := blueprint.MaxGeodes(24)
		sum += id * max
		//fmt.Printf("%d: %d\n", id, max)
	}
	fmt.Println(sum)
}

func mode2() {
	scanner := bufio.NewScanner(os.Stdin)
	parser := MakeParser()

	count := 0
	product := 1
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		id, blueprint := parser.ParseLine(line)
		count++
		max := blueprint.MaxGeodes(32)
		//fmt.Printf("%d: %d\n", id, max)
		product *= max
		if count == 3 {
			break
		}
	}
	fmt.Println(product)
}

func main() {
	if (len(os.Args) > 1) && (os.Args[1] == "2") {
		mode2()
		return
	}

	mode1()

}
