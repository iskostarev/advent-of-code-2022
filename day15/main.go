package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type Pos struct {
	X, Y int
}

type Cell int

const (
	CellEmpty Cell = iota
	CellSensor
	CellBeacon
)

type LineChunk struct {
	Min, Max int
}

type SparseLine struct {
	min, max int
	removed  []LineChunk
}

type Grid struct {
	sensors map[Pos]int
	beacons map[Pos]bool
}

type Parser struct {
	regex *regexp.Regexp
}

func MakeSparseLine(min, max int) (result SparseLine) {
	result.min = min
	result.max = max
	result.removed = []LineChunk{}
	return
}

func (sl *SparseLine) simplify() {
	done := false
	for !done {
		done = true
		newRemoved := []LineChunk{}
		for _, chunk := range sl.removed {
			if chunk.Max < sl.min || chunk.Min > sl.max {
				continue
			}

			if chunk.Min <= sl.min && chunk.Max > sl.min {
				sl.min = chunk.Max
				done = false
			}
			if chunk.Max >= sl.max && chunk.Min < sl.max {
				sl.max = chunk.Min
				done = false
			}

			newRemoved = append(newRemoved, chunk)
		}
		sl.removed = newRemoved
	}
}

func IntersectsOrAdjacent(lhs, rhs LineChunk) bool {
	if lhs.Min > rhs.Min {
		lhs, rhs = rhs, lhs
	}
	return lhs.Max >= rhs.Min
}

func Intersection(lhs, rhs LineChunk) LineChunk {
	if lhs.Min > rhs.Min {
		lhs, rhs = rhs, lhs
	}
	lhs.Min = rhs.Min
	if rhs.Max < lhs.Max {
		lhs.Max = rhs.Max
	}
	return lhs
}

func Union(lhs, rhs LineChunk) LineChunk {
	if lhs.Min > rhs.Min {
		lhs, rhs = rhs, lhs
	}
	if rhs.Max > lhs.Max {
		lhs.Max = rhs.Max
	}
	return lhs
}

func (sl *SparseLine) RemoveChunk(chunk LineChunk) {
	if chunk.Max < sl.min || chunk.Min > sl.max {
		return
	}

	added := false
	for i, hole := range sl.removed {
		if IntersectsOrAdjacent(chunk, hole) {
			sl.removed[i] = Union(chunk, hole)
			added = true
			break
		}
	}
	if !added {
		sl.removed = append(sl.removed, chunk)
	}
	sl.simplify()
	slices.SortFunc(sl.removed, func(lhs, rhs LineChunk) bool {
		return lhs.Min < rhs.Min
	})
}

func (sl *SparseLine) Traverse(cb func(int)) {
	for i := sl.min; i <= sl.max; i++ {
		ok := true
		for _, chunk := range sl.removed {
			if chunk.Min <= i && i <= chunk.Max {
				ok = false
				break
			}
		}
		if ok {
			cb(i)
		}
	}
}

func MakeGrid() (result Grid) {
	result.sensors = make(map[Pos]int)
	result.beacons = make(map[Pos]bool)
	return
}

func Abs(val int) int {
	if val < 0 {
		return -val
	}
	return val
}

func Distance(lhs, rhs Pos) int {
	return Abs(lhs.X-rhs.X) + Abs(lhs.Y-rhs.Y)
}

func (grid *Grid) At(pos Pos) Cell {
	if grid.sensors[pos] != 0 {
		return CellSensor
	} else if grid.beacons[pos] {
		return CellBeacon
	} else {
		return CellEmpty
	}
}

func (grid *Grid) Reachable(pos Pos) bool {
	if grid.At(pos) != CellEmpty {
		return true
	}
	for sensor, radius := range grid.sensors {
		if Distance(sensor, pos) <= radius {
			return true
		}
	}
	return false
}

func (grid *Grid) AddSensor(sensor, beacon Pos) {
	grid.sensors[sensor] = Distance(sensor, beacon)
	grid.beacons[beacon] = true
}

