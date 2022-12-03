package main

import (
	"bufio"
	"fmt"
	"golang.org/x/exp/slices"
	"os"
	"strings"
)

type RucksackItem byte

type Rucksack struct {
	First  []RucksackItem
	Second []RucksackItem
}

func (item RucksackItem) priority() int {
	return int(item)
}

func parseRucksackCompartment(str string) (result []RucksackItem) {
	result = make([]RucksackItem, len(str))
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c >= 'a' && c <= 'z' {
			result[i] = RucksackItem(c - 'a' + 1)
		} else if c >= 'A' && c <= 'Z' {
			result[i] = RucksackItem(c - 'A' + 27)
		} else {
			panic("Unexpected item")
		}
	}
	slices.Sort(result)
	return
}

func parseRucksack(str string) (result Rucksack) {
	if len(str)%2 == 1 {
		panic("Uneven compartments")
	}

	compLen := len(str) / 2
	result.First = parseRucksackCompartment(str[:compLen])
	result.Second = parseRucksackCompartment(str[compLen:])
	return
}

func (rucksack *Rucksack) findCommonItem() (result RucksackItem) {
	resultSet := false

	for _, item := range rucksack.First {
		_, found := slices.BinarySearch(rucksack.Second, item)
		if found {
			if resultSet && item != result {
				panic(fmt.Sprintf("More then one common item type: %d, %d", result, item))
			}
			result = item
			resultSet = true
		}
	}
	if resultSet {
		return
	}
	panic("No common item")
}

func (rucksack *Rucksack) contains(item RucksackItem) bool {
	for _, compartment := range [][]RucksackItem{rucksack.First, rucksack.Second} {
		_, found := slices.BinarySearch(compartment, item)
		if found {
			return true
		}
	}
	return false
}

func findBadge(group []Rucksack) (result RucksackItem) {
	resultSet := false
	checkCompartment := func(compartment []RucksackItem) {
		for _, item := range compartment {
			common := true
			for i := 1; i < len(group); i++ {
				if !group[i].contains(item) {
					common = false
					break
				}
			}

			if common {
				if resultSet && item != result {
					panic(fmt.Sprintf("More then one common item type: %d, %d", result, item))
				}
				result = item
				resultSet = true
			}
		}
	}

	checkCompartment(group[0].First)
	checkCompartment(group[0].Second)
	return
}

func mode1() {
	scanner := bufio.NewScanner(os.Stdin)
	priosum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		rucksack := parseRucksack(line)
		priosum += rucksack.findCommonItem().priority()
	}
	fmt.Println(priosum)
}

func mode2() {
	const GROUP_SIZE = 3

	scanner := bufio.NewScanner(os.Stdin)
	priosum := 0
	group := make([]Rucksack, 0, GROUP_SIZE)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		group = append(group, parseRucksack(line))
		if len(group) != GROUP_SIZE {
			continue
		}

		priosum += findBadge(group).priority()
		group = group[:0]
	}

	if len(group) != 0 {
		panic("Excess lines")
	}

	fmt.Println(priosum)
}

func main() {
	if (len(os.Args) > 1) && (os.Args[1] == "2") {
		mode2()
	} else {
		mode1()
	}
}
