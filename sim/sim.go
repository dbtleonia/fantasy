package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/dbtleonia/fantasy"
)

func main() {
	if len(os.Args) != 6 {
		log.Fatalf("Usage: %s <order-csv> <players-csv> <rules-csv> <schema> <strategies>", os.Args[0])
	}
	var (
		orderCsv       = os.Args[1]
		playersCsv     = os.Args[2]
		rulesCsv       = os.Args[3]
		schema         = os.Args[4]
		strategyString = os.Args[5]
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
		optStrategies[i] = fantasy.NewHumanoid(rules, 0.36)
	}
	scorer := &fantasy.Scorer{[]byte(schema)}

	strategies := make([]fantasy.Strategy, numTeams)
	for i, ch := range strategyString {
		switch ch {
		case 'A':
			strategies[i] = fantasy.NewAutopick(rules)
		case 'H':
			strategies[i] = fantasy.NewHumanoid(rules, 0.36)
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