func SensorRow(sensor Pos, radius, y int) (int, int) {
	d := radius - Abs(sensor.Y-y)
	return sensor.X - d, sensor.X + d
}

func (grid *Grid) NoBeaconsInRow(y int) int {
	marked := map[int]bool{}
	for sensor, radius := range grid.sensors {
		begin, end := SensorRow(sensor, radius, y)
		for x := begin; x <= end; x++ {
			if grid.At(Pos{x, y}) == CellEmpty {
				marked[x] = true
			}
		}
	}
	return len(marked)
}

func (grid *Grid) findDistressBeaconInRow(y, maxCoord int) (found bool, result int) {
	candidates := MakeSparseLine(0, maxCoord)
	for sensor, radius := range grid.sensors {
		min, max := SensorRow(sensor, radius, y)
		candidates.RemoveChunk(LineChunk{min, max})
	}

	found = false
	candidates.Traverse(func(x int) {
		if found {
			panic("More than 1 beacon found")
		}
		found = true
		result = x
	})

	return
}

func splitIntoChunks(min, max, chunks int) (result []LineChunk) {
	result = []LineChunk{}
	size := max - min + 1
	if size%chunks != 0 {
		size += (chunks - size%chunks)
	}
	length := size / chunks
	curMin := min
	for {
		curMax := curMin + length - 1
		if curMax > max {
			curMax = max
		}
		result = append(result, LineChunk{curMin, curMax})
		curMin += length
		if curMin > max {
			return
		}
	}
}

func (grid *Grid) FindDistressBeacon(maxCoord int) (result Pos) {
	const jobCount = 16
	const progressInterval = 100
	segments := splitIntoChunks(0, maxCoord, jobCount)
	resultChan := make(chan Pos)
	done := make(chan int)
	progressReports := make(chan int)
	remainingJobs := len(segments)
	progress := 0

	for _, segment := range segments {
		go func(segment LineChunk) {
			lastProgress := segment.Min
			for y := segment.Min; y <= segment.Max; y++ {
				if y%progressInterval == 0 {
					progressReports <- (y - lastProgress)
					lastProgress = y
				}
				found, x := grid.findDistressBeaconInRow(y, maxCoord)
				if found {
					resultChan <- Pos{x, y}
					break
				}
			}
			progressReports <- (segment.Max - lastProgress)
			done <- 0
		}(segment)
	}

	for {
		select {
		case result = <-resultChan:
			return
		case progrep := <-progressReports:
			progress += progrep
			//fmt.Fprintf(os.Stderr, "Progress: %d/%d (%d%%)\n", progress, maxCoord, progress*100/maxCoord)
		case <-done:
			remainingJobs--
			if remainingJobs == 0 {
				panic("Beacon not found!")
			}
		}
	}

}

func MakeParser() (result Parser) {
	result.regex = regexp.MustCompile(`^Sensor at x=(-?\d+), y=(-?\d+): closest beacon is at x=(-?\d+), y=(-?\d+)$`)
	return
}

func ParseInt(s string) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return result
}

func (parser *Parser) ParseLine(line string) (sensor, beacon Pos) {
	matches := parser.regex.FindStringSubmatch(line)
	if len(matches) != 5 {
		panic("Invalid line")
	}
	sensor = Pos{ParseInt(matches[1]), ParseInt(matches[2])}
	beacon = Pos{ParseInt(matches[3]), ParseInt(matches[4])}
	return
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	parser := MakeParser()
	grid := MakeGrid()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		sensor, beacon := parser.ParseLine(line)
		grid.AddSensor(sensor, beacon)
	}

	if len(os.Args) < 3 {
		panic("Arguments expected")
	}

	if os.Args[1] == "1" {
		row := ParseInt(os.Args[2])
		fmt.Println(grid.NoBeaconsInRow(row))
	} else if os.Args[1] == "2" {
		maxCoord := ParseInt(os.Args[2])
		pos := grid.FindDistressBeacon(maxCoord)
		fmt.Println(pos.X*4000000 + pos.Y)
	} else {
		panic("First argument must be 1 or 2")
	}
}
