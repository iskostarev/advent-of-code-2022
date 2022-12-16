package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
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

type Grid struct {
	sensors map[Pos]int
	beacons map[Pos]bool
}

type Parser struct {
	regex *regexp.Regexp
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
	if len(os.Args) < 2 {
		panic("Argument expected")
	}

	row := ParseInt(os.Args[1])

	scanner := bufio.NewScanner(os.Stdin)
	parser := MakeParser()
	grid := MakeGrid()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		sensor, beacon := parser.ParseLine(line)
		grid.AddSensor(sensor, beacon)
	}

	fmt.Println(grid.NoBeaconsInRow(row))
}
