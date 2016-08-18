package main

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
)

func allowedPos(schema, roster []byte) (autopick, humanoid string) {
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
		return allowed, allowed
	}

	// Rule #2: Give priority to starters.
	// Autopick fills all the starters, then allows any position.
	if startersCount > 0 {
		autopick = allowed
	} else {
		autopick = "DKQRTW"
	}
	// Humanoid fills all the starters except D and K, then allows any
	// position.
	if starters['Q'] > 0 || starters['R'] > 0 || starters['T'] > 0 || starters['W'] > 0 || starters['X'] > 0 {
		humanoid = allowed
	} else {
		humanoid = "DKQRTW"
	}
	return autopick, humanoid
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <combos-file> <schema>", os.Args[0])
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	schema := []byte(os.Args[2])
	scanner := bufio.NewScanner(f)
	out := csv.NewWriter(os.Stdout)
	for scanner.Scan() {
		roster := scanner.Text()
		autopick, humanoid := allowedPos(schema, []byte(roster))
		out.Write([]string{roster, autopick, humanoid})
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	out.Flush()
}
