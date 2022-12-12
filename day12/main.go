package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

type Height byte

type Pos struct {
	X, Y int
}

type HeightMap struct {
	grid  [][]Height
	xsize int
}

type DistanceNode struct {
	Calculated bool
	Traversed  bool
	Distance   int
}

type DistanceMap struct {
	hmap         *HeightMap
	nodes        []DistanceNode
	xsize, ysize int
	next         mapset.Set[Pos]
}

func CharToHeight(c byte) Height {
	return Height(c)
}

func EligibleMove(from, to Height) bool {
	return to <= from+1
}

func MakeHeightMap() (result HeightMap) {
	result.grid = make([][]Height, 0)
	return
}

func (hmap *HeightMap) AppendRow(row []Height) int {
	if hmap.xsize == 0 {
		hmap.xsize = len(row)
	} else if hmap.xsize != len(row) {
		panic("Row size mismatch")
	}
	hmap.grid = append(hmap.grid, row)
	return len(hmap.grid)
}

func (hmap *HeightMap) XSize() int {
	return hmap.xsize
}

func (hmap *HeightMap) YSize() int {
	return len(hmap.grid)
}

func (hmap *HeightMap) At(pos Pos) Height {
	return hmap.grid[pos.Y][pos.X]
}

func ParseHeightMap(scanner *bufio.Scanner) (hmap HeightMap, startPos, endPos Pos) {
	hmap = MakeHeightMap()
	y := 0

	startFlag := false
	endFlag := false

	setPos := func(pos *Pos, flag *bool, x, y int) {
		if *flag {
			panic("Position already set")
		}
		pos.X = x
		pos.Y = y
		*flag = true
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		row := make([]Height, len(line))
		for x := 0; x < len(line); x++ {
			c := line[x]
			if c == 'S' {
				setPos(&startPos, &startFlag, x, y)
				c = 'a'
			} else if c == 'E' {
				setPos(&endPos, &endFlag, x, y)
				c = 'z'
			}
			row[x] = CharToHeight(c)
		}

		y = hmap.AppendRow(row)
	}

	return
}

func (hmap *HeightMap) TraverseEligibleDestinations(from Pos, cb func(Pos)) {
	checkDst := func(x, y int) {
		if x < 0 || x >= hmap.XSize() {
			return
		}
		if y < 0 || y >= hmap.YSize() {
			return
		}
		to := Pos{x, y}
		if EligibleMove(hmap.At(from), hmap.At(to)) {
			cb(to)
		}
	}

	checkDst(from.X-1, from.Y)
	checkDst(from.X+1, from.Y)
	checkDst(from.X, from.Y-1)
	checkDst(from.X, from.Y+1)
}

func MakeDistanceMap(hmap *HeightMap, start Pos) (result DistanceMap) {
	result.hmap = hmap
	result.xsize = hmap.XSize()
	result.ysize = hmap.YSize()
	result.nodes = make([]DistanceNode, result.xsize*result.ysize)
	result.next = mapset.NewThreadUnsafeSet[Pos]()

	result.At(start).Calculated = true
	result.next.Add(start)
	return
}

func (dmap *DistanceMap) XSize() int {
	return dmap.xsize
}

func (dmap *DistanceMap) YSize() int {
	return dmap.ysize
}

func (dmap *DistanceMap) At(pos Pos) *DistanceNode {
	return &dmap.nodes[pos.Y*dmap.xsize+pos.X]
}

func (dmap *DistanceMap) Propagate() bool {
	if dmap.next.Cardinality() == 0 {
		return false
	}

	next := mapset.NewThreadUnsafeSet[Pos]()
	dmap.next.Each(func(from Pos) bool {
		fromNode := dmap.At(from)
		if fromNode.Traversed {
			return false
		}
		//fmt.Printf("from %v\n", from)
		dmap.hmap.TraverseEligibleDestinations(from, func(to Pos) {
			//fmt.Printf("%v -> %v\n", from, to)
			next.Add(to)
			toNode := dmap.At(to)
			if !toNode.Calculated {
				toNode.Calculated = true
				toNode.Distance = fromNode.Distance + 1
				//fmt.Printf("%v: dist %d\n", to, toNode.Distance)
			}
		})
		fromNode.Traversed = true
		return false
	})
	dmap.next = next
	return true
}

func (dmap *DistanceMap) CalcDistanceTo(target Pos) int {
	targetNode := dmap.At(target)
	for dmap.Propagate() {
		if targetNode.Calculated {
			return targetNode.Distance
		}
	}

	panic("No path was found")
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	hmap, startPos, endPos := ParseHeightMap(scanner)
	dmap := MakeDistanceMap(&hmap, startPos)
	dist := dmap.CalcDistanceTo(endPos)
	fmt.Println(dist)
}
