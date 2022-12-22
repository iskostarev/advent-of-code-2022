package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Rotation int

const (
	RotNone Rotation = iota
	RotLeft
	RotRight
)

type Tile int

const (
	TileVoid = iota
	TileOpen
	TileWall
)

type Direction int

const (
	DirRight Direction = 0
	DirDown            = 1
	DirLeft            = 2
	DirUp              = 3
)

type Pos struct {
	X, Y int
	Dir  Direction
}

type Board struct {
	rows [][]Tile
}

type Instruction struct {
	Rotate Rotation
	Steps  int
}

func (rot Rotation) String() string {
	switch rot {
	case RotNone:
		return "_"
	case RotLeft:
		return "L"
	case RotRight:
		return "R"
	}
	panic("Invalid rotation")
}

func (dir Direction) String() string {
	switch dir {
	case DirRight:
		return ">"
	case DirDown:
		return "v"
	case DirLeft:
		return "<"
	case DirUp:
		return "^"
	}
	panic("Invalid direction")
}

func (dir Direction) DxDy() (int, int) {
	switch dir {
	case DirRight:
		return 1, 0
	case DirDown:
		return 0, 1
	case DirLeft:
		return -1, 0
	case DirUp:
		return 0, -1
	}
	panic("Invalid direction")
}

func (tile Tile) String() string {
	switch tile {
	case TileVoid:
		return " "
	case TileOpen:
		return "."
	case TileWall:
		return "#"
	}
	panic("Invalid tile")
}

func MakeBoard() (result Board) {
	result.rows = [][]Tile{}
	return
}

func (board *Board) AppendRow(row []Tile) {
	board.rows = append(board.rows, row)
}

func (board *Board) At(x, y int) Tile {
	x--
	y--
	if y < 0 || y >= len(board.rows) {
		return TileVoid
	}

	if x < 0 || x >= len(board.rows[y]) {
		return TileVoid
	}

	return board.rows[y][x]
}

func (board *Board) YSize() int {
	return len(board.rows)
}

func (board *Board) XSize(y int) int {
	for y < 1 {
		y += board.YSize()
	}
	for y > board.YSize() {
		y -= board.YSize()
	}
	return len(board.rows[y-1])
}

func (board *Board) StartingPosition() (result Pos) {
	result.Y = 1
	for x := 1; x <= board.XSize(1); x++ {
		if board.At(x, 1) == TileOpen {
			result.X = x
			return
		}
	}
	panic("Failed to find starting position")
}

func (board *Board) Shift(x, y, dx, dy int) (int, int) {
	for {
		if dy != 0 {
			y += dy
			if y > board.YSize() {
				y = 1
			} else if y < 1 {
				y = board.YSize()
			}
		}

		if dx != 0 {
			x += dx
			if x > board.XSize(y) {
				x = 1
			} else if x < 1 {
				x = board.XSize(y)
			}
		}

		if board.At(x, y) != TileVoid {
			break
		}
	}
	return x, y
}

func (board *Board) Walk(pos Pos, steps int) Pos {
	dx, dy := pos.Dir.DxDy()
	for i := 0; i < steps; i++ {
		x, y := board.Shift(pos.X, pos.Y, dx, dy)
		if board.At(x, y) == TileWall {
			return pos
		}
		pos.X = x
		pos.Y = y
	}
	return pos
}

func (board *Board) Print(pos Pos, window int) (result string) {
	y0 := 1
	y1 := board.YSize()
	if window != 0 {
		y0 = pos.Y - window
		y1 = pos.Y + window
	}
	for y := y0; y <= y1; y++ {
		for x := 1; x <= board.XSize(y); x++ {
			if x == pos.X && y == pos.Y {
				fmt.Print(pos.Dir.String())
			} else {
				fmt.Print(board.At(x, y).String())
			}
		}
		fmt.Println()
	}
	return
}

func (pos Pos) ApplyRotation(rot Rotation) Pos {
	var d int
	switch rot {
	case RotNone:
		return pos
	case RotRight:
		d = 1
	case RotLeft:
		d = 3 // 3 == -1 (mod 4)
	default:
		panic("Invalid rotation")
	}

	pos.Dir = Direction((int(pos.Dir) + d) % 4)
	return pos
}

func (pos Pos) ApplyPath(board *Board, path []Instruction) Pos {
	for _, ins := range path {
		pos = pos.ApplyRotation(ins.Rotate)
		pos = board.Walk(pos, ins.Steps)
		// fmt.Println(ins, pos)
		// board.Print(pos, 0)
		// fmt.Println()
	}
	return pos
}

func ParseInt(s string) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return result
}

func ParseBoardRow(row string) (result []Tile) {
	result = make([]Tile, len(row))
	for i := 0; i < len(row); i++ {
		var tile Tile
		switch row[i] {
		case ' ':
			tile = TileVoid
		case '.':
			tile = TileOpen
		case '#':
			tile = TileWall
		default:
			panic("Invalid tile")
		}
		result[i] = tile
	}
	return
}

func ParsePath(line string) (result []Instruction) {
	result = []Instruction{}

	rot := RotNone

	for {
		rotIdx := strings.IndexAny(line, "LR")
		var steps int
		if rotIdx != -1 {
			steps = ParseInt(line[:rotIdx])
			ins := Instruction{Rotate: rot, Steps: steps}
			result = append(result, ins)

			switch line[rotIdx] {
			case 'L':
				rot = RotLeft
			case 'R':
				rot = RotRight
			default:
				panic("Invalid rotation")
			}
			line = line[rotIdx+1:]
		} else {
			steps = ParseInt(line)
			ins := Instruction{Rotate: rot, Steps: steps}
			result = append(result, ins)
			return
		}
	}
}

func ParseInput(scanner *bufio.Scanner) (board Board, path []Instruction) {
	board = MakeBoard()
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n")
		if line == "" {
			break
		}

		board.AppendRow(ParseBoardRow(line))
	}
	if !scanner.Scan() {
		panic("Empty line expected after board")
	}
	line := strings.TrimSpace(scanner.Text())
	path = ParsePath(line)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			panic("Extra lines after path")
		}
	}
	return
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	board, path := ParseInput(scanner)

	pos := board.StartingPosition()
	pos = pos.ApplyPath(&board, path)

	password := 1000 * pos.Y
	password += 4 * pos.X
	password += int(pos.Dir)
	fmt.Println(password)
}
