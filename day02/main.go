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

func parseGoal(code string) GameOutcome {
	switch code {
	case "X":
		return Loss
	case "Y":
		return Draw
	case "Z":
		return Win
	default:
		panic("Unknown goal code: " + code)
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

func chooseShape(opp Shape, goal GameOutcome) Shape {
	for _, my := range [...]Shape{Rock, Paper, Scissors} {
		if getGameOutcome(my, opp) == goal {
			return my
		}
	}
	panic("Failed to choose shape")
}

func calcScore(my, opp Shape) int {
	outcome := getGameOutcome(my, opp)
	score := int(outcome) * 3
	score += int(my) + 1
	return score
}

func main() {
	mode1 := true
	if (len(os.Args) > 1) && (os.Args[1] == "2") {
		mode1 = false
	}

	scanner := bufio.NewScanner(os.Stdin)
	totalScore := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		columns := strings.Split(line, " ")
		if len(columns) != 2 {
			panic("Failed to parse line")
		}

		oppShape := parseOppShape(columns[0])
		var myShape Shape

		if mode1 {
			myShape = parseMyShape(columns[1])
		} else {
			goal := parseGoal(columns[1])
			myShape = chooseShape(oppShape, goal)
		}
		totalScore += calcScore(myShape, oppShape)
	}

	fmt.Println(totalScore)
}
