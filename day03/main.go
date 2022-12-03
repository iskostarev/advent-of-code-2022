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

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	priosum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		rucksack := parseRucksack(line)
		priosum += rucksack.findCommonItem().priority()
	}
	fmt.Println(priosum)
}
