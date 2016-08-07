package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/dbtleonia/fantasy"
)

var (
	lambda = flag.Float64("lambda", 0.36, "rate parameter for humanoid")
)

func main() {
	flag.Parse()
	if flag.NArg() != 5 {
		fmt.Fprintf(os.Stderr, "Usage: %s [<flags>] <order-csv> <players-csv> <rules-csv> <schema> <num-teams>", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	var (
		orderCsv   = flag.Arg(0)
		playersCsv = flag.Arg(1)
		rulesCsv   = flag.Arg(2)
		schema     = flag.Arg(3)
	)
	numTeams, err := strconv.Atoi(flag.Arg(4))
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
		optStrategies[i] = fantasy.NewHumanoid(rules, *lambda)
	}
	scorer := &fantasy.Scorer{[]byte(schema)}

	optimize := fantasy.NewOptimize(rules, optStrategies, scorer, 200)

	rand.Seed(time.Now().Unix())
	for _, c := range optimize.Candidates(state, order) {
		fmt.Printf("%3d %.2f %s\n", c.Index, c.Value, state.Undrafted[c.Index])
	}
}
