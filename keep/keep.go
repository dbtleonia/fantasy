package keep

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sort"
)

type gridderid int
type managerid int

type gridder struct {
	name  string
	value float64
	mid   managerid // -1 if unowned
	round int       //  0 if unowned
	picks []int     // allowed picks in descending order
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
	picks       []managerid
	gidsByValue []gridderid
}

type keep struct {
	pick int
	gid  gridderid
}

type action []keep

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

func Handle(w http.ResponseWriter, r *http.Request) {
	if err := Run(r.Body, w); err != nil {
		log.Printf("%v", err)
	}
}

func Run(body io.Reader, w io.Writer) error {
	var rows [][]interface{}
	if err := json.NewDecoder(body).Decode(&rows); err != nil {
		return err
	}

	// Set up gridders & managers.
	var gridders []*gridder
	var managers []*manager
	managerids := make(map[string]managerid)
	var maxRound int
	for g, row := range rows[1:] { // skip header
		gridderName, ok := row[0].(string)
		if !ok {
			return fmt.Errorf("field 0 not a string in %v", row)
		}
		value, ok := row[1].(float64)
		if !ok {
			return fmt.Errorf("field 1 not a float64 in %v", row)
		}
		managerName, ok := row[2].(string)
		if !ok {
			return fmt.Errorf("field 2 not a string in %v", row)
		}
		mid := managerid(-1)
		var round int
		if managerName != "" {
			var ok bool
			mid, ok = managerids[managerName]
			if !ok {
				mid = managerid(len(managers)) // generate new mid
				managerids[managerName] = mid
				managers = append(managers, &manager{
					name: managerName,
				})
			}
			managers[mid].gids = append(managers[mid].gids, gridderid(g))

			roundFloat, ok := row[3].(float64)
			if !ok {
				return fmt.Errorf("field 3 is not a float64 in %v", row)
			}
			round = int(roundFloat)
			if round > maxRound {
				maxRound = round
			}
		}
		gridders = append(gridders, &gridder{
			name:  gridderName,
			value: value,
			mid:   mid,
			round: round,
		})
	}

	// Set up picks.  Assume picks are snake draft in alphabetical order
	// of manager names.
	order := make([]managerid, len(managers))
	for m := 0; m < len(managers); m++ {
		order[m] = managerid(m)
	}
	sort.Sort(byManagerName{order, managers})

	var picks []managerid
	for i := 0; i < (maxRound+1)/2; i++ {
		for _, mid := range order {
			picks = append(picks, mid)
		}
		for j := len(order) - 1; j >= 0; j-- {
			picks = append(picks, order[j])
		}
	}

	// Index gridder picks in descending order.
	managerPicks := make([][]int, len(managers))
	for j := len(picks) - 1; j >= 0; j-- {
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

	consts := &constants{managers, gridders, picks, gidsByValue}
	profiles := iteratedProfiles(consts)

	result := make([][][]interface{}, len(profiles))
	for i, profile := range profiles {
		for _, action := range profile {
			for _, k := range action {
				result[i] = append(result[i], []interface{}{k.gid + 1, "X"})
			}
		}
	}

	return json.NewEncoder(w).Encode(result)
}

func iteratedProfiles(consts *constants) [][]action {
	profiles := [][]action{
		{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
	}
	for i := 0; i < 30; i++ {
		prevProfile := profiles[len(profiles)-1]
		profile := make([]action, len(consts.managers))
		for m := 0; m < len(consts.managers); m++ {
			profile[m] = bestResponse(consts, managerid(m), prevProfile)
		}
		profiles = append(profiles, profile)
	}

	return profiles[1:]
}

func bestResponse(c *constants, mid managerid, actions []action) action {
	// Instead of copying here, there are other options we could
	// investigate if performance is an issue.
	newActions := make([]action, len(actions))
	for i, a := range actions {
		if managerid(i) != mid {
			newActions[i] = a
		}
	}
	var response action
	prevU := math.Inf(-1)
	for k := 0; k < 4; k++ {
		bestU := math.Inf(-1)
		bestPick := 0
		bestGid := gridderid(0)
		for _, gid := range c.managers[mid].gids {
			if response.hasGID(gid) {
				continue
			}
			allowedPick := -1
			for _, pick := range c.gridders[gid].picks {
				if _, ok := response.findPick(pick); !ok {
					allowedPick = pick
					break
				}
			}
			if allowedPick == -1 {
				continue
			}
			newActions[mid] = append(action{{allowedPick, gid}}, response...)
			u := utilityOne(c, newActions, mid)
			if u > bestU {
				bestU = u
				bestPick = allowedPick
				bestGid = gid
			}
		}
		if bestU < prevU {
			break
		}
		prevU = bestU
		response = append(response, keep{bestPick, bestGid})
	}
	return response
}

func utilityOne(c *constants, actions []action, mid1 managerid) float64 {
	result := 0.0
	utilityAccum(c, actions, func(pick int, mid managerid, gid gridderid) {
		if mid == mid1 {
			result += c.gridders[gid].value
		}
	})
	return result
}

func utilityAccum(c *constants, actions []action, accum func(pick int, mid managerid, gid gridderid)) {
	next := 0
	for pick, mid := range c.picks {
		if gid, ok := actions[mid].findPick(pick); ok {
			accum(pick, mid, gid)
		} else {
			gid := c.gidsByValue[next]
			for {
				if c.gridders[gid].mid == -1 || !actions[c.gridders[gid].mid].hasGID(gid) {
					break
				}
				next++
				gid = c.gidsByValue[next]
			}
			accum(pick, mid, gid)
			next++
		}
	}
}
