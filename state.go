package fantasy

import (
	"sort"
)

type State struct {
	Teams          []*Team // drafted players
	UndraftedByVOR []*Player
	UndraftedByADP []*Player
	Pick           int
}

func ReadState(playersCsv string, numTeams int, order []int) (*State, error) {
	players, err := ReadPlayers(playersCsv)
	if err != nil {
		return nil, err
	}

	drafted := make(map[int]*Player)
	var undraftedByVOR, undraftedByADP []*Player
	for _, player := range players {
		if player.Pick == 0 {
			undraftedByVOR = append(undraftedByVOR, player)
			undraftedByADP = append(undraftedByADP, player)
		} else {
			drafted[player.Pick] = player
			drafted[player.Pick].Pick = 0 // we'll add it back later
		}
	}
	sort.Sort(sort.Reverse(ByVOR(undraftedByVOR)))
	sort.Sort(ByADP(undraftedByADP))

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
		Teams:          teams,
		UndraftedByVOR: undraftedByVOR,
		UndraftedByADP: undraftedByADP,
		Pick:           pick,
	}, nil
}
