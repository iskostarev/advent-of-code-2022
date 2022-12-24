package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type CellType int
type Direction int

const (
	CellEmpty CellType = iota
	CellWall
)

const (
	DirUp Direction = iota
	DirRight
	DirDown
	DirLeft
)

const NumDirections = 4

type Pos struct {
	X, Y int
}

type Cell struct {
	Type      CellType
	Blizzards []Direction
}

type Valley struct {
	sizeX int
	cells [][]Cell
}

func (dir Direction) DxDy() (int, int) {
	switch dir {
	case DirUp:
		return 0, -1
	case DirDown:
		return 0, 1
	case DirLeft:
		return -1, 0
	case DirRight:
		return 1, 0
	}

	panic("Invalid direction")
}

func (dir Direction) Opposide() Direction {
	return Direction((int(dir) + 2) % NumDirections)
}

func (dir Direction) String() string {
	switch dir {
	case DirUp:
		return "^"
	case DirDown:
		return "v"
	case DirLeft:
		return "<"
	case DirRight:
		return ">"
	}

	panic("Invalid direction")
}

func (pos Pos) Move(dir Direction) Pos {
	dx, dy := dir.DxDy()
	pos.X += dx
	pos.Y += dy
	return pos
}

func (pos Pos) ForEachMove(cb func(Pos)) {
	cb(pos)
	for i := 0; i < NumDirections; i++ {
		cb(pos.Move(Direction(i)))
	}
}

func (cell *Cell) IsEmpty() bool {
	if cell.Type != CellEmpty {
		return false
	}
	return len(cell.Blizzards) == 0
}

func (cell *Cell) String() string {
	if cell.Type == CellWall {
		return "#"
	}

	if cell.Type != CellEmpty {
		panic("Invalid cell type")
	}

	if len(cell.Blizzards) == 0 {
		return "."
	}

	if len(cell.Blizzards) == 1 {
		return cell.Blizzards[0].String()
	}

	if len(cell.Blizzards) <= 9 {
		return fmt.Sprintf("%d", len(cell.Blizzards))
	}

	return "+"
}

func MakeValley() (result Valley) {
	result.cells = [][]Cell{}
	return
}

func (valley *Valley) CopyBase() (result Valley) {
	result.sizeX = valley.sizeX
	result.cells = make([][]Cell, len(valley.cells))
	for i := 0; i < len(valley.cells); i++ {
		result.cells[i] = make([]Cell, result.sizeX)
		for j := 0; j < result.sizeX; j++ {
			result.cells[i][j].Type = valley.cells[i][j].Type
		}
	}
	return
}

func (valley *Valley) AppendRow(row []Cell) {
	if len(valley.cells) == 0 {
		valley.sizeX = len(row)
	} else {
		if valley.sizeX != len(row) {
			panic("Row length mismatch")
		}
	}
	valley.cells = append(valley.cells, row)
}

func (valley *Valley) SizeX() int {
	return valley.sizeX
}

func (valley *Valley) SizeY() int {
	return len(valley.cells)
}

func (valley *Valley) ValidCoords(x, y int) bool {
	if x < 0 || x >= valley.sizeX || y < 0 || y >= len(valley.cells) {
		return false
	}
	return true
}

func (valley *Valley) At(x, y int) *Cell {
	if !valley.ValidCoords(x, y) {
		panic("Invalid coords")
	}
	return &valley.cells[y][x]
}

func (valley *Valley) AddBlizzard(x, y int, blizzard Direction) {
	cell := valley.At(x, y)
	if cell.Blizzards == nil {
		cell.Blizzards = []Direction{blizzard}
	} else {
		cell.Blizzards = append(cell.Blizzards, blizzard)
	}
}

func (valley *Valley) FindSingleEmptyCell(y int) (result int) {
	found := false
	for x := 0; x < valley.SizeX(); x++ {
		if valley.At(x, y).Type == CellEmpty {
			if found {
				panic("Only one empty cell expected")
			}
			result = x
			found = true
		}
	}

	if !found {
		panic("No empty row found")
	}

	return
}

func (valley *Valley) FindInitPos() (result Pos) {
	result.Y = 0
	result.X = valley.FindSingleEmptyCell(result.Y)
	return
}

func (valley *Valley) FindGoalPos() (result Pos) {
	result.Y = valley.SizeY() - 1
	result.X = valley.FindSingleEmptyCell(result.Y)
	return
}

func wrap(val, min, max int) int {
	if val < min {
		val = max
	} else if val > max {
		val = min
	}
	return val
}

func (valley *Valley) MoveBlizzard(x, y int, dir Direction) Pos {
	dx, dy := dir.DxDy()
	for {
		x = wrap(x+dx, 0, valley.SizeX()-1)
		y = wrap(y+dy, 0, valley.SizeY()-1)
		if valley.At(x, y).Type == CellEmpty {
			return Pos{x, y}
		}
	}
}

func (valley *Valley) Next() (result Valley) {
	result = valley.CopyBase()
	for y := 0; y < valley.SizeY(); y++ {
		for x := 0; x < valley.SizeX(); x++ {
			for _, blizzard := range valley.At(x, y).Blizzards {
				pos := valley.MoveBlizzard(x, y, blizzard)
				result.AddBlizzard(pos.X, pos.Y, blizzard)
			}
		}
	}
	return
}

func (valley Valley) FindMinPathSteps(init, goal Pos) (steps int) {
	positions := map[Pos]bool{init: true}
	for {
		valley = valley.Next()
		nextPositions := map[Pos]bool{}

		for p, _ := range positions {
			if p == goal {
				return steps
			}
			p.ForEachMove(func(np Pos) {
				if valley.ValidCoords(np.X, np.Y) && valley.At(np.X, np.Y).IsEmpty() {
					nextPositions[np] = true
				}
			})
		}

		steps++
		positions = nextPositions
	}
}

func (valley *Valley) Print() {
	for y := 0; y < valley.SizeY(); y++ {
		for x := 0; x < valley.SizeX(); x++ {
			fmt.Print(valley.At(x, y))
		}
		fmt.Println()
	}
}

func ParseRow(line string) (result []Cell) {
	result = make([]Cell, len(line))
	for i := 0; i < len(line); i++ {
		var dir Direction

		switch line[i] {
		case '.':
			continue
		case '#':
			result[i].Type = CellWall
			continue
		case '^':
			dir = DirUp
		case '>':
			dir = DirRight
		case 'v':
			dir = DirDown
		case '<':
			dir = DirLeft
		default:
			panic("Unexpected char")
		}
		result[i].Blizzards = []Direction{dir}
	}
	return
}

func ParseInput(scanner *bufio.Scanner) (valley Valley) {
	valley = MakeValley()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		valley.AppendRow(ParseRow(line))
	}
	return
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	valley := ParseInput(scanner)
	init := valley.FindInitPos()
	goal := valley.FindGoalPos()
	fmt.Println(valley.FindMinPathSteps(init, goal))
}
