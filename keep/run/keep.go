package main

import (
	"sort"

	"gonum.org/v1/gonum/stat/combin"
)

type gridderid int
type managerid int

type gridder struct {
	name  string
	value float64
	mid   managerid // -1 if unowned
	round int       //  0 if unowned

	// Allowed picks in descending order. Does not include picks
	// acquired via trade, which are not allowed to be used for keeping.
	picks []int
}

type byGridderValue struct {
	o        []gridderid
	gridders []*gridder
}

func (b byGridderValue) Len() int      { return len(b.o) }
func (b byGridderValue) Swap(i, j int) { b.o[i], b.o[j] = b.o[j], b.o[i] }
func (b byGridderValue) Less(i, j int) bool {
	return b.gridders[b.o[i]].value > b.gridders[b.o[j]].value
}

type byManagerName struct {
	o        []managerid
	managers []*manager
}

func (b byManagerName) Len() int      { return len(b.o) }
func (b byManagerName) Swap(i, j int) { b.o[i], b.o[j] = b.o[j], b.o[i] }
func (b byManagerName) Less(i, j int) bool {
	return b.managers[b.o[i]].name < b.managers[b.o[j]].name
}

type manager struct {
	name string
	gids []gridderid
}

type constants struct {
	managers    []*manager
	gridders    []*gridder
	picks       []managerid // includes picks acquired via trade
	gidsByValue []gridderid
	combos      [][]int

	// Only when doing reveal.
	// TODO: Possibly move into a separate type.
	// TODO: Validate actions on input.
	ideal  []action
	actual []action
}

type keep struct {
	pick int
	gid  gridderid
}

type action []*keep

func (a action) findPick(pick int) (gridderid, bool) {
	for _, k := range a {
		if k.pick == pick {
			return k.gid, true
		}
	}
	return -1, false
}

func (a action) hasGID(gid gridderid) bool {
	for _, k := range a {
		if k.gid == gid {
			return true
		}
	}
	return false
}

func newConstants(gridders []*gridder, managers []*manager, picks []managerid, picksViaTrade []bool, ideal, actual []action) *constants {
	// Index gridder picks in descending order.
	managerPicks := make([][]int, len(managers))
	for j := len(picks) - 1; j >= 0; j-- {
		if picksViaTrade[j] {
			continue
		}
		mid := managerid(picks[j])
		managerPicks[mid] = append(managerPicks[mid], j)
	}
	for _, gridder := range gridders {
		if gridder.mid >= 0 {
			for _, pick := range managerPicks[gridder.mid] {
				if pick/len(managers) < gridder.round {
					gridder.picks = append(gridder.picks, pick)
				}
			}
		}
	}

	gidsByValue := make([]gridderid, len(gridders))
	for i, _ := range gridders {
		gidsByValue[i] = gridderid(i)
	}
	sort.Sort(byGridderValue{gidsByValue, gridders})

	const maxKeepers = 3
	numRounds := len(picks) / len(managers)
	var combos [][]int
	for k := 0; k <= maxKeepers; k++ {
		combos = append(combos, combin.Combinations(numRounds, k)...)
	}

	return &constants{managers, gridders, picks, gidsByValue, combos, ideal, actual}
}

func iteratedBestResponse(consts *constants) [][]action {
	// TODO: Don't hardcode length.
	profiles := [][]action{
		{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
	}
	for i := 0; i < 20; i++ {
		prevProfile := profiles[len(profiles)-1]
		profile := make([]action, len(consts.managers))
		for m := 0; m < len(consts.managers); m++ {
			profile[m] = bestResponse(consts, managerid(m), prevProfile)
		}
		profiles = append(profiles, profile)
	}

	return profiles[1:]
}

type actionUtility struct {
	act     action
	utility float64
}

func allResponses(c *constants, mid managerid, actions []action) []*actionUtility {
	var result []*actionUtility

	// Instead of copying here, there are other options we could
	// investigate if performance is an issue.
	newActions := make([]action, len(actions))
	for i, a := range actions {
		if managerid(i) != mid {
			newActions[i] = a
		}
	}

next_combo:
	for _, combo := range c.combos {
		response := action{}
		for _, index := range combo {
			if index >= len(c.managers[mid].gids) {
				continue next_combo
			}
			gid := c.managers[mid].gids[index]

			allowedPick := -1
			for _, pick := range c.gridders[gid].picks {
				if _, ok := response.findPick(pick); !ok {
					allowedPick = pick
					break
				}
			}
			if allowedPick == -1 {
				continue next_combo
			}
			response = append(response, &keep{allowedPick, gid})
		}
		newActions[mid] = response
		u := utilityOne(c, newActions, mid)
		result = append(result, &actionUtility{response, u})
	}

	return result
}

func bestResponse(c *constants, mid managerid, actions []action) action {
	responses := allResponses(c, mid, actions)
	bestI := 0
	for i := 1; i < len(responses); i++ {
		if responses[i].utility > responses[bestI].utility {
			bestI = i
		}
	}
	return responses[bestI].act
}

func utilityOne(c *constants, actions []action, mid1 managerid) float64 {
	result := 0.0
	utilityAccum(c, actions, func(pick int, mid managerid, gid gridderid, _ bool) {
		if mid == mid1 {
			result += c.gridders[gid].value
		}
	})
	return result
}

func utilityAll(c *constants, actions []action) []float64 {
	result := make([]float64, len(actions))
	utilityAccum(c, actions, func(pick int, mid managerid, gid gridderid, _ bool) {
		result[mid] += c.gridders[gid].value
	})
	return result
}

func utilityAccum(c *constants, actions []action, accum func(pick int, mid managerid, gid gridderid, iskeep bool)) {
	next := 0
	for pick, mid := range c.picks {
		if gid, ok := actions[mid].findPick(pick); ok {
			accum(pick, mid, gid, true)
		} else {
			gid := c.gidsByValue[next]
			for {
				if c.gridders[gid].mid == -1 || !actions[c.gridders[gid].mid].hasGID(gid) {
					break
				}
				next++
				gid = c.gidsByValue[next]
			}
			accum(pick, mid, gid, false)
			next++
		}
	}
}

func pickValues(c *constants, actions []action) []float64 {
	var result []float64
	utilityAccum(c, actions, func(pick int, mid managerid, gid gridderid, _ bool) {
		result = append(result, c.gridders[gid].value)
	})
	return result
}

func pickKeepers(c *constants, actions []action) []string {
	var result []string
	utilityAccum(c, actions, func(pick int, mid managerid, gid gridderid, iskeep bool) {
		if iskeep {
			result = append(result, c.gridders[gid].name)
		} else {
			result = append(result, "")
		}
	})
	return result
}

// Note: Not currently used.
func rosters(c *constants, actions []action) [][]gridderid {
	result := make([][]gridderid, len(actions))
	utilityAccum(c, actions, func(pick int, mid managerid, gid gridderid, _ bool) {
		result[mid] = append(result[mid], gid)
	})
	return result
}
