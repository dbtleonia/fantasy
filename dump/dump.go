package main

import (
	"fmt"
	"log"
	"os"

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
	var errors []string
	drafted := make(map[int]*fantasy.Player)
	maxPick := 0
	for _, player := range players {
		if player.Pick == 0 {
			continue
		}
		if _, present := drafted[player.Pick]; present {
			errors = append(errors, fmt.Sprintf("Pick %d has multiple players: %s, %s", player.Pick, drafted[player.Pick].Name, player.Name))
		}
		drafted[player.Pick] = player
		if player.Pick > maxPick {
			maxPick = player.Pick
		}
	}
	foundNextPick := false
	for i := 1; i <= maxPick; i++ {
		if player, ok := drafted[i]; ok {
			fmt.Printf("%s\n", player)
		} else {
			if !foundNextPick {
				foundNextPick = true
				fmt.Printf(">>>>>>>>>>>>>>> next pick = %d\n", i)
			}
		}
	}
	for _, e := range errors {
		fmt.Printf("*** %s\n", e)
	}
}
