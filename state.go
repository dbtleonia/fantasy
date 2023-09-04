package fantasy

import (
	"sort"
)

type State struct {
	Teams             []*Team // drafted players
	UndraftedByPoints []*Player
	Pick              int

	Drafted map[int]bool
	Players map[int]*Player
}

func ReadState(playersCsv string, numTeams int, order []int) (*State, []int, error) {
	players, err := ReadPlayers(playersCsv)
	if err != nil {
		return nil, nil, err
	}

	drafted := make(map[int]*Player)
	var undraftedByPoints []*Player
	for _, player := range players {
		if player.Pick == 0 {
			undraftedByPoints = append(undraftedByPoints, player)
		} else {
			drafted[player.Pick] = player
			drafted[player.Pick].Pick = 0 // we'll add it back later
		}
	}
	sort.Sort(sort.Reverse(ByPoints(undraftedByPoints)))

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
			teams[i].Add(player, pk, "*** KEEPER ***")
			newOrder[pk] = -1
		}
	}

	draftedBool := make(map[int]bool)
	for _, player := range drafted {
		draftedBool[player.ID] = true
	}

	playerMap := make(map[int]*Player)
	for _, p := range players {
		playerMap[p.ID] = p
	}

	return &State{
		Teams:             teams,
		UndraftedByPoints: undraftedByPoints,
		Pick:              pick,
		Drafted:           draftedBool,
		Players:           playerMap,
	}, newOrder, nil
}

func (st *State) Clone() *State {
	drafted := make(map[int]bool)
	for id := range st.Drafted {
		drafted[id] = true
	}
	return &State{
		Teams:             cloneTeams(st.Teams),
		UndraftedByPoints: clonePlayers(st.UndraftedByPoints),
		Pick:              st.Pick,
		Drafted:           drafted,
		Players:           st.Players,
	}
}

func (st *State) Update(team int, player *Player, justification string) {
	st.Teams[team].Add(player, st.Pick, justification)
	st.UndraftedByPoints = removePlayer(st.UndraftedByPoints, player.ID)
	st.Drafted[player.ID] = true
}
