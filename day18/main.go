package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Vector3 struct {
	X, Y, Z int
}

type Cube struct {
	Surfaces map[Vector3]bool
}

type Model struct {
	Cubes map[Vector3]Cube
}

func (v Vector3) Neg() (result Vector3) {
	result.X = -v.X
	result.Y = -v.Y
	result.Z = -v.Z
	return
}

func (lhs Vector3) Add(rhs Vector3) (result Vector3) {
	result.X = lhs.X + rhs.X
	result.Y = lhs.Y + rhs.Y
	result.Z = lhs.Z + rhs.Z
	return
}

func traverseDirections(cb func(Vector3)) {
	cb(Vector3{-1, 0, 0})
	cb(Vector3{1, 0, 0})
	cb(Vector3{0, -1, 0})
	cb(Vector3{0, 1, 0})
	cb(Vector3{0, 0, -1})
	cb(Vector3{0, 0, 1})
}

func MakeCube() (result Cube) {
	result.Surfaces = map[Vector3]bool{}
	traverseDirections(func(d Vector3) {
		result.Surfaces[d] = true
	})
	return
}

func MakeModel() (result Model) {
	result.Cubes = map[Vector3]Cube{}
	return
}

func (model *Model) Has(coords Vector3) bool {
	_, ok := model.Cubes[coords]
	return ok
}

func (model *Model) AddCube(coords Vector3) {
	model.Cubes[coords] = MakeCube()
	traverseDirections(func(dir Vector3) {
		neighbour := coords.Add(dir)
		if model.Has(neighbour) {
			model.Cubes[coords].Surfaces[dir] = false
			model.Cubes[neighbour].Surfaces[dir.Neg()] = false
		}
	})
}

func (model *Model) SurfaceArea() (result int) {
	for _, cube := range model.Cubes {
		for _, visible := range cube.Surfaces {
			if visible {
				result++
			}
		}
	}
	return
}

func ParseInt(str string) (result int) {
	result, err := strconv.Atoi(strings.TrimSpace(str))
	if err != nil {
		panic(err)
	}
	return
}

func ParseVector3(str string) (result Vector3) {
	fields := strings.Split(str, ",")
	if len(fields) != 3 {
		panic("Expected 3 coords")
	}
	result.X = ParseInt(fields[0])
	result.Y = ParseInt(fields[1])
	result.Z = ParseInt(fields[2])
	return
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	model := MakeModel()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		model.AddCube(ParseVector3(line))
	}
	fmt.Println(model.SurfaceArea())
}
