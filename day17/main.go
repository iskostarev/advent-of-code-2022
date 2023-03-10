package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
)

const TowerWidth = 7

type Direction int

const (
	DirLeft Direction = iota
	DirRight
)

type JetPattern struct {
	pattern []Direction
	pos     int
}

type RockSprite struct {
	Height, Width int
	Mask          [][]bool
}

type FallingRock struct {
	Sprite    *RockSprite
	LeftPos   int
	BottomPos int
}

type RockGenerator struct {
	sprites []RockSprite
	pos     int
}

type TowerStateKey struct {
	blocks             Bitstring
	jetPos, rockGenPos int
}

type TowerStateInfo struct {
	Turn, Height int
}

type Tower struct {
	filled     [][]bool
	falling    FallingRock
	jetPattern JetPattern
	rockGen    RockGenerator
	culled     int
	turn       int

	prevStates map[TowerStateKey]TowerStateInfo
	loopBegin  int
	loopLength int
	loopHeight int
}

func (dir Direction) String() string {
	switch dir {
	case DirLeft:
		return "<"
	case DirRight:
		return ">"
	}
	panic("Invalid direction")
}

func (jp *JetPattern) Next() (result Direction) {
	result = jp.pattern[jp.pos]
	jp.pos++
	if jp.pos == len(jp.pattern) {
		jp.pos = 0
	}
	return
}

func (jp *JetPattern) Cur() Direction {
	return jp.pattern[jp.pos]
}

func (jp JetPattern) String() (result string) {
	result = "("
	for i, dir := range jp.pattern {
		if i == jp.pos {
			result += fmt.Sprintf("[%v]", dir)
		} else {
			result += dir.String()
		}
	}
	result += ")"
	return
}

func (jp JetPattern) StringNextUp() (result string) {
	result = "("
	for i := 0; i < 20; i++ {
		p := (jp.pos + i) % len(jp.pattern)
		result += jp.pattern[p].String()
	}
	result += "...)"
	return
}

func ParseJetPattern(reader *bufio.Reader) (result JetPattern) {
	result.pattern = []Direction{}
	for {
		c, err := reader.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		var dir Direction

		switch c {
		case ' ', '\n':
			continue
		case '<':
			dir = DirLeft
		case '>':
			dir = DirRight
		default:
			panic("Unexpected char")
		}

		result.pattern = append(result.pattern, dir)
	}

	return
}

func MakeRockGenerator() (result RockGenerator) {
	result.sprites = []RockSprite{
		RockSprite{
			Height: 1,
			Width:  4,
			Mask: [][]bool{
				[]bool{true, true, true, true},
			},
		},
		RockSprite{
			Height: 3,
			Width:  3,
			Mask: [][]bool{
				[]bool{false, true, false},
				[]bool{true, true, true},
				[]bool{false, true, false},
			},
		},
		RockSprite{
			Height: 3,
			Width:  3,
			Mask: [][]bool{
				[]bool{true, true, true},
				[]bool{false, false, true},
				[]bool{false, false, true},
			},
		},
		RockSprite{
			Height: 4,
			Width:  1,
			Mask: [][]bool{
				[]bool{true},
				[]bool{true},
				[]bool{true},
				[]bool{true},
			},
		},
		RockSprite{
			Height: 2,
			Width:  2,
			Mask: [][]bool{
				[]bool{true, true},
				[]bool{true, true},
			},
		},
	}
	return
}

func (rockGen *RockGenerator) Next() (result *RockSprite) {
	result = &rockGen.sprites[rockGen.pos]
	rockGen.pos++
	if rockGen.pos == len(rockGen.sprites) {
		rockGen.pos = 0
	}
	return
}

func (sprite *RockSprite) At(x, y int) bool {
	if x < 0 || x >= sprite.Width {
		return false
	}
	if y < 0 || y >= sprite.Height {
		return false
	}
	return sprite.Mask[y][x]
}

