package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <num-teams> <num-rounds>", os.Args[0])
	}
	numTeams, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	numRounds, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	out := csv.NewWriter(os.Stdout)
	pick := 1
	for round := 0; round < numRounds; round++ {
		for i := 0; i < numTeams; i++ {
			t := i
			if round%2 == 1 {
				t = numTeams - i - 1
			}
			out.Write([]string{
				strconv.Itoa(pick),
				fmt.Sprintf("(Round %d #%d)", round, i),
				strconv.Itoa(t),
			})
			pick++
		}
	}
	out.Flush()
}
