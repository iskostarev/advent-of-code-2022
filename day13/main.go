package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
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
			lhs = PacketListData([]PacketData{lhs})
		} else if rhs.IsInt() {
			rhs = PacketListData([]PacketData{rhs})
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

func ParsePacket(scanner *bufio.Scanner) (ok bool, result PacketData) {
	ok = scanner.Scan()
	if !ok {
		return
	}
	var j any
	err := json.Unmarshal([]byte(scanner.Text()), &j)
	if err != nil {
		panic(err)
	}
	result = convertRawPacket(j)
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

func main() {
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
