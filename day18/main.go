package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const Dim = 3

type Vector [Dim]int

type MovingPoint struct {
	Pos, Prev Vector
}

type Cube struct {
	Surfaces         map[Vector]bool
	ExteriorSurfaces map[Vector]bool
}

type Model struct {
	Cubes              map[Vector]Cube
	BoundMin, BoundMax Vector
}

func Ones() (result Vector) {
	for i := 0; i < Dim; i++ {
		result[i] = 1
	}
	return
}

func (v Vector) Neg() (result Vector) {
	for i := 0; i < Dim; i++ {
		result[i] = -v[i]
	}
	return
}

func (lhs Vector) Add(rhs Vector) Vector {
	for i := 0; i < Dim; i++ {
		lhs[i] += rhs[i]
	}
	return lhs
}

func (lhs Vector) Sub(rhs Vector) Vector {
	for i := 0; i < Dim; i++ {
		lhs[i] -= rhs[i]
	}
	return lhs
}

func (vec Vector) InBounds(min, max Vector) bool {
	for i := 0; i < Dim; i++ {
		if vec[i] < min[i] {
			return false
		}
		if vec[i] > max[i] {
			return false
		}
	}
	return true
}

func (vec Vector) IsUnit() bool {
	one := false
	for i := 0; i < Dim; i++ {
		if vec[i] == 0 {
			continue
		} else if vec[i] == 1 || vec[i] == -1 {
			if one {
				return false
			}
			one = true
		} else {
			return false
		}
	}
	return one
}

func Min(lhs, rhs int) int {
	if lhs < rhs {
		return lhs
	} else {
		return rhs
	}
}

func Max(lhs, rhs int) int {
	if lhs > rhs {
		return lhs
	} else {
		return rhs
	}
}

func LowerBound(lhs, rhs Vector) (result Vector) {
	for i := 0; i < Dim; i++ {
		result[i] = Min(lhs[i], rhs[i])
	}
	return
}

func UpperBound(lhs, rhs Vector) (result Vector) {
	for i := 0; i < Dim; i++ {
		result[i] = Max(lhs[i], rhs[i])
	}
	return
}

func traverseDirections(cb func(Vector)) {
	for i := 0; i < Dim; i++ {
		vec := Vector{}
		vec[i] = 1
		cb(vec)
		vec[i] = -1
		cb(vec)
	}
}

func MakeCube() (result Cube) {
	result.Surfaces = map[Vector]bool{}
	result.ExteriorSurfaces = map[Vector]bool{}
	traverseDirections(func(d Vector) {
		result.Surfaces[d] = true
	})
	return
}

func MakeModel() (result Model) {
	result.Cubes = map[Vector]Cube{}
	return
}

func (model *Model) Has(coords Vector) bool {
	_, ok := model.Cubes[coords]
	return ok
}

func (model *Model) AddCube(coords Vector) {
	model.Cubes[coords] = MakeCube()
	model.BoundMin = LowerBound(model.BoundMin, coords)
	model.BoundMax = UpperBound(model.BoundMax, coords)
	traverseDirections(func(dir Vector) {
		neighbour := coords.Add(dir)
		if model.Has(neighbour) {
			model.Cubes[coords].Surfaces[dir] = false
			model.Cubes[neighbour].Surfaces[dir.Neg()] = false
		}
	})
}

func (model *Model) MarkExterior() {
	min := model.BoundMin.Sub(Ones())
	max := model.BoundMax.Add(Ones())

	scanners := []MovingPoint{MovingPoint{min, min}}
	visited := map[MovingPoint]bool{}

	for len(scanners) != 0 {
		next := []MovingPoint{}
		for _, point := range scanners {
			if visited[point] {
				continue
			}
			visited[point] = true
			if model.Has(point.Pos) {
				cube := model.Cubes[point.Pos]
				dir := point.Prev.Sub(point.Pos)
				if !dir.IsUnit() {
					panic("Impossible condition")
				}
				cube.ExteriorSurfaces[dir] = true
				model.Cubes[point.Pos] = cube
				continue
			}

			traverseDirections(func(dir Vector) {
				neighbour := point.Pos.Add(dir)
				nextPoint := MovingPoint{neighbour, point.Pos}
				if neighbour.InBounds(min, max) && !visited[nextPoint] {
					next = append(next, nextPoint)
				}
			})
		}
		scanners = next
	}
}

func (model *Model) SurfaceArea(exteriorOnly bool) (result int) {
	if exteriorOnly {
		model.MarkExterior()
	}
	for _, cube := range model.Cubes {
		surfaces := cube.Surfaces
		if exteriorOnly {
			surfaces = cube.ExteriorSurfaces
		}
		for _, visible := range surfaces {
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

func ParseVector(str string) (result Vector) {
	fields := strings.Split(str, ",")
	if len(fields) != Dim {
		panic("Invalid coord count")
	}
	for i := 0; i < Dim; i++ {
		result[i] = ParseInt(fields[i])
	}
	return
}

func main() {
	mode2 := false
	if (len(os.Args) > 1) && (os.Args[1] == "2") {
		mode2 = true
	}

	scanner := bufio.NewScanner(os.Stdin)
	model := MakeModel()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		model.AddCube(ParseVector(line))
	}
	fmt.Println(model.SurfaceArea(mode2))
}
