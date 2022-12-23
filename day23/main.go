package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Direction int

const (
	DirNorth Direction = 0
	DirSouth Direction = 1
	DirWest  Direction = 2
	DirEast  Direction = 3
)

const NumDirections = 4

type Pos struct {
	X, Y int
}

type Board struct {
	minX, minY int
	maxX, maxY int
	elves      map[Pos]bool
	startDir   Direction
}

func (dir Direction) DxDy() (int, int) {
	switch dir {
	case DirNorth:
		return 0, -1
	case DirSouth:
		return 0, 1
	case DirWest:
		return -1, 0
	case DirEast:
		return 1, 0
	}
	panic("Invalid direction")
}

func (dir Direction) Next() Direction {
	return Direction((int(dir) + 1) % NumDirections)
}

func (dir Direction) ForEach(cb func(Direction) bool) {
	d := dir
	for {
		if cb(d) {
			return
		}
		d = d.Next()
		if d == dir {
			return
		}
	}
}

func (pos Pos) Add(dx, dy int) Pos {
	pos.X += dx
	pos.Y += dy
	return pos
}

func (pos Pos) Move(dir Direction) Pos {
	dx, dy := dir.DxDy()
	return pos.Add(dx, dy)
}

func (pos Pos) Adjacent(dir Direction, cb func(Pos)) {
	dx, dy := dir.DxDy()
	if dx == 0 {
		cb(pos.Add(-1, dy))
		cb(pos.Add(0, dy))
		cb(pos.Add(1, dy))
	} else if dy == 0 {
		cb(pos.Add(dx, -1))
		cb(pos.Add(dx, 0))
		cb(pos.Add(dx, 1))
	}
}

func (pos Pos) AllAdjacent(cb func(Pos)) {
	for _, dx := range [...]int{-1, 0, 1} {
		for _, dy := range [...]int{-1, 0, 1} {
			if dx == 0 && dy == 0 {
				continue
			}
			cb(pos.Add(dx, dy))
		}
	}
}

func MakeBoard() (result Board) {
	result.elves = map[Pos]bool{}
	return
}

func (board *Board) recalcSize() {
	first := true
	for pos, _ := range board.elves {
		if first {
			board.minX = pos.X
			board.maxX = pos.X
			board.minY = pos.Y
			board.maxY = pos.Y
			first = false
			continue
		}

		if pos.X < board.minX {
			board.minX = pos.X
		}

		if pos.X > board.maxX {
			board.maxX = pos.X
		}

		if pos.Y < board.minY {
			board.minY = pos.Y
		}

		if pos.Y > board.maxY {
			board.maxY = pos.Y
		}
	}
}

func (board *Board) AddElf(pos Pos) {
	_, exists := board.elves[pos]
	if exists {
		panic("Adding one elf on top of another")
	}
	board.elves[pos] = true
	board.recalcSize()
}

func (board *Board) HasElf(pos Pos) bool {
	_, exists := board.elves[pos]
	return exists
}

func (board *Board) MoveElf(from, to Pos) {
	if !board.HasElf(from) {
		panic("No elf to move")
	}
	if board.HasElf(to) {
		panic("Destination already occupied")
	}

	delete(board.elves, from)
	board.elves[to] = true
	board.recalcSize()
}

func (board *Board) ElfCount() int {
	return len(board.elves)
}

func (board *Board) XSize() int {
	return board.maxX - board.minX + 1
}

func (board *Board) YSize() int {
	return board.maxY - board.minY + 1
}

func (board *Board) ForEachElf(cb func(Pos)) {
	for pos, _ := range board.elves {
		cb(pos)
	}
}

func (board *Board) RunRound() (moved bool) {
	destinations := map[Pos]Pos{}
	totals := map[Pos]int{}

	board.ForEachElf(func(pos Pos) {
		hasNeighbours := false
		pos.AllAdjacent(func(apos Pos) {
			if board.HasElf(apos) {
				hasNeighbours = true
			}
		})

		if !hasNeighbours {
			return
		}

		board.startDir.ForEach(func(dir Direction) bool {
			canMove := true
			pos.Adjacent(dir, func(apos Pos) {
				if board.HasElf(apos) {
					canMove = false
				}
			})

			if canMove {
				dst := pos.Move(dir)
				// fmt.Printf("%v --> %v\n", pos, dst)
				destinations[pos] = dst
				totals[dst]++
				return true
			}

			return false
		})
	})

	for from, to := range destinations {
		if totals[to] != 1 {
			continue
		}
		board.MoveElf(from, to)
		moved = true
	}

	board.startDir = board.startDir.Next()
	return
}

func (board *Board) Print() {
	for y := board.minY; y <= board.maxY; y++ {
		for x := board.minX; x <= board.maxX; x++ {
			c := "."
			if board.HasElf(Pos{x, y}) {
				c = "#"
			}
			fmt.Print(c)
		}
		fmt.Println()
	}
}

func ParseRow(board *Board, row string) {
	return
}

func ParseInput(scanner *bufio.Scanner) (board Board) {
	board = MakeBoard()

	y := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		ParseRow(&board, line)

		for x := 0; x < len(line); x++ {
			switch line[x] {
			case '.':
				continue
			case '#':
				board.AddElf(Pos{x, y})
			default:
				panic("Invalid cell")
			}
		}

		y++
	}
	return
}

func main() {
	mode2 := false
	if (len(os.Args) > 1) && (os.Args[1] == "2") {
		mode2 = true
	}

	scanner := bufio.NewScanner(os.Stdin)

	board := ParseInput(scanner)

	if !mode2 {
		for i := 0; i < 10; i++ {
			board.RunRound()
		}

		cells := board.XSize() * board.YSize()
		empty := cells - board.ElfCount()

		fmt.Println(empty)
	} else {
		round := 1
		for board.RunRound() {
			round++
		}

		fmt.Println(round)
	}
}
