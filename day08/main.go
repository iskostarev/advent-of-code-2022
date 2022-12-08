package main

import (
	"bufio"
	"fmt"
	"os"
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

type TreeValue byte

const MaxTreeValue TreeValue = 9

type CachedViewingDistances [MaxTreeValue + 1]int

type Tree struct {
	Value    TreeValue
	Visible  bool
	VDCaches [DirCount]CachedViewingDistances
}

type Grid struct {
	trees   [][]Tree
	rowsize int
}

func (grid *Grid) AppendRow(row []TreeValue) {
	if grid.rowsize == 0 {
		grid.rowsize = len(row)
	} else if grid.rowsize != len(row) {
		panic("Row size mismatch")
	}

	finalRow := make([]Tree, len(row))
	for i, value := range row {
		finalRow[i].Value = value
	}
	grid.trees = append(grid.trees, finalRow)
}

func (grid *Grid) Width() int {
	return grid.rowsize
}

func (grid *Grid) Height() int {
	return len(grid.trees)
}

func (grid *Grid) At(x, y int) *Tree {
	if x < 0 || x >= grid.Width() {
		panic("X out of bounds")
	}
	if y < 0 || y >= grid.Height() {
		panic("Y out of bounds")
	}
	return &grid.trees[y][x]
}

func (grid *Grid) InBounds(x, y int) bool {
	return x >= 0 && x < grid.Width() && y >= 0 && y < grid.Height()
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

func (grid *Grid) Trace(x, y int, dir Direction) {
	dx, dy := dir.DxDy()
	grid.At(x, y).Visible = true
	highest := grid.At(x, y).Value
	for {
		x += dx
		y += dy

		if !grid.InBounds(x, y) {
			break
		}

		curVal := grid.At(x, y).Value
		if curVal > highest {
			highest = curVal
			grid.At(x, y).Visible = true
		}
	}
}

func (grid *Grid) CalcVisibility() {
	xm := grid.Width() - 1
	ym := grid.Height() - 1
	for i := 0; i <= xm; i++ {
		grid.Trace(i, 0, DirDown)
		grid.Trace(i, ym, DirUp)
	}
	for i := 0; i <= ym; i++ {
		grid.Trace(0, i, DirRight)
		grid.Trace(xm, i, DirLeft)
	}
}

func (grid *Grid) CountVisible() (result int) {
	for i := 0; i < grid.Width(); i++ {
		for j := 0; j < grid.Height(); j++ {
			if grid.At(i, j).Visible {
				result++
			}
		}
	}

	return
}

func (grid *Grid) String() (result string) {
	for i := 0; i < grid.Height(); i++ {
		for j := 0; j < grid.Width(); j++ {
			tree := grid.At(j, i)
			vis := ' '
			if tree.Visible {
				vis = '^'
			}
			result += fmt.Sprintf("[%d%c]", tree.Value, vis)
		}
		result += "\n"
	}
	return
}

func ParseRow(line string) (result []TreeValue) {
	result = make([]TreeValue, 0, len(line))
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c < '0' || c > '9' {
			panic("Invalid digit")
		}
		result = append(result, TreeValue(c-'0'))
	}
	return
}

func (grid *Grid) viewingDistanceForValue(dir Direction, x, y int, value TreeValue) (result int) {
	if value > MaxTreeValue {
		panic("Tree height exceeds max value")
	}
	cache := &grid.At(x, y).VDCaches[dir]
	if (*cache)[value] > 0 {
		return (*cache)[value]
	}

	dx, dy := dir.DxDy()
	x2 := x + dx
	y2 := y + dy
	if !grid.InBounds(x2, y2) {
		return 0
	}

	value2 := grid.At(x2, y2).Value
	if value2 > MaxTreeValue {
		panic("Tree height exceeds max value")
	}

	for i := TreeValue(0); i <= value2; i++ {
		(*cache)[i] = 1
	}

	if value <= value2 {
		return 1
	}
	(*cache)[value] = grid.viewingDistanceForValue(dir, x2, y2, value) + 1
	return (*cache)[value]
}

func (grid *Grid) ViewingDistance(dir Direction, x, y int) (result int) {
	return grid.viewingDistanceForValue(dir, x, y, grid.At(x, y).Value)
}

func (grid *Grid) ScenicScore(x, y int) (result int) {
	result = 1
	for _, dir := range [...]Direction{DirUp, DirDown, DirLeft, DirRight} {
		result *= grid.ViewingDistance(dir, x, y)
	}
	return
}

func (grid *Grid) MaxScenicScore() (result int) {
	for i := 0; i < grid.Height(); i++ {
		for j := 0; j < grid.Width(); j++ {
			ss := grid.ScenicScore(i, j)
			if ss > result {
				result = ss
			}
		}
	}
	return
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	grid := Grid{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		grid.AppendRow(ParseRow(line))
	}

	if (len(os.Args) > 1) && (os.Args[1] == "2") {
		fmt.Println(grid.MaxScenicScore())
	} else {
		grid.CalcVisibility()
		fmt.Println(grid.CountVisible())
	}
}
