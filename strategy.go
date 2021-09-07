package fantasy

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
)

type Strategy interface {
	Select(state *State) (player *Player, justification string)
}

type Autopick struct {
	order  []int
	rules  *Rules
	useADP bool
}

func NewAutopick(order []int, rules *Rules, useADP bool) *Autopick {
	return &Autopick{order, rules, useADP}
}

func (a *Autopick) Select(state *State) (*Player, string) {
	i := a.order[state.Pick]
	team := state.Teams[i]
	allowedPos := a.rules.AutopickMap[team.PosString()]
	undrafted := state.UndraftedByValue
	if a.useADP {
		undrafted = state.UndraftedByADP
	}
	for _, player := range undrafted {
		if allowedPos[player.Pos[0]] {
			return player, a.rules.AutopickRaw[team.PosString()]
		}
	}
	return undrafted[0], ""
}

type Humanoid struct {
	order  []int
	rules  *Rules
	useADP bool
	lambda float64
}

func NewHumanoid(order []int, rules *Rules, useADP bool, lambda float64) *Humanoid {
	return &Humanoid{order, rules, useADP, lambda}
}

func (h *Humanoid) Select(state *State) (*Player, string) {
	i := h.order[state.Pick]
	team := state.Teams[i]
	allowedPos := h.rules.HumanoidMap[team.PosString()]
	r := int(rand.ExpFloat64() / h.lambda)
	justification := fmt.Sprintf("%-6s reached %d", h.rules.HumanoidRaw[team.PosString()], r)
	undrafted := state.UndraftedByValue
	if h.useADP {
		undrafted = state.UndraftedByADP
	}
	for _, player := range undrafted {
		if allowedPos[player.Pos[0]] {
			r--
		}
		if r < 0 {
			return player, justification
		}
	}
	return undrafted[0], ""
}

type Optimize struct {
	order      []int
	strategies []Strategy
	rules      *Rules
	scorer     *Scorer
	numTrials  int
}

func NewOptimize(order []int, strategies []Strategy, rules *Rules, scorer *Scorer, numTrials int) *Optimize {
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
	for _, player := range posLeaders(state.UndraftedByValue) {
		score := 0.0
		for trial := 0; trial < o.numTrials; trial++ {
			undraftedByValue := removePlayer(clonePlayers(state.UndraftedByValue), player.ID)
			undraftedByADP := removePlayer(clonePlayers(state.UndraftedByADP), player.ID)
			teams := cloneTeams(state.Teams)
			teams[i].Add(player, state.Pick, "")
			RunDraft(&State{teams, undraftedByValue, undraftedByADP, state.Pick + 1}, o.order, o.strategies)
			score += o.scorer.Score(teams[i])
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
		justification = append(justification, fmt.Sprintf("%c%02d=%d", c.Player.Pos[0], c.Player.PosRank, int(c.Score)))
	}

	return candidates[0].Player, strings.Join(justification, " ")
}
