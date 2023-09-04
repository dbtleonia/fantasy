package fantasy

import (
	"sort"
)

type State struct {
	Teams             []*Team // drafted players
	UndraftedByPoints []*Player
	UndraftedByADP    []*Player
	Pick              int
}

func ReadState(playersCsv string, numTeams int, order []int) (*State, []int, error) {
	players, err := ReadPlayers(playersCsv)
	if err != nil {
		return nil, nil, err
	}

	drafted := make(map[int]*Player)
	var undraftedByPoints, undraftedByADP []*Player
	for _, player := range players {
		if player.Pick == 0 {
			undraftedByPoints = append(undraftedByPoints, player)
			undraftedByADP = append(undraftedByADP, player)
		} else {
			drafted[player.Pick] = player
			drafted[player.Pick].Pick = 0 // we'll add it back later
		}
	}
	sort.Sort(sort.Reverse(ByPoints(undraftedByPoints)))
	sort.Sort(ByADP(undraftedByADP))

	teams := make([]*Team, numTeams)
	for i := 0; i < numTeams; i++ {
		teams[i] = &Team{}
	}

	pick := 1
	for {
		if _, ok := drafted[pick]; !ok {
			break
		}
		pick++
	}

	newOrder := make([]int, len(order))
	for pk, o := range order {
		newOrder[pk] = o
	}

	for pk, i := range order {
		if pk == 0 {
			continue
		}
		if player, ok := drafted[pk]; ok {
			teams[i].Add(player, pk, "")
			newOrder[pk] = -1
		}
	}

	return &State{
		Teams:             teams,
		UndraftedByPoints: undraftedByPoints,
		UndraftedByADP:    undraftedByADP,
		Pick:              pick,
	}, newOrder, nil
}
