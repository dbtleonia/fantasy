package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/dbtleonia/fantasy"
)

var (
	numTrials = flag.Int("num_trials", 50, "number of trials to run for optimize")
	seed      = flag.Int64("seed", 0, "seed for rand; if 0 uses time")
	bench     = flag.Bool("bench", false, "score bench (using hardcoded weights)")
)

func main() {
	flag.Parse()
	if flag.NArg() != 5 {
		fmt.Fprintf(os.Stderr, "Usage: %s [<flags>] <order-csv> <players-csv> <rules-csv> <schema> <strategies>\n", os.Args[0])
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

	// Generate random ADP rankings for each manager.
	rankedPlayers := make([][]fantasy.PlayerADP, numTeams)
	for t := 0; t < numTeams; t++ {
		for _, player := range state.Players {
			rankedPlayers[t] = append(rankedPlayers[t], fantasy.PlayerADP{
				PlayerID: player.ID,
				ADP:      rand.NormFloat64()*player.Stddev + player.ADP,
			})
		}
		sort.Slice(rankedPlayers[t], func(i, j int) bool { return rankedPlayers[t][i].ADP < rankedPlayers[t][j].ADP })
	}

	optStrategiesFn := func() []fantasy.Strategy {
		optRankedPlayers := make([][]fantasy.PlayerADP, numTeams)
		for t := 0; t < numTeams; t++ {
			for _, player := range state.Players {
				optRankedPlayers[t] = append(optRankedPlayers[t], fantasy.PlayerADP{
					PlayerID: player.ID,
					ADP:      rand.NormFloat64()*player.Stddev + player.ADP,
				})
			}
			sort.Slice(optRankedPlayers[t], func(i, j int) bool { return optRankedPlayers[t][i].ADP < optRankedPlayers[t][j].ADP })
		}
		optStrategies := make([]fantasy.Strategy, numTeams)
		for t := 0; t < numTeams; t++ {
			ch := strategyString[t]
			switch ch {
			case 'A':
				optStrategies[t] = fantasy.NewAutopick(order, rules)
			case 'H':
				optStrategies[t] = fantasy.NewHumanoid(order, rules, optRankedPlayers[t])
			case 'O':
				// Approximate Optimize with Humanoid.
				// TODO: Figure out a better approximation.
				optStrategies[t] = fantasy.NewHumanoid(order, rules, optRankedPlayers[t])
			default:
				log.Fatalf("Invalid strategy: %c", ch)
			}
		}
		return optStrategies
	}

	scorer := &fantasy.Scorer{[]byte(schema), *bench}

	strategies := make([]fantasy.Strategy, numTeams)
	for t := 0; t < numTeams; t++ {
		ch := strategyString[t]
		switch ch {
		case 'A':
			strategies[t] = fantasy.NewAutopick(order, rules)
		case 'H':
			strategies[t] = fantasy.NewHumanoid(order, rules, rankedPlayers[t])
		case 'O':
			strategies[t] = fantasy.NewOptimize(order, optStrategiesFn, rules, scorer, *numTrials)
		default:
			log.Fatalf("Invalid strategy: %c", ch)
		}
	}

	fantasy.RunDraft(state, order, strategies)

	for i, team := range state.Teams {
		fmt.Printf("Team #%d [%c] = %.2f\n", i, strategyString[i], scorer.Score(team))
		for _, player := range team.PlayersByPick() {
			fmt.Printf("  %s\n", player)
		}
	}
	for i, team := range state.Teams {
		fmt.Printf("Team #%2d [%c] = %8.2f\n", i, strategyString[i], scorer.Score(team))
	}
}
