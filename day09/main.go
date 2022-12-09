package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

const DirCount = 4

type Motion struct {
	Dir   Direction
	Steps int
}

type Pos struct {
	X, Y int
}

type Rope struct {
	knots []Pos
}

type PositionMap struct {
	visited map[Pos]bool
}

func (dir Direction) String() string {
	switch dir {
	case DirUp:
		return "U"
	case DirDown:
		return "D"
	case DirLeft:
		return "L"
	case DirRight:
		return "R"
	}
	panic("Invalid direction")
}

func (dir Direction) DxDy() (int, int) {
	switch dir {
	case DirUp:
		return 0, 1
	case DirDown:
		return 0, -1
	case DirLeft:
		return -1, 0
	case DirRight:
		return 1, 0
	}
	panic("Invalid direction")
}

func ParseDirection(s string) Direction {
	switch s {
	case "U":
		return DirUp
	case "D":
		return DirDown
	case "L":
		return DirLeft
	case "R":
		return DirRight
	default:
		panic("Failed to parse direction")
	}
}

func ParseMotion(s string) (result Motion) {
	fields := strings.Fields(s)
	if len(fields) != 2 {
		panic("Expected 2 fields")
	}
	result.Dir = ParseDirection(fields[0])

	var err error
	result.Steps, err = strconv.Atoi(fields[1])
	if err != nil {
		panic("Failed to parse step count")
	}
	return
}

func Abs(val int) int {
	if val >= 0 {
		return val
	} else {
		return -val
	}
}

func Adjacent(lhs, rhs Pos) bool {
	return Abs(lhs.X-rhs.X) <= 1 && Abs(lhs.Y-rhs.Y) <= 1
}

func (pos *Pos) ApplyDirection(direction Direction) {
	dx, dy := direction.DxDy()
	pos.X += dx
	pos.Y += dy
}

func MoveCoordTowards(coord *int, target int) {
	if *coord > target {
		*coord--
	} else if *coord < target {
		*coord++
	}
}

func (pos *Pos) MoveTowards(target Pos) {
	MoveCoordTowards(&pos.X, target.X)
	MoveCoordTowards(&pos.Y, target.Y)
}

func (rope *Rope) ApplyDirection(dir Direction) {
	rope.Head().ApplyDirection(dir)
	prev := *rope.Head()

	for i := 1; i < len(rope.knots); i++ {
		if Adjacent(prev, rope.knots[i]) {
			return
		}
		rope.knots[i].MoveTowards(prev)
		prev = rope.knots[i]
	}
}

func MakeRope(knotCount int) (result Rope) {
	if knotCount < 2 {
		panic("Invalid knot count")
	}
	result.knots = make([]Pos, knotCount)
	return
}

func (rope *Rope) Head() *Pos {
	return &rope.knots[0]
}

func (rope *Rope) Tail() *Pos {
	return &rope.knots[len(rope.knots)-1]
}

func MakePositionMap() (result PositionMap) {
	result.visited = make(map[Pos]bool)
	return
}

func (posmap *PositionMap) MarkPos(pos Pos) {
	posmap.visited[pos] = true
}

func (posmap *PositionMap) TraverseVisited(cb func(Pos)) {
	for k, v := range posmap.visited {
		if v {
			cb(k)
		}
	}
}

func main() {
	knotCount := 2

	if len(os.Args) > 1 {
		var err error
		knotCount, err = strconv.Atoi(os.Args[1])
		if err != nil {
			panic(err)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	rope := MakeRope(knotCount)
	posmap := MakePositionMap()

	posmap.MarkPos(*rope.Tail())

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		motion := ParseMotion(line)
		for i := 0; i < motion.Steps; i++ {
			rope.ApplyDirection(motion.Dir)
			posmap.MarkPos(*rope.Tail())
		}
	}
	visited := 0
	posmap.TraverseVisited(func(Pos) {
		visited++
	})
	fmt.Println(visited)
}
