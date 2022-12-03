package main

import (
	"bufio"
	"fmt"
	"golang.org/x/exp/slices"
	"os"
	"strings"
)

type RucksackItem byte

type RucksackWithCompartments struct {
	First  []RucksackItem
	Second []RucksackItem
}

type WholeRucksack []RucksackItem

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

func parseRucksackWithCompartments(str string) (result RucksackWithCompartments) {
	if len(str)%2 == 1 {
		panic("Uneven compartments")
	}

	compLen := len(str) / 2
	result.First = parseRucksackCompartment(str[:compLen])
	result.Second = parseRucksackCompartment(str[compLen:])
	return
}

func parseWholeRucksack(str string) (result WholeRucksack) {
	return parseRucksackCompartment(str)
}

func appendUniqItem(items []RucksackItem, next RucksackItem) []RucksackItem {
	if len(items) > 0 && items[len(items)-1] == next {
		return items
	}
	return append(items, next)
}

func findCommonItems(lhs, rhs []RucksackItem) (result []RucksackItem) {
	result = make([]RucksackItem, 0, len(lhs))

	for _, item := range lhs {
		i, found := slices.BinarySearch(rhs, item)
		if found {
			result = appendUniqItem(result, item)
			rhs = rhs[i:]
		}
	}
	return
}

func ensureSingleItem(arr []RucksackItem) RucksackItem {
	if len(arr) == 0 {
		panic("No item")
	} else if len(arr) > 1 {
		panic(fmt.Sprintf("Multiple items: %v", arr))
	} else {
		return arr[0]
	}
}

func findSingleCommonItem(lhs, rhs []RucksackItem) (result RucksackItem) {
	return ensureSingleItem(findCommonItems(lhs, rhs))
}

func findBadge(group []WholeRucksack) (result RucksackItem) {
	common := group[0]
	for _, next := range group[1:] {
		common = findCommonItems(common, next)
	}
	return ensureSingleItem(common)
}

func mode1() {
	scanner := bufio.NewScanner(os.Stdin)
	priosum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		rucksack := parseRucksackWithCompartments(line)
		priosum += findSingleCommonItem(rucksack.First, rucksack.Second).priority()
	}
	fmt.Println(priosum)
}

func mode2() {
	const GROUP_SIZE = 3

	scanner := bufio.NewScanner(os.Stdin)
	priosum := 0
	group := make([]WholeRucksack, 0, GROUP_SIZE)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		group = append(group, parseWholeRucksack(line))
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
