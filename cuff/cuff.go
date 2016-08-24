package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/dbtleonia/fantasy"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <players-csv>", os.Args[0])
	}
	players, err := fantasy.ReadPlayers(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	byTeam := make(map[string][]*fantasy.Player)
	for _, player := range players {
		if player.Pos == "RB" && player.Team != "FA" {
			byTeam[player.Team] = append(byTeam[player.Team], player)
		}
	}
	var teams []string
	for t, _ := range byTeam {
		teams = append(teams, t)
	}
	sort.Strings(teams)
	for _, t := range teams {
		fmt.Printf("%s\n", t)
		for _, p := range byTeam[t] {
			fmt.Printf("  %s\n", p)
		}
	}
}
