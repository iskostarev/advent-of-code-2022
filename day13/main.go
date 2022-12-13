package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"

	"golang.org/x/exp/slices"
)

type CompareResult int

const (
	CmpLess CompareResult = iota
	CmpEqual
	CmpMore
)

type PacketData interface {
	IsInt() bool
	IsList() bool
	AsInt() PacketIntData
	AsList() PacketListData
}

type PacketIntData int
type PacketListData []PacketData

func (PacketIntData) IsInt() bool {
	return true
}

func (PacketIntData) IsList() bool {
	return false
}

func (d PacketIntData) AsInt() PacketIntData {
	return d
}

func (PacketIntData) AsList() PacketListData {
	panic("Not a list")
}

func (PacketListData) IsInt() bool {
	return false
}

func (PacketListData) IsList() bool {
	return true
}

func (PacketListData) AsInt() PacketIntData {
	panic("Not an int")
}

func (d PacketListData) AsList() PacketListData {
	return d
}

func Compare(lhs, rhs PacketData) CompareResult {
	//fmt.Printf("Compare %v, %v\n", lhs, rhs)
	if lhs.IsInt() && rhs.IsInt() {
		if lhs.AsInt() < rhs.AsInt() {
			return CmpLess
		} else if lhs.AsInt() == rhs.AsInt() {
			return CmpEqual
		} else {
			return CmpMore
		}
	} else if lhs.IsList() && rhs.IsList() {
		lhsList := lhs.AsList()
		rhsList := rhs.AsList()
		for i, lhsItem := range lhsList {
			if i >= len(rhsList) {
				return CmpMore
			}
			res := Compare(lhsItem, rhsList[i])
			if res != CmpEqual {
				return res
			}
		}
		if len(lhsList) == len(rhsList) {
			return CmpEqual
		} else {
			return CmpLess
		}
	} else {
		if lhs.IsInt() {
			lhs = PacketListData{lhs}
		} else if rhs.IsInt() {
			rhs = PacketListData{rhs}
		}
		return Compare(lhs, rhs)
	}
}

func convertRawPacket(data any) PacketData {
	switch d := data.(type) {
	case float64:
		if d != math.Trunc(d) {
			panic("Integer expected")
		}
		return PacketIntData(d)
	case []any:
		result := []PacketData{}
		for _, item := range d {
			result = append(result, convertRawPacket(item))
		}
		return PacketListData(result)
	}

	panic(fmt.Sprintf("Unrecognized type: %T", data))
}

func ParsePacket(scanner *bufio.Scanner) (ok bool, result PacketListData) {
	ok = scanner.Scan()
	if !ok {
		return
	}
	var j any
	err := json.Unmarshal([]byte(scanner.Text()), &j)
	if err != nil {
		panic(err)
	}
	result = convertRawPacket(j).(PacketListData)
	return
}

func ParsePacketPair(scanner *bufio.Scanner) (ok bool, left PacketData, right PacketData) {
	ok, left = ParsePacket(scanner)
	if !ok {
		return
	}
	ok, right = ParsePacket(scanner)
	if !ok {
		panic("Second packet expected")
	}
	if scanner.Scan() {
		if scanner.Text() != "" {
			panic("Empty line expected")
		}
	}
	return
}

func mode1() {
	scanner := bufio.NewScanner(os.Stdin)
	index := 0
	sum := 0
	for {
		index++
		ok, left, right := ParsePacketPair(scanner)
		if !ok {
			break
		}

		if Compare(left, right) == CmpLess {
			sum += index
		}
	}
	fmt.Println(sum)
}

func mode2() {
	scanner := bufio.NewScanner(os.Stdin)
	divider2 := PacketListData{PacketListData{PacketIntData(2)}}
	divider6 := PacketListData{PacketListData{PacketIntData(6)}}
	packets := PacketListData{divider2, divider6}

	for {
		ok, left, right := ParsePacketPair(scanner)
		if !ok {
			break
		}

		packets = append(packets, left, right)
	}

	slices.SortStableFunc(packets, func(lhs, rhs PacketData) bool {
		return Compare(lhs, rhs) == CmpLess
	})

	divider2Loc := -1
	divider6Loc := -1

	for i, item := range packets {
		if Compare(item, divider2) == CmpEqual {
			divider2Loc = i + 1
		} else if Compare(item, divider6) == CmpEqual {
			divider6Loc = i + 1
			break
		}
	}

	if divider2Loc == -1 || divider6Loc == -1 {
		panic("Divider packets not found")
	}

	fmt.Println(divider2Loc * divider6Loc)
}

func main() {
	if (len(os.Args) > 1) && (os.Args[1] == "2") {
		mode2()
	} else {
		mode1()
	}
}
