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
	seed      = flag.Int64("seed", 0, "seed for rand; if 0 uses time")
	bench     = flag.Bool("bench", false, "score bench (using hardcoded weights)")
)

func main() {
	flag.Parse()
	if flag.NArg() != 5 {
		fmt.Fprintf(os.Stderr, "Usage: %s [<flags>] <order-csv> <players-csv> <rules-csv> <schema> <strategies>", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	s := *seed
	if s == 0 {
		s = time.Now().Unix()
	}
	fmt.Printf("Using seed %d\n", s)
	rand.Seed(s)

	var (
		orderCsv       = flag.Arg(0)
		playersCsv     = flag.Arg(1)
		rulesCsv       = flag.Arg(2)
		schema         = flag.Arg(3)
		strategyString = flag.Arg(4)
		numTeams       = len(strategyString)
	)
	rawOrder, err := fantasy.ReadOrder(orderCsv)
	if err != nil {
		log.Fatal(err)
	}
	state, order, err := fantasy.ReadState(playersCsv, numTeams, rawOrder)
	if err != nil {
		log.Fatal(err)
	}
	rules, err := fantasy.ReadRules(rulesCsv)
	if err != nil {
		log.Fatal(err)
	}

	optStrategies := make([]fantasy.Strategy, len(order))
	for i := 1; i < len(order); i++ {
		t := order[i]
		if t == -1 { // this pick is a keeper, skip it
			continue
		}
		ch := strategyString[t]
		switch ch {
		case 'A':
			optStrategies[i] = fantasy.NewAutopick(order, rules, true)
		case 'H':
			optStrategies[i] = fantasy.NewHumanoid(order, rules, true, *lambda)
		case 'O':
			// Approximate Optimize with Humanoid.
			// TODO: Figure out a better approximation.
			optStrategies[i] = fantasy.NewHumanoid(order, rules, false, *lambda)
		default:
			log.Fatalf("Invalid strategy: %c", ch)
		}
	}

	scorer := &fantasy.Scorer{[]byte(schema), *bench}

	// Use optimize for the next pick regardless of what the strategies
	// arg says.
	optimize := fantasy.NewOptimize(order, optStrategies, rules, scorer, *numTrials)

	for _, c := range optimize.Candidates(state) {
		fmt.Printf("%.2f %s\n", c.Value, c.Player)
	}
}