func (falling *FallingRock) At(x, y int) bool {
	if falling.Sprite == nil {
		return false
	}
	return falling.Sprite.At(x-falling.LeftPos, y-falling.BottomPos)
}

func MakeTower(jp JetPattern) (result Tower) {
	result.filled = [][]bool{}
	result.jetPattern = jp
	result.rockGen = MakeRockGenerator()
	result.prevStates = map[TowerStateKey]TowerStateInfo{}
	return
}

func (tower *Tower) At(x, y int) bool {
	if x < 0 || x >= TowerWidth {
		panic("x out of bounds")
	}
	if y < 0 {
		return true
	}
	if y >= len(tower.filled) {
		return false
	}
	return tower.filled[y][x]
}

func (tower *Tower) set(x, y int) {
	if x < 0 || x >= TowerWidth {
		panic("x out of bounds")
	}
	if y > len(tower.filled) {
		panic("setting in air?")
	} else if y == len(tower.filled) {
		tower.filled = append(tower.filled, make([]bool, TowerWidth))
	} else if y < 0 {
		panic("setting underground?")
	}
	tower.filled[y][x] = true
}

func (tower *Tower) height() int {
	return len(tower.filled)
}

func (tower *Tower) TotalHeight() int {
	return len(tower.filled) + tower.culled
}

func (tower *Tower) blockingBottom(x, y int) bool {
	if !tower.At(x, y) {
		return false
	}

	for _, nx := range [...]int{x - 1, x + 1} {
		if nx < 0 || nx >= TowerWidth {
			continue
		}
		blocking := false
		for _, ny := range [...]int{y - 1, y, y + 1} {
			if tower.At(nx, ny) {
				blocking = true
				break
			}
		}

		if !blocking {
			return false
		}
	}

	return true
}

func (tower *Tower) cull() {
	if tower.height() == 0 {
		return
	}

	minHeight := tower.height() - 1
	for x := 0; x < TowerWidth; x++ {
		curHeight := -1
		for y := tower.height() - 1; y >= 0; y-- {
			if tower.blockingBottom(x, y) {
				curHeight = y
				break
			}
		}

		if curHeight < minHeight {
			minHeight = curHeight
		}
	}

	if minHeight > 0 {
		tower.culled += minHeight
		tower.filled = tower.filled[minHeight:]
	}
}

func (falling *FallingRock) traverseFallingRock(cb func(int, int)) {
	if falling.Sprite == nil {
		return
	}

	for y := 0; y < falling.Sprite.Height; y++ {
		for x := 0; x < falling.Sprite.Width; x++ {
			if falling.Sprite.At(x, y) {
				cb(falling.LeftPos+x, falling.BottomPos+y)
			}
		}
	}
}

func (falling FallingRock) shift(dir Direction) (result FallingRock) {
	result = falling
	if dir == DirLeft {
		result.LeftPos--
		if result.LeftPos < 0 {
			return falling
		}
	} else if dir == DirRight {
		result.LeftPos++
		if result.LeftPos+result.Sprite.Width > TowerWidth {
			return falling
		}
	} else {
		panic("Invalid direction")
	}

	return
}

func (tower *Tower) wouldConnect(falling FallingRock) (result bool) {
	if falling.Sprite == nil {
		panic("Impossible condition")
	}

	falling.traverseFallingRock(func(x, y int) {
		if !result && y < 0 && tower.culled > 0 {
			panic("Accesssing culled rows")
		}
		if tower.At(x, y) {
			result = true
		}
	})

	return
}

func (tower *Tower) dumpState() (key TowerStateKey, info TowerStateInfo) {
	key.jetPos = tower.jetPattern.pos
	key.rockGenPos = tower.rockGen.pos
	key.blocks = MakeBitstring(len(tower.filled) * TowerWidth)
	for y := 0; y < len(tower.filled); y++ {
		for x := 0; x < TowerWidth; x++ {
			key.blocks.Set(y*TowerWidth+x, tower.At(x, y))
		}
	}

	info.Turn = tower.turn
	info.Height = tower.TotalHeight()
	return
}

