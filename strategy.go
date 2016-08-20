package fantasy

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
)

type Strategy interface {
	Select(state *State, order []int) (player *Player, justification string)
}

type Autopick struct {
	rules *Rules
}

func NewAutopick(rules *Rules) *Autopick {
	return &Autopick{rules}
}

func (a *Autopick) Select(state *State, order []int) (*Player, string) {
	i := order[state.Pick]
	team := state.Teams[i]
	allowedPos := a.rules.AutopickMap[team.PosString()]
	for _, player := range state.Undrafted {
		if allowedPos[player.Pos[0]] {
			return player, a.rules.AutopickRaw[team.PosString()]
		}
	}
	return state.Undrafted[0], ""
}

type Humanoid struct {
	rules  *Rules
	lambda float64
}

func NewHumanoid(rules *Rules, lambda float64) *Humanoid {
	return &Humanoid{rules, lambda}
}

func (h *Humanoid) Select(state *State, order []int) (*Player, string) {
	i := order[state.Pick]
	team := state.Teams[i]
	allowedPos := h.rules.HumanoidMap[team.PosString()]
	r := int(rand.ExpFloat64() / h.lambda)
	justification := fmt.Sprintf("%-6s reached %d", h.rules.HumanoidRaw[team.PosString()], r)
	for _, player := range state.Undrafted {
		if allowedPos[player.Pos[0]] {
			r--
		}
		if r < 0 {
			return player, justification
		}
	}
	return state.Undrafted[0], ""
}

type Optimize struct {
	rules      *Rules
	strategies []Strategy
	scorer     *Scorer
	numTrials  int
}

func NewOptimize(rules *Rules, strategies []Strategy, scorer *Scorer, numTrials int) *Optimize {
	return &Optimize{rules, strategies, scorer, numTrials}
}

func posLeaders(undrafted []*Player) map[string]*Player {
	leaders := make(map[string]*Player)
	for _, player := range undrafted {
		if len(leaders) == len("DKQRTW") {
			break
		}
		if _, present := leaders[player.Pos]; !present {
			leaders[player.Pos] = player
		}
	}
	return leaders
}

type Candidate struct {
	Player *Player
	Value  float64
}
type ByValue []*Candidate

func (x ByValue) Len() int           { return len(x) }
func (x ByValue) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ByValue) Less(i, j int) bool { return x[i].Value < x[j].Value }

func (o *Optimize) Candidates(state *State, order []int) []*Candidate {
	i := order[state.Pick]
	var result []*Candidate
	for _, player := range posLeaders(state.Undrafted) {
		points := 0.0
		for trial := 0; trial < o.numTrials; trial++ {
			undrafted := removePlayer(clonePlayers(state.Undrafted), player.ID)
			teams := cloneTeams(state.Teams)
			teams[i].Add(player, state.Pick, "")
			RunDraft(&State{teams, undrafted, state.Pick + 1}, order, o.strategies)
			points += o.scorer.Score(teams[i])
		}
		result = append(result, &Candidate{player, points})
	}
	sort.Sort(sort.Reverse(ByValue(result)))
	return result
}

func (o *Optimize) Select(state *State, order []int) (*Player, string) {
	candidates := o.Candidates(state, order)

	var justification []string
	for _, c := range candidates {
		justification = append(justification, fmt.Sprintf("%c%02d=%d", c.Player.Pos[0], c.Player.PosRank, int(c.Value)))
	}

	return candidates[0].Player, strings.Join(justification, " ")
}
