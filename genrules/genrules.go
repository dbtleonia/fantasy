package main

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
)

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
		starters := make(map[byte]int)
		for _, ch := range schema {
			if ch != 'B' {
				starters[ch]++
			}
		}
		if starters['D'] != 1 || starters['K'] != 1 {
			log.Fatal("Illegal roster")
		}
		roster := scanner.Text()
		for _, pos := range []byte(roster) {
			if starters[pos] > 0 {
				starters[pos]--
				continue
			}
			switch pos {
			case 'R', 'T', 'W':
				if starters['X'] > 0 {
					starters['X']--
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

		// Autopick fills all the starters, then allows any position.
		autopick := allowed
		if allowed == "" {
			autopick = "DKQRTW"
		}

		// Humanoid fills all the starters except D and K, then allows any
		// position, except that D and K are required at the end.
		humanoid := allowed
		switch humanoid {
		case "":
			humanoid = "DKQRTW"
		case "D":
			if len(roster)+1 < len(schema) {
				humanoid = "DQRTW"
			}
		case "K":
			if len(roster)+1 < len(schema) {
				humanoid = "KQRTW"
			}
		case "DK":
			if len(roster)+2 < len(schema) {
				humanoid = "DKQRTW"
			}
		}

		out.Write([]string{roster, autopick, humanoid})
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	out.Flush()
}
