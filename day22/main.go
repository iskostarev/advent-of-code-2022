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
	RotNone     Rotation = 0
	RotCW       Rotation = 1
	RotOpposite Rotation = 2
	RotCCW      Rotation = 3
)

type Tile int

const (
	TileVoid = iota
	TileOpen
	TileWall
)

const NumDirections = 4

type Direction int

const (
	DirRight Direction = 0
	DirDown  Direction = 1
	DirLeft  Direction = 2
	DirUp    Direction = 3
)

type Coords struct {
	X, Y int
}

type Pos struct {
	X, Y int
	Dir  Direction
}

type CoordRange struct {
	X, Y, MinX, MinY, MaxX, MaxY int
	Reversed                     bool
}

type CubeEdgeBindings struct {
	Bound      bool
	MinX, MinY int
	Size       int
}

type CubeEdgeConnection struct {
	Invert bool
	Side   Direction
	Edge   *CubeEdge
}

type CubeEdge struct {
	Bindings    CubeEdgeBindings
	Orientation Rotation
	Connected   [NumDirections]CubeEdgeConnection
	Traversed   bool
}

type Board struct {
	rows     [][]Tile
	topology map[Pos]Pos
	maxX     int
}

type Instruction struct {
	Rotate Rotation
	Steps  int
}

const NumCubeEdges = 6

