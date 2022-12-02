package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Shape int

const (
	Rock     Shape = 0
	Paper          = 1
	Scissors       = 2
)

type GameOutcome int

const (
	Loss GameOutcome = 0
	Draw             = 1
	Win              = 2
)

func parseOppShape(code string) Shape {
	switch code {
	case "A":
		return Rock
	case "B":
		return Paper
	case "C":
		return Scissors
	default:
		panic("Unknown shape code: " + code)
	}
}

func parseMyShape(code string) Shape {
	switch code {
	case "X":
		return Rock
	case "Y":
		return Paper
	case "Z":
		return Scissors
	default:
		panic("Unknown shape code: " + code)
	}
}

func getGameOutcome(my, opp Shape) GameOutcome {
	matrix := [3][3]GameOutcome{
		{Draw, Loss, Win},
		{Win, Draw, Loss},
		{Loss, Win, Draw},
	}

	return matrix[int(my)][int(opp)]
}

func calcScore(my, opp Shape) int {
	outcome := getGameOutcome(my, opp)
	score := int(outcome) * 3
	score += int(my) + 1
	return score
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	totalScore := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		columns := strings.Split(line, " ")
		if len(columns) != 2 {
			panic("Failed to parse line")
		}

		oppShape := parseOppShape(columns[0])
		myShape := parseMyShape(columns[1])
		totalScore += calcScore(myShape, oppShape)
	}
	fmt.Println(totalScore)
}
