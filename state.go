package fantasy

import (
	"sort"
)

type State struct {
	Teams     []*Team // drafted players
	Undrafted []*Player
	Pick      int
}

func ReadState(playersCsv string, numTeams int, order []int) (*State, error) {
	players, err := ReadPlayers(playersCsv)
	if err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(ByVOR(players)))

	drafted := make(map[int]*Player)
	var undrafted []*Player
	for _, player := range players {
		if player.Pick == 0 {
			undrafted = append(undrafted, player)
		} else {
			drafted[player.Pick] = player
			drafted[player.Pick].Pick = 0 // we'll add it back later
		}
	}

	teams := make([]*Team, numTeams)
	for i := 0; i < numTeams; i++ {
		teams[i] = &Team{}
	}

	pick := 1
	for {
		player, ok := drafted[pick]
		if !ok {
			break
		}
		i := order[pick]
		teams[i].Add(player, pick, "")
		pick++
	}

	return &State{
		Teams:     teams,
		Undrafted: undrafted,
		Pick:      pick,
	}, nil
}
