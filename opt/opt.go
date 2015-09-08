package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/dbtleonia/fantasy"
)

func main() {
	if len(os.Args) != 6 {
		log.Fatalf("Usage: %s <order-csv> <players-csv> <rules-csv> <schema> <num-teams>", os.Args[0])
	}
	var (
		orderCsv   = os.Args[1]
		playersCsv = os.Args[2]
		rulesCsv   = os.Args[3]
		schema     = os.Args[4]
	)
	numTeams, err := strconv.Atoi(os.Args[5])
	if err != nil {
		log.Fatal(err)
	}
	order, err := fantasy.ReadOrder(orderCsv)
	if err != nil {
		log.Fatal(err)
	}
	state, err := fantasy.ReadState(playersCsv, numTeams, order)
	if err != nil {
		log.Fatal(err)
	}
	rules, err := fantasy.ReadRules(rulesCsv)
	if err != nil {
		log.Fatal(err)
	}
	optStrategies := make([]fantasy.Strategy, len(state.Teams))
	for i, _ := range state.Teams {
		optStrategies[i] = fantasy.NewHumanoid(rules, 0.36)
	}
	scorer := &fantasy.Scorer{[]byte(schema)}

	optimize := fantasy.NewOptimize(rules, optStrategies, scorer, 200)

	rand.Seed(time.Now().Unix())
	for _, c := range optimize.Candidates(state, order) {
		fmt.Printf("%3d %.2f %s\n", c.Index, c.Value, state.Undrafted[c.Index])
	}
}
