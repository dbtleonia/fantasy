package main

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
)

func allowedPos(schema, priorityStarters, posMin, posMax, roster []byte) string {
	starters := make(map[byte]int)
	startersCount := 0
	for _, ch := range schema {
		if ch != 'B' {
			starters[ch]++
			startersCount++
		}
	}
	for _, pos := range roster {
		if starters[pos] > 0 {
			starters[pos]--
			startersCount--
			continue
		}
		switch pos {
		case 'R', 'T', 'W':
			if starters['X'] > 0 {
				starters['X']--
				startersCount--
			}
		}
	}
	var allowed string
	for _, pos := range []byte("DKQRTW") {
		if starters[pos] > 0 {
			allowed += string(pos)
			continue
		}
		switch pos {
		case 'R', 'T', 'W':
			if starters['X'] > 0 {
				allowed += string(pos)
			}
		}
	}

	// Rule #1: Fill starters by end of draft.
	if len(roster)+startersCount >= len(schema) {
		return allowed
	}

	// Rule #2: Give priority to starters.
	for _, pos := range priorityStarters {
		if starters[pos] > 0 {
			return allowed
		}
	}

	// Rule #3: Fill position minimums by end of draft.
	needMin := make(map[byte]int)
	needMinCount := 0
	for _, pos := range posMin {
		needMin[pos]++
		needMinCount++
	}
	for _, pos := range roster {
		if needMin[pos] > 0 {
			needMin[pos]--
			needMinCount--
		}
	}
	if len(roster)+needMinCount >= len(schema) {
		var allowedMin string
		for _, pos := range []byte("DKQRTW") {
			if needMin[pos] > 0 {
				allowedMin += string(pos)
			}
		}
		return allowedMin
	}

	// Rule #4: Filter out positions that already reached maximum.
	leftMax := make(map[byte]int)
	for _, pos := range posMax {
		leftMax[pos]++
	}
	for _, pos := range roster {
		if leftMax[pos] > 0 {
			leftMax[pos]--
		}
	}
	var allowedMax string
	for _, pos := range []byte("DKQRTW") {
		if leftMax[pos] > 0 {
			allowedMax += string(pos)
		}
	}
	return allowedMax
}

func main() {
	if len(os.Args) != 7 {
		log.Fatalf("Usage: %s <combos-file> <schema> <autopick-min> <autopick-max> <humanoid-min> <humanoid-max>", os.Args[0])
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	var (
		schema      = []byte(os.Args[2])
		autopickMin = []byte(os.Args[3])
		autopickMax = []byte(os.Args[4])
		humanoidMin = []byte(os.Args[5])
		humanoidMax = []byte(os.Args[6])

		// TOOD: Make these args?
		autopickPriority = []byte("DKQRTWX")
		humanoidPriority = []byte("QRWX")
	)
	scanner := bufio.NewScanner(f)
	out := csv.NewWriter(os.Stdout)
	for scanner.Scan() {
		roster := scanner.Text()
		autopick := allowedPos(schema, autopickPriority, autopickMin, autopickMax, []byte(roster))
		humanoid := allowedPos(schema, humanoidPriority, humanoidMin, humanoidMax, []byte(roster))
		out.Write([]string{roster, autopick, humanoid})
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	out.Flush()
}
