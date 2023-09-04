package fantasy

import (
	"fmt"
	"sort"
	"strings"
)

type Strategy interface {
	Select(state *State) (player *Player, justification string)
}

type Autopick struct {
	order []int
	rules *Rules
}

func NewAutopick(order []int, rules *Rules) *Autopick {
	return &Autopick{order, rules}
}

func (a *Autopick) Select(state *State) (*Player, string) {
	i := a.order[state.Pick]
	team := state.Teams[i]
	allowedPos := a.rules.AutopickMap[team.PosString()]
	// TODO: Use ADP instead.
	for _, player := range state.UndraftedByPoints {
		if allowedPos[player.Pos[0]] {
			return player, a.rules.AutopickRaw[team.PosString()]
		}
	}
	return state.UndraftedByPoints[0], ""
}

type PlayerADP struct {
	PlayerID int
	ADP      float64
}

type Humanoid struct {
	order         []int
	rules         *Rules
	rankedPlayers []PlayerADP
}

func NewHumanoid(order []int, rules *Rules, rankedPlayers []PlayerADP) *Humanoid {
	return &Humanoid{order, rules, rankedPlayers}
}

func (h *Humanoid) Select(state *State) (*Player, string) {
	i := h.order[state.Pick]
	team := state.Teams[i]
	allowedPos := h.rules.HumanoidMap[team.PosString()]
	for _, want := range h.rankedPlayers {
		if !state.Drafted[want.PlayerID] && allowedPos[state.Players[want.PlayerID].Pos[0]] {
			return state.Players[want.PlayerID], fmt.Sprintf("adp = %5.1f, pos = %s, allowed = %s", want.ADP, team.PosString(), h.rules.HumanoidRaw[team.PosString()])
		}
	}
	return state.UndraftedByPoints[0], ""
}

type Optimize struct {
	order      []int
	strategies func() []Strategy
	rules      *Rules
	scorer     *Scorer
	numTrials  int
}

func NewOptimize(order []int, strategies func() []Strategy, rules *Rules, scorer *Scorer, numTrials int) *Optimize {
	return &Optimize{order, strategies, rules, scorer, numTrials}
}

func posLeaders(undrafted []*Player) []*Player {
	n := 0
	leaders := make(map[string][]*Player)
	for _, player := range undrafted {
		if len(leaders[player.Pos]) < 3 {
			leaders[player.Pos] = append(leaders[player.Pos], player)
			n++
		}
		if n == 18 {
			break
		}
	}
	var result []*Player
	for _, players := range leaders {
		result = append(result, players...)
	}
	return result
}

type Candidate struct {
	Player *Player
	Score  float64
}
type ByScore []*Candidate

func (x ByScore) Len() int           { return len(x) }
func (x ByScore) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ByScore) Less(i, j int) bool { return x[i].Score < x[j].Score }

func (o *Optimize) Candidates(state *State) []*Candidate {
	i := o.order[state.Pick]
	var result []*Candidate
	for _, player := range posLeaders(state.UndraftedByPoints) {
		score := 0.0
		for trial := 0; trial < o.numTrials; trial++ {
			newState := state.Clone()
			newState.Update(i, player, "")
			newState.Pick++
			RunDraft(newState, o.order, o.strategies())
			score += o.scorer.Score(newState.Teams[i])
		}
		result = append(result, &Candidate{player, score})
	}
	sort.Sort(sort.Reverse(ByScore(result)))
	return result
}

func (o *Optimize) Select(state *State) (*Player, string) {
	candidates := o.Candidates(state)

	var justification []string
	for _, c := range candidates {
		justification = append(justification, fmt.Sprintf("%c%.1f=%d", c.Player.Pos[0], c.Player.ADP, int(c.Score)))
	}

	return candidates[0].Player, strings.Join(justification, " ")
}
