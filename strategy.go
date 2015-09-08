package fantasy

import (
	"math/rand"
	"sort"
)

type Strategy interface {
	Select(state *State, order []int) int
}

type Autopick struct {
	rules *Rules
}

func NewAutopick(rules *Rules) *Autopick {
	return &Autopick{rules}
}

func (a *Autopick) Select(state *State, order []int) int {
	i := order[state.Pick]
	team := state.Teams[i]
	allowedPos := a.rules.AutopickMap[team.PosString()]
	for j, player := range state.Undrafted {
		if allowedPos[player.Pos[0]] {
			return j
		}
	}
	return 0
}

type Humanoid struct {
	rules  *Rules
	lambda float64
}

func NewHumanoid(rules *Rules, lambda float64) *Humanoid {
	return &Humanoid{rules, lambda}
}

func (h *Humanoid) Select(state *State, order []int) int {
	i := order[state.Pick]
	team := state.Teams[i]
	allowedPos := h.rules.HumanoidMap[team.PosString()]
	r := int(rand.ExpFloat64() / h.lambda)
	for j, player := range state.Undrafted {
		if allowedPos[player.Pos[0]] {
			r--
		}
		if r < 0 {
			return j
		}
	}
	return 0
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

func posLeaders(undrafted []*Player) map[string]int {
	leaders := make(map[string]int)
	for j, player := range undrafted {
		if len(leaders) == len("DKQRTW") {
			break
		}
		if _, present := leaders[player.Pos]; !present {
			leaders[player.Pos] = j
		}
	}
	return leaders
}

type Candidate struct {
	Index int
	Value float64
}
type ByValue []*Candidate

func (x ByValue) Len() int           { return len(x) }
func (x ByValue) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ByValue) Less(i, j int) bool { return x[i].Value < x[j].Value }

func (o *Optimize) Candidates(state *State, order []int) []*Candidate {
	i := order[state.Pick]
	team := state.Teams[i]
	var result []*Candidate
	for _, j := range posLeaders(state.Undrafted) {
		player := state.Undrafted[j]
		if !o.rules.HumanoidMap[team.PosString()][player.Pos[0]] {
			continue // in theory can remove this check
		}
		points := 0.0
		for trial := 0; trial < o.numTrials; trial++ {
			undrafted := clonePlayers(state.Undrafted)
			teams := cloneTeams(state.Teams)
			teams[i].Add(player, state.Pick)
			RunDraft(&State{teams, undrafted, state.Pick + 1}, order, o.strategies)
			points += o.scorer.Score(teams[i])
		}
		result = append(result, &Candidate{j, points})
	}
	sort.Sort(ByValue(result))
	return result
}

func (o *Optimize) Select(state *State, order []int) int {
	c := o.Candidates(state, order)
	return c[len(c)-1].Index
}