func (rot Rotation) String() string {
	switch rot {
	case RotNone:
		return "_"
	case RotCCW:
		return "L"
	case RotOpposite:
		return "U"
	case RotCW:
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

func (dir Direction) Opposite() Direction {
	return Direction((int(dir) + 2) % NumDirections)
}

func (dir Direction) Rotate(rot Rotation) Direction {
	return Direction((int(dir) + int(rot)) % NumDirections)
}

func AddRotations(lhs, rhs Rotation) Rotation {
	return Rotation(int(lhs+rhs) % NumDirections)
}

func DiffRotations(lhs, rhs Rotation) Rotation {
	return Rotation(int(lhs-rhs+NumDirections) % NumDirections)
}

func (rot Rotation) Reverse() Rotation {
	if rot == RotCW {
		return RotCCW
	} else if rot == RotCCW {
		return RotCW
	} else {
		return rot
	}
}

func RequiredRotation(target, from Direction) Rotation {
	return Rotation(int(target-from+NumDirections) % NumDirections)
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

func (r *CoordRange) Empty() bool {
	return r.X > r.MaxX || r.Y > r.MaxY || r.X < r.MinX || r.Y < r.MinY
}

func (r *CoordRange) Reset() {
	if r.Reversed {
		r.X = r.MaxX
		r.Y = r.MaxY
	} else {
		r.X = r.MinX
		r.Y = r.MinY
	}
}

func (r *CoordRange) Forward() bool {
	r.X++
	if r.X > r.MaxX {
		r.X = r.MinX
		r.Y++
		if r.Y > r.MaxY {
			return false
		}
	}
	return true
}

func (r *CoordRange) Backward() bool {
	r.X--
	if r.X < r.MinX {
		r.X = r.MaxX
		r.Y--
		if r.Y < r.MinY {
			return false
		}
	}
	return true
}

func (r *CoordRange) Next() bool {
	if r.Reversed {
		return r.Backward()
	} else {
		return r.Forward()
	}
}

func (r CoordRange) Reverse() CoordRange {
	r.Reversed = !r.Reversed
	r.Reset()
	return r
}

func (r CoordRange) Traverse(cb func(int, int)) {
	if r.Empty() {
		return
	}

	cb(r.X, r.Y)
	for r.Next() {
		cb(r.X, r.Y)
	}
}

func TraverseRangePair(r1, r2 CoordRange, cb func(int, int, int, int)) {
	if r1.Empty() && r2.Empty() {
		return
	}

	for {
		if r1.Empty() || r2.Empty() {
			panic("Range length mismatch")
		}

		cb(r1.X, r1.Y, r2.X, r2.Y)
		r1.Next()
		r2.Next()
		if r1.Empty() && r2.Empty() {
			return
		}
	}
}

func (bindings CubeEdgeBindings) SideRange(side Direction) CoordRange {
	var minX, minY, maxX, maxY int

	switch side {
	case DirRight:
		minX = bindings.MinX + bindings.Size - 1
		maxX = minX
		minY = bindings.MinY
		maxY = bindings.MinY + bindings.Size - 1
	case DirDown:
		minX = bindings.MinX
		maxX = bindings.MinX + bindings.Size - 1
		minY = bindings.MinY + bindings.Size - 1
		maxY = minY
	case DirLeft:
		minX = bindings.MinX
		maxX = minX
		minY = bindings.MinY
		maxY = bindings.MinY + bindings.Size - 1
	case DirUp:
		minX = bindings.MinX
		maxX = bindings.MinX + bindings.Size - 1
		minY = bindings.MinY
		maxY = minY
	default:
		panic("Invalid direction")
	}

	result := CoordRange{
		MinX: minX,
		MinY: minY,
		MaxX: maxX,
		MaxY: maxY,
	}
	result.Reset()
	return result
}

func (edge *CubeEdge) SideRange(dir Direction) CoordRange {
	result := edge.Bindings.SideRange(dir.Rotate(edge.Orientation.Reverse()))
	switch edge.Orientation {
	case RotOpposite:
		result = result.Reverse()
	case RotCW:
		if dir == DirDown || dir == DirUp {
			result = result.Reverse()
		}
	case RotCCW:
		if dir == DirLeft || dir == DirRight {
			result = result.Reverse()
		}
	}
	return result
}

func MakeBoard() (result Board) {
	result.rows = [][]Tile{}
	result.topology = map[Pos]Pos{}
	return
}

func (board *Board) AppendRow(row []Tile) {
	board.rows = append(board.rows, row)
	if len(row) > board.maxX {
		board.maxX = len(row)
	}
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

func (board *Board) ShiftPlanar(x, y, dx, dy int) (int, int) {
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

func (board *Board) CalculateWrapAroundTopology() {
	for y := 1; y <= board.YSize(); y++ {
		for x := 1; x <= board.XSize(y); x++ {
			for _, dir := range [...]Direction{DirRight, DirDown, DirLeft, DirUp} {
				from := Pos{x, y, dir}
				to := from
				dx, dy := dir.DxDy()
				to.X, to.Y = board.ShiftPlanar(x, y, dx, dy)
				board.topology[from] = to
			}
		}
	}
}

func (board *Board) calculateInnerTopology() {
	for y := 1; y <= board.YSize(); y++ {
		for x := 1; x <= board.XSize(y); x++ {
			for _, dir := range [...]Direction{DirRight, DirDown, DirLeft, DirUp} {
				from := Pos{x, y, dir}
				to := from
				dx, dy := dir.DxDy()
				to.X, to.Y = x+dx, y+dy
				if board.At(to.X, to.Y) != TileVoid {
					board.topology[from] = to
				}
			}
		}
	}
}

func (board *Board) FindSolidHorizontalLine(y int) (min, max int) {
	xSize := board.XSize(y)
	for min = 0; min <= xSize; min++ {
		if board.At(min, y) != TileVoid {
			break
		}
	}
	if board.At(min, y) == TileVoid {
		panic("Unexpected empty line")
	}

	for max = xSize; max >= 0; max-- {
		if board.At(max, y) != TileVoid {
			break
		}
	}
	if board.At(max, y) == TileVoid {
		panic("Unexpected empty line")
	}

	return
}

func (board *Board) FindSolidVerticalLine(x int) (min, max int) {
	ySize := board.YSize()
	for min = 0; min <= ySize; min++ {
		if board.At(x, min) != TileVoid {
			break
		}
	}
	if board.At(x, min) == TileVoid {
		panic("Unexpected empty line")
	}

	for max = ySize; max >= 0; max-- {
		if board.At(x, max) != TileVoid {
			break
		}
	}
	if board.At(x, max) == TileVoid {
		panic("Unexpected empty line")
	}

	return
}

func gcd(x, y int) int {
	for y != 0 {
		x, y = y, x%y
	}
	return x
}

func (board *Board) calculateCubeEdgeSize() (result int) {
	result = 0
	for y := 1; y <= board.YSize(); y++ {
		y0, y1 := board.FindSolidHorizontalLine(y)
		size := y1 - y0 + 1
		if result == 0 {
			result = size
		} else {
			result = gcd(result, size)
		}
	}
	for x := 1; x <= board.maxX; x++ {
		x0, x1 := board.FindSolidVerticalLine(x)
		size := x1 - x0 + 1
		result = gcd(result, size)
	}
	return
}

func (board *Board) allVoid(minX, maxX, minY, maxY int) (result bool) {
	first := true
	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			if first {
				result = board.At(x, y) == TileVoid
				continue
			}
			if result != (board.At(x, y) == TileVoid) {
				panic("Invalid cube")
			}
		}
	}
	return
}

func (board *Board) locateCubeEdges() (result map[Coords]*CubeEdge, root *CubeEdge) {
	edgeSize := board.calculateCubeEdgeSize()
	if board.YSize()%edgeSize != 0 || board.maxX%edgeSize != 0 {
		panic("Invalid cube")
	}

	xcount := board.maxX / edgeSize
	ycount := board.YSize() / edgeSize

	result = map[Coords]*CubeEdge{}

	for i := 0; i < xcount; i++ {
		for j := 0; j < ycount; j++ {
			bindings := CubeEdgeBindings{
				Bound: true,
				MinX:  1 + i*edgeSize,
				MinY:  1 + j*edgeSize,
				Size:  edgeSize,
			}
			if !board.allVoid(bindings.MinX, bindings.MinX+bindings.Size-1, bindings.MinY, bindings.MinY+bindings.Size-1) {
				edge := CubeEdge{Bindings: bindings}
				result[Coords{i, j}] = &edge
				root = &edge
			}
		}
	}

	return
}

func buildReferenceCube() *CubeEdge {
	var top, bottom CubeEdge
	sides := [NumDirections]CubeEdge{}

	topdir := DirUp
	bottomdir := DirDown
	for i := 0; i < NumDirections; i++ {
		right := (i + 1) % NumDirections
		left := (i + NumDirections - 1) % NumDirections

		dir := Direction(i)
		opp := dir.Opposite()

		sides[i].Connected[DirLeft].Edge = &sides[left]
		sides[i].Connected[DirLeft].Side = DirRight
		sides[i].Connected[DirRight].Edge = &sides[right]
		sides[i].Connected[DirRight].Side = DirLeft

		sides[i].Connected[DirUp].Edge = &top
		sides[i].Connected[DirUp].Side = topdir
		sides[i].Connected[DirDown].Edge = &bottom
		sides[i].Connected[DirDown].Side = bottomdir

		top.Connected[topdir.Opposite()].Edge = &sides[opp]
		top.Connected[i].Side = DirUp

		bottom.Connected[bottomdir.Opposite()].Edge = &sides[opp]
		bottom.Connected[i].Side = DirDown

		topdir = topdir.Rotate(RotCCW)
		bottomdir = bottomdir.Rotate(RotCW)

		sides[i].Bindings.MinX = i + 1 // debug value
	}

	top.Bindings.MinY = 2
	bottom.Bindings.MinY = 1

	sides[0].Connected[DirUp].Invert = true
	sides[0].Connected[DirDown].Invert = true
	sides[1].Connected[DirDown].Invert = true
	sides[3].Connected[DirUp].Invert = true
	top.Connected[DirUp].Invert = true
	top.Connected[DirRight].Invert = true
	bottom.Connected[DirDown].Invert = true
	bottom.Connected[DirLeft].Invert = true

	return &sides[0]
}

func connectAdjacent(edges *map[Coords]*CubeEdge) {
	for key, edge := range *edges {
		for _, dir := range [...]Direction{DirRight, DirDown, DirLeft, DirUp} {
			dx, dy := dir.DxDy()
			c := Coords{key.X + dx, key.Y + dy}
			conn, ok := (*edges)[c]
			if ok {
				edge.Connected[dir].Edge = conn
				edge.Connected[dir].Side = dir.Opposite()
			}
		}
	}

}

func doCopyEdges(src, dst *CubeEdge, rotation Rotation, traversed *map[*CubeEdge]bool) {
	if src == nil || dst == nil {
		panic("Traversal error")
	}
	if (*traversed)[src] {
		return
	}
	dst.Bindings = src.Bindings
	dst.Orientation = rotation
	(*traversed)[src] = true
	for _, dir := range [...]Direction{DirRight, DirDown, DirLeft, DirUp} {
		if src.Connected[dir].Edge != nil {
			srcConn := src.Connected[dir]
			dstConn := dst.Connected[dir.Rotate(rotation)]
			rot := RequiredRotation(dstConn.Side, srcConn.Side)
			doCopyEdges(srcConn.Edge, dstConn.Edge, rot, traversed)
		}
	}
}

func copyEdges(src, dst *CubeEdge) {
	traversed := map[*CubeEdge]bool{}
	doCopyEdges(src, dst, RotNone, &traversed)
	if len(traversed) != NumCubeEdges {
		panic("Invalid cube")
	}
}

func doTraverseEdges(edge *CubeEdge, edgeNums *map[*CubeEdge]int, counter *int) {
	_, ok := (*edgeNums)[edge]
	if ok || edge == nil {
		return
	}
	num := (*counter)
	(*counter)++
	(*edgeNums)[edge] = num
	for _, dir := range [...]Direction{DirRight, DirDown, DirLeft, DirUp} {
		doTraverseEdges(edge.Connected[dir].Edge, edgeNums, counter)
	}
}

func traverseEdgesFullInfo(edge *CubeEdge, cb func(int, *CubeEdge, *map[*CubeEdge]int)) {
	edgeNums := map[*CubeEdge]int{}
	revEdges := map[int]*CubeEdge{}
	edgeCount := 0

	doTraverseEdges(edge, &edgeNums, &edgeCount)

	for edge, num := range edgeNums {
		revEdges[num] = edge
	}

	for num := 0; num < edgeCount; num++ {
		cb(num, revEdges[num], &edgeNums)
	}
}

func traverseEdges(edge *CubeEdge, cb func(*CubeEdge)) {
	traverseEdgesFullInfo(edge, func(num int, edge *CubeEdge, edgeNums *map[*CubeEdge]int) {
		cb(edge)
	})
}

func PrintEdges(edge *CubeEdge) {
	traverseEdgesFullInfo(edge, func(num int, edge *CubeEdge, edgeNums *map[*CubeEdge]int) {
		fmt.Printf("Edge %d:\n", num)
		fmt.Printf(" Orientation: %v, Bindings: %v\n", edge.Orientation, edge.Bindings)
		for _, dir := range [...]Direction{DirRight, DirDown, DirLeft, DirUp} {
			if edge.Connected[dir].Edge == nil {
				continue
			}

			inv := ""
			if edge.Connected[dir].Invert {
				inv = " (inverted)"
			}
			c := (*edgeNums)[edge.Connected[dir].Edge]
			fmt.Printf(" Connected to %d at %s, side %s%s\n", c, dir, edge.Connected[dir].Side, inv)
		}
	})
}

func assertBound(edge *CubeEdge) {
	traverseEdges(edge, func(edge *CubeEdge) {
		if !edge.Bindings.Bound {
			panic("Unbound edge")
		}
	})
}

func (board *Board) calculateCubeEdgeTopology(edge *CubeEdge) {
	for _, dir := range [...]Direction{DirRight, DirDown, DirLeft, DirUp} {
		curRange := edge.SideRange(dir)

		conn := edge.Connected[dir]
		connRange := conn.Edge.SideRange(conn.Side)

		if conn.Invert {
			connRange = connRange.Reverse()
		}

		odir := dir.Rotate(edge.Orientation.Reverse())
		cdir := conn.Side.Opposite().Rotate(conn.Edge.Orientation.Reverse())

		TraverseRangePair(curRange, connRange, func(x, y, cx, cy int) {
			pos := Pos{x, y, odir}
			cpos := Pos{cx, cy, cdir}
			orig, ok := board.topology[pos]
			if ok && orig != cpos {
				panic(fmt.Sprintf("Mismatch detected at %v: %v != %v", pos, orig, cpos))
			}
			board.topology[pos] = cpos
		})
	}
}

func (board *Board) CalculateCubeTopology() {
	board.calculateInnerTopology()
	edges, root := board.locateCubeEdges()

	if len(edges) != NumCubeEdges {
		panic("Invalid cube")
	}

	connectAdjacent(&edges)

	cube := buildReferenceCube()

	copyEdges(root, cube)
	assertBound(cube)

	traverseEdges(cube, func(edge *CubeEdge) {
		board.calculateCubeEdgeTopology(edge)
	})
}

func (board *Board) Walk(pos Pos, steps int) Pos {
	for i := 0; i < steps; i++ {
		npos, ok := board.topology[pos]
		if !ok {
			panic("Invalid topology")
		}
		if board.At(npos.X, npos.Y) == TileWall {
			return pos
		}
		pos = npos
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

func (pos Pos) ApplyPath(board *Board, path []Instruction) Pos {
	for _, ins := range path {
		pos.Dir = pos.Dir.Rotate(ins.Rotate)
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
				rot = RotCCW
			case 'R':
				rot = RotCW
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
	mode2 := false
	if (len(os.Args) > 1) && (os.Args[1] == "2") {
		mode2 = true
	}

	scanner := bufio.NewScanner(os.Stdin)

	board, path := ParseInput(scanner)

	if !mode2 {
		board.CalculateWrapAroundTopology()
	} else {
		board.CalculateCubeTopology()
	}

	pos := board.StartingPosition()
	pos = pos.ApplyPath(&board, path)

	password := 1000 * pos.Y
	password += 4 * pos.X
	password += int(pos.Dir)
	fmt.Println(password)
}
