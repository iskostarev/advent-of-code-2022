package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type TreeValue byte

type Tree struct {
	Value   TreeValue
	Visible bool
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

func (grid *Grid) Trace(x, y, dx, dy int) {
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
		grid.Trace(i, 0, 0, 1)
		grid.Trace(i, ym, 0, -1)
	}
	for i := 0; i <= ym; i++ {
		grid.Trace(0, i, 1, 0)
		grid.Trace(xm, i, -1, 0)
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

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	grid := Grid{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		grid.AppendRow(ParseRow(line))
	}

	grid.CalcVisibility()
	fmt.Println(grid.CountVisible())
}
