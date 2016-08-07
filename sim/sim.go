package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/dbtleonia/fantasy"
)

var (
	lambda = flag.Float64("lambda", 0.36, "rate parameter for humanoid")
)

func main() {
	flag.Parse()
	if flag.NArg() != 5 {
		fmt.Fprintf(os.Stderr, "Usage: %s [<flags>] <order-csv> <players-csv> <rules-csv> <schema> <strategies>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	var (
		orderCsv       = flag.Arg(0)
		playersCsv     = flag.Arg(1)
		rulesCsv       = flag.Arg(2)
		schema         = flag.Arg(3)
		strategyString = flag.Arg(4)
		numTeams       = len(strategyString)
	)
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
		optStrategies[i] = fantasy.NewHumanoid(rules, *lambda)
	}
	scorer := &fantasy.Scorer{[]byte(schema)}

	strategies := make([]fantasy.Strategy, numTeams)
	for i, ch := range strategyString {
		switch ch {
		case 'A':
			strategies[i] = fantasy.NewAutopick(rules)
		case 'H':
			strategies[i] = fantasy.NewHumanoid(rules, *lambda)
		case 'O':
			strategies[i] = fantasy.NewOptimize(rules, optStrategies, scorer, 100)
		default:
			log.Fatalf("Invalid strategy: %c", ch)
		}
	}

	rand.Seed(time.Now().Unix())
	fantasy.RunDraft(state, order, strategies)

	for i, team := range state.Teams {
		fmt.Printf("Team #%d [%c] = %f\n", i, strategyString[i], scorer.Score(team))
		for _, player := range team.PlayersByPick() {
			fmt.Printf("  %s\n", player)
		}
	}
	for i, team := range state.Teams {
		fmt.Printf("Team #%d [%c] = %f\n", i, strategyString[i], scorer.Score(team))
	}
}
