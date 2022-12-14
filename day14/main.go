package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Cell int

const (
	CellEmpty Cell = iota
	CellRock
	CellSand
)

type Pos struct {
	X, Y int
}

type Cave struct {
	cells      [][]Cell
	sizeY      int
	sandSource Pos
	sandCount  int
	floor      bool
}

func (cell Cell) String() string {
	switch cell {
	case CellEmpty:
		return "."
	case CellRock:
		return "#"
	case CellSand:
		return "o"
	}
	panic("Invalid cell")
}

func (pos Pos) Delta(delta Pos) (result Pos) {
	result = pos
	result.X += delta.X
	result.Y += delta.Y
	return
}

func MakeCave() (result Cave) {
	result.cells = [][]Cell{}
	return
}

func (cave *Cave) SetSandSource(pos Pos) {
	cave.sandSource = pos
}

func (cave *Cave) SetFloor(y int) {
	cave.floor = true
	if y < cave.SizeY() {
		panic("Floor not low enough")
	}
	cave.ResizeY(y)
}

func (cave *Cave) SizeX() int {
	return len(cave.cells)
}

func (cave *Cave) SizeY() int {
	return cave.sizeY
}

func (cave *Cave) SandCount() int {
	return cave.sandCount
}

func (cave *Cave) ResizeX(newSize int) {
	for i := len(cave.cells); i < newSize; i++ {
		newRow := make([]Cell, cave.SizeY())
		cave.cells = append(cave.cells, newRow)
	}
}

func (cave *Cave) ResizeY(newSize int) {
	for i := 0; i < len(cave.cells); i++ {
		newRow := make([]Cell, newSize)
		copy(newRow, cave.cells[i])
		cave.cells[i] = newRow
	}
	cave.sizeY = newSize
}

func (cave *Cave) Set(point Pos, cell Cell) {
	if point.Y >= cave.SizeY() {
		cave.ResizeY(point.Y + 1)
	}
	if point.X >= cave.SizeX() {
		cave.ResizeX(point.X + 1)
	}
	if cell == CellSand && cave.cells[point.X][point.Y] != CellSand {
		cave.sandCount++
	}
	cave.cells[point.X][point.Y] = cell
}

func (cave *Cave) At(point Pos) Cell {
	return cave.cells[point.X][point.Y]
}

func (cave *Cave) IsFloor(point Pos) bool {
	if !cave.floor {
		return false
	}
	return point.Y == cave.SizeY()
}

func (cave *Cave) InBounds(point Pos) bool {
	if point.X < 0 || point.X >= cave.SizeX() {
		return false
	}
	if point.Y < 0 || point.Y >= cave.SizeY() {
		return false
	}
	return true
}

func (cave *Cave) String() (result string) {
	for y := 0; y < cave.SizeY(); y++ {
		for x := 0; x < cave.SizeX(); x++ {
			result += cave.At(Pos{x, y}).String()
		}
		result += "\n"
	}
	return
}

func (cave *Cave) AddStraightLine(a, b1, b2 int, x bool) {
	if b2 < b1 {
		b1, b2 = b2, b1
	}

	for b := b1; b <= b2; b++ {
		var pos Pos
		if x {
			pos = Pos{b, a}
		} else {
			pos = Pos{a, b}
		}
		cave.Set(pos, CellRock)
	}
}

func (cave *Cave) AddLine(line []Pos) {
	prev := line[0]
	for _, pos := range line[1:] {
		if prev.X == pos.X {
			cave.AddStraightLine(pos.X, prev.Y, pos.Y, false)
		} else if prev.Y == pos.Y {
			cave.AddStraightLine(pos.Y, prev.X, pos.X, true)
		} else {
			panic("Invalid line")
		}

		prev = pos
	}
}

func (cave *Cave) AddSand() bool {
	sand := cave.sandSource

outer:
	for {
		prev := sand
		if cave.At(prev) == CellSand {
			return false
		} else if cave.At(prev) != CellEmpty {
			panic("Cell expected to be empty")
		}

		for _, d := range [...]Pos{Pos{0, 1}, Pos{-1, 1}, Pos{1, 1}} {
			sand = prev.Delta(d)
			if sand.X >= cave.SizeX() {
				cave.ResizeX(sand.X + 1)
			}
			if cave.IsFloor(sand) {
				continue
			}
			if !cave.InBounds(sand) {
				return false
			}
			if cave.At(sand) == CellEmpty {
				continue outer
			}
		}

		cave.Set(prev, CellSand)
		return true
	}
}

func ParseInt(str string) (result int) {
	result, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return
}

func ParsePoint(str string) (result Pos) {
	coords := strings.Split(str, ",")
	if len(coords) != 2 {
		panic("Point must have exactly 2 coordinates")
	}
	result.X = ParseInt(coords[0])
	result.Y = ParseInt(coords[1])
	return
}

func ParsePath(str string) (result []Pos) {
	result = []Pos{}
	for _, point := range strings.Split(str, "->") {
		point := strings.TrimSpace(point)
		result = append(result, ParsePoint(point))
	}
	return
}

func main() {
	mode2 := false
	if (len(os.Args) > 1) && (os.Args[1] == "2") {
		mode2 = true
	}

	scanner := bufio.NewScanner(os.Stdin)
	cave := MakeCave()
	cave.SetSandSource(Pos{500, 0})

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		cave.AddLine(ParsePath(line))
	}

	if mode2 {
		cave.SetFloor(cave.SizeY() + 1)
	}

	for cave.AddSand() {
	}
	fmt.Println(cave.SandCount())
}
