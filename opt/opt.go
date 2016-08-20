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
	lambda    = flag.Float64("lambda", 0.36, "rate parameter for humanoid")
	numTrials = flag.Int("num_trials", 200, "number of trials to run for optimize")
)

func main() {
	flag.Parse()
	if flag.NArg() != 5 {
		fmt.Fprintf(os.Stderr, "Usage: %s [<flags>] <order-csv> <players-csv> <rules-csv> <schema> <strategies>", os.Args[0])
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

	optStrategies := make([]fantasy.Strategy, numTeams)
	for i, ch := range strategyString {
		switch ch {
		case 'A':
			optStrategies[i] = fantasy.NewAutopick(rules, true)
		case 'H':
			optStrategies[i] = fantasy.NewHumanoid(rules, true, *lambda)
		case 'O':
			// Approiximate Optimize with Autopick.
			// TODO: Figure out a better approximation.
			optStrategies[i] = fantasy.NewAutopick(rules, false)
		default:
			log.Fatalf("Invalid strategy: %c", ch)
		}
	}

	scorer := &fantasy.Scorer{[]byte(schema)}

	// Use optimize for the next pick regardless of what the strategies
	// arg says.
	optimize := fantasy.NewOptimize(rules, optStrategies, scorer, *numTrials)

	rand.Seed(time.Now().Unix())
	for _, c := range optimize.Candidates(state, order) {
		fmt.Printf("%.2f %s\n", c.Value, c.Player)
	}
}
