package fantasy

import (
	"log"
	"sort"
)

type Team struct {
	players   []*Player // sorted by value descending
	positions string    // sorted ascending eg "DQRWW"
}

// Add copies a player onto the team, setting its pick field.
func (t *Team) Add(player *Player, pick int, justification string) {
	if player.Pick != 0 {
		log.Fatalf("Add player with Pick != 0: %s", player)
	}

	// Sort by value descending.
	i := 0
	for ; i < len(t.players) && player.Value < t.players[i].Value; i++ {
	}
	playerCopy := *player
	playerCopy.Pick = pick
	playerCopy.Justification = justification
	// https://github.com/golang/go/wiki/SliceTricks
	t.players = append(t.players, nil)
	copy(t.players[i+1:], t.players[i:])
	t.players[i] = &playerCopy

	// Sort by letters Low to High.
	ch := player.Pos[0]
	j := 0
	for ; j < len(t.positions) && ch > t.positions[j]; j++ {
	}
	t.positions = t.positions[:j] + string(ch) + t.positions[j:]
}

// PlayersByValue returns the players sorted by value
// descending.  Callers should not modify the returned list.
func (t *Team) PlayersByValue() []*Player {
	return t.players
}

// PlayersByPick returns a copy of the players list sorted by pick in
// ascending order.
func (t *Team) PlayersByPick() []*Player {
	players := append([]*Player(nil), t.players...)
	sort.Sort(ByPick(players))
	return players
}

// PosString returns a string representing the players' positions, one
// char per player, sorted in ascending order.
func (t *Team) PosString() string {
	return t.positions
}

func (t *Team) Clone() *Team {
	return &Team{
		players:   clonePlayers(t.players),
		positions: t.positions,
	}
}

func cloneTeams(teams []*Team) []*Team {
	result := make([]*Team, len(teams))
	for i, team := range teams {
		result[i] = team.Clone()
	}
	return result
}