func (tower *Tower) detectLoop() {
	if tower.loopLength != 0 {
		return
	}

	stateKey, stateInfo := tower.dumpState()
	prevInfo, ok := tower.prevStates[stateKey]
	if ok {
		tower.loopBegin = prevInfo.Turn
		tower.loopLength = stateInfo.Turn - prevInfo.Turn
		tower.loopHeight = stateInfo.Height - prevInfo.Height
		return
	}
	tower.prevStates[stateKey] = stateInfo
}

func (tower *Tower) looped() bool {
	if tower.loopLength == 0 {
		return false
	}

	return (tower.turn-tower.loopBegin)%tower.loopLength == 0
}

func (tower *Tower) SkipLooped(maxSteps int) int {
	if !tower.looped() {
		return 0
	}

	skippedTurns := maxSteps - maxSteps%tower.loopLength
	skippedLoops := skippedTurns / tower.loopLength
	tower.turn += skippedTurns
	tower.culled += skippedLoops * tower.loopHeight

	return skippedTurns
}

func (tower *Tower) DropRock(debug bool) {
	tower.falling = FallingRock{
		LeftPos:   2,
		BottomPos: tower.height() + 3,
		Sprite:    tower.rockGen.Next(),
	}

	debugPrint := func(caption string) {
		if debug {
			if caption != "" {
				fmt.Println(caption)
			}
			tower.Print()
			fmt.Println()
			fmt.Println()
		}
	}

	debugPrint("")

	for {
		jnext := tower.jetPattern.Next()
		shifted := tower.falling.shift(jnext)
		if !tower.wouldConnect(shifted) {
			tower.falling = shifted
		}

		debugPrint(fmt.Sprintf(":: Shifting %v", jnext))

		dropped := tower.falling
		dropped.BottomPos--
		if tower.wouldConnect(dropped) {
			debugPrint(":: Setting")
			tower.falling.traverseFallingRock(func(x, y int) {
				tower.set(x, y)
			})
			tower.falling.Sprite = nil
			tower.cull()
			tower.turn++
			tower.detectLoop()
			return
		}
		tower.falling = dropped

		debugPrint(":: Dropping")
	}
}

func (tower *Tower) DropRocks(steps int, debug bool) {
	for i := 0; i < steps; i++ {
		skipped := tower.SkipLooped(steps - i - 1)
		i += skipped
		if i > steps {
			panic("Skipped too much")
		}
		if debug {
			fmt.Printf(":: Rock %d: %s\n", i, tower.jetPattern.StringNextUp())
		}
		tower.DropRock(debug)
		if debug {
			tower.Print()
			fmt.Println()
			fmt.Println()
		}
	}
}

func (tower *Tower) PrintWithHeight(height int) {
	to := tower.height() + 6
	from := 0
	if height > 0 {
		from := to - height
		if from < 0 {
			from = 0
		}
	}

	for y := to; y >= from; y-- {
		fmt.Print("|")
		for x := 0; x < TowerWidth; x++ {
			c := "."
			if tower.At(x, y) {
				c = "#"
			} else if tower.falling.At(x, y) {
				c = "@"
			}
			fmt.Print(c)
		}
		fmt.Println("|")
	}
	if tower.culled == 0 {
		fmt.Print("+")
		for x := 0; x < TowerWidth; x++ {
			fmt.Print("-")
		}
		fmt.Println("+")
	} else {
		for x := -1; x <= TowerWidth; x++ {
			fmt.Print("^")
		}
		fmt.Println()
	}
}

func (tower *Tower) Print() {
	tower.PrintWithHeight(100)
}

func main() {
	steps := 2022

	if len(os.Args) > 1 {
		var err error
		steps, err = strconv.Atoi(os.Args[1])
		if err != nil {
			panic(err)
		}
	}

	reader := bufio.NewReader(os.Stdin)
	jp := ParseJetPattern(reader)

	const debug = false

	tower := MakeTower(jp)
	tower.DropRocks(steps, debug)
	fmt.Println(tower.TotalHeight())
}
