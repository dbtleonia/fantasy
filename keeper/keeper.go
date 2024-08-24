package keeper

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/stat/combin"
)

type GridderID int
type ManagerID int

type Gridder struct {
	Name  string
	Value float64
	MID   ManagerID // -1 if unowned
	Round int       //  0 if unowned

	// Allowed Picks in descending order. Does not include Picks
	// acquired via trade, which are not allowed to be used for keeping.
	Picks []int
}

type byGridderValue struct {
	o        []GridderID
	gridders []*Gridder
}

func (b byGridderValue) Len() int      { return len(b.o) }
func (b byGridderValue) Swap(i, j int) { b.o[i], b.o[j] = b.o[j], b.o[i] }
func (b byGridderValue) Less(i, j int) bool {
	return b.gridders[b.o[i]].Value > b.gridders[b.o[j]].Value
}

type byManagerName struct {
	o        []ManagerID
	managers []*Manager
}

func (b byManagerName) Len() int      { return len(b.o) }
func (b byManagerName) Swap(i, j int) { b.o[i], b.o[j] = b.o[j], b.o[i] }
func (b byManagerName) Less(i, j int) bool {
	return b.managers[b.o[i]].Name < b.managers[b.o[j]].Name
}

type Manager struct {
	Name string
	GIDs []GridderID
}

type Constants struct {
	Managers    []*Manager
	Gridders    []*Gridder
	picks       []ManagerID // includes picks acquired via trade
	gidsByValue []GridderID
	combos      [][]int

	// Only when doing reveal.
	// TODO: Possibly move into a separate type.
	// TODO: Validate actions on input.
	Ideal  []Action
	Actual []Action
}

type Keep struct {
	Pick int
	GID  GridderID
}

type Action []*Keep

func (a Action) findPick(pick int) (GridderID, bool) {
	for _, k := range a {
		if k.Pick == pick {
			return k.GID, true
		}
	}
	return -1, false
}

func (a Action) HasGID(gid GridderID) bool {
	for _, k := range a {
		if k.GID == gid {
			return true
		}
	}
	return false
}

func newConstants(gridders []*Gridder, managers []*Manager, picks []ManagerID, picksViaTrade []bool, ideal, actual []Action) *Constants {
	// Index gridder picks in descending order.
	managerPicks := make([][]int, len(managers))
	for j := len(picks) - 1; j >= 0; j-- {
		if picksViaTrade[j] {
			continue
		}
		mid := ManagerID(picks[j])
		managerPicks[mid] = append(managerPicks[mid], j)
	}
	for _, gridder := range gridders {
		if gridder.MID >= 0 {
			for _, pick := range managerPicks[gridder.MID] {
				if pick/len(managers) < gridder.Round {
					gridder.Picks = append(gridder.Picks, pick)
				}
			}
		}
	}

	gidsByValue := make([]GridderID, len(gridders))
	for i, _ := range gridders {
		gidsByValue[i] = GridderID(i)
	}
	sort.Sort(byGridderValue{gidsByValue, gridders})

	const maxKeepers = 3
	numRounds := len(picks) / len(managers)
	var combos [][]int
	for k := 0; k <= maxKeepers; k++ {
		combos = append(combos, combin.Combinations(numRounds, k)...)
	}

	return &Constants{managers, gridders, picks, gidsByValue, combos, ideal, actual}
}

func IteratedBestResponse(consts *Constants) [][]Action {
	// TODO: Don't hardcode length.
	profiles := [][]Action{
		{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
	}
	for i := 0; i < 20; i++ {
		prevProfile := profiles[len(profiles)-1]
		profile := make([]Action, len(consts.Managers))
		for m := 0; m < len(consts.Managers); m++ {
			profile[m] = bestResponse(consts, ManagerID(m), prevProfile)
		}
		profiles = append(profiles, profile)
	}

	return profiles[1:]
}

type ActionUtility struct {
	Act     Action
	Utility float64
}

func AllResponses(c *Constants, mid ManagerID, actions []Action) []*ActionUtility {
	var result []*ActionUtility

	// Instead of copying here, there are other options we could
	// investigate if performance is an issue.
	newActions := make([]Action, len(actions))
	for i, a := range actions {
		if ManagerID(i) != mid {
			newActions[i] = a
		}
	}

next_combo:
	for _, combo := range c.combos {
		response := Action{}
		for _, index := range combo {
			if index >= len(c.Managers[mid].GIDs) {
				continue next_combo
			}
			gid := c.Managers[mid].GIDs[index]

			allowedPick := -1
			for _, pick := range c.Gridders[gid].Picks {
				if _, ok := response.findPick(pick); !ok {
					allowedPick = pick
					break
				}
			}
			if allowedPick == -1 {
				continue next_combo
			}
			response = append(response, &Keep{allowedPick, gid})
		}
		newActions[mid] = response
		u := utilityOne(c, newActions, mid)
		result = append(result, &ActionUtility{response, u})
	}

	return result
}

func bestResponse(c *Constants, mid ManagerID, actions []Action) Action {
	responses := AllResponses(c, mid, actions)
	bestI := 0
	for i := 1; i < len(responses); i++ {
		if responses[i].Utility > responses[bestI].Utility {
			bestI = i
		}
	}
	return responses[bestI].Act
}

func utilityOne(c *Constants, actions []Action, mid1 ManagerID) float64 {
	result := 0.0
	utilityAccum(c, actions, func(pick int, mid ManagerID, gid GridderID, _ bool) {
		if mid == mid1 {
			result += c.Gridders[gid].Value
		}
	})
	return result
}

func UtilityAll(c *Constants, actions []Action) []float64 {
	result := make([]float64, len(actions))
	utilityAccum(c, actions, func(pick int, mid ManagerID, gid GridderID, _ bool) {
		result[mid] += c.Gridders[gid].Value
	})
	return result
}

func utilityAccum(c *Constants, actions []Action, accum func(pick int, mid ManagerID, gid GridderID, iskeep bool)) {
	next := 0
	for pick, mid := range c.picks {
		if gid, ok := actions[mid].findPick(pick); ok {
			accum(pick, mid, gid, true)
		} else {
			gid := c.gidsByValue[next]
			for {
				if c.Gridders[gid].MID == -1 || !actions[c.Gridders[gid].MID].HasGID(gid) {
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

func pickValues(c *Constants, actions []Action) []float64 {
	var result []float64
	utilityAccum(c, actions, func(pick int, mid ManagerID, gid GridderID, _ bool) {
		result = append(result, c.Gridders[gid].Value)
	})
	return result
}

func pickKeepers(c *Constants, actions []Action) []string {
	var result []string
	utilityAccum(c, actions, func(pick int, mid ManagerID, gid GridderID, iskeep bool) {
		if iskeep {
			result = append(result, c.Gridders[gid].Name)
		} else {
			result = append(result, "")
		}
	})
	return result
}

// Note: Not currently used.
func rosters(c *Constants, actions []Action) [][]GridderID {
	result := make([][]GridderID, len(actions))
	utilityAccum(c, actions, func(pick int, mid ManagerID, gid GridderID, _ bool) {
		result[mid] = append(result[mid], gid)
	})
	return result
}

func ReadConstants(dataDir string, reveal bool) (*Constants, error) {
	g, err := os.Open(path.Join(dataDir, "out", "player-values.csv"))
	if err != nil {
		return nil, err
	}
	defer g.Close()

	grecords, err := csv.NewReader(g).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("player values: %s", err)
	}

	var gridders []*Gridder
	gids := make(map[string]GridderID)
	for _, record := range grecords[1:] { // skip header
		gridderName := record[0]
		value, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, err
		}
		gids[gridderName] = GridderID(len(gridders))
		gridders = append(gridders, &Gridder{
			Name:  gridderName,
			Value: value,
			MID:   -1,
		})
	}

	k, err := os.Open(path.Join(dataDir, "out", "keeper-options.csv"))
	if err != nil {
		return nil, err
	}
	defer k.Close()

	krecords, err := csv.NewReader(k).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("keeper options: %s", err)
	}

	var managers []*Manager
	mids := make(map[string]ManagerID)
	for _, record := range krecords[1:] { // skip header
		managerName := record[0]
		playerName := record[1]

		round := 0
		if record[3] != "n/a" {
			round, err = strconv.Atoi(record[3])
			if err != nil {
				log.Fatal(err)
			}
		}

		mid, ok := mids[managerName]
		if !ok {
			mid = ManagerID(len(managers))
			mids[managerName] = mid
			managers = append(managers, &Manager{
				Name: managerName,
			})
		}
		gid := gids[playerName]
		managers[mid].GIDs = append(managers[mid].GIDs, gid)
		gridders[gid].MID = mid
		gridders[gid].Round = round
	}

	o, err := os.Open(path.Join(dataDir, "yahoo", "draft-order.csv"))
	if err != nil {
		return nil, err
	}
	defer o.Close()

	orecords, err := csv.NewReader(o).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("draft order: %s", err)
	}

	var picks []ManagerID
	var picksViaTrade []bool
	for _, record := range orecords[1:] { // skip header
		pick, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatal(err)
		}
		managerName := record[1]
		viaTrade := strings.TrimSpace(record[2]) != ""

		if pick != len(picks)+1 {
			log.Fatalf("Out of order pick: %d\n", pick)
		}

		mid, ok := mids[managerName]
		if !ok {
			log.Fatalf("No manager with name %q", managerName)
		}

		picks = append(picks, mid)
		picksViaTrade = append(picksViaTrade, viaTrade)
	}

	// Only for reveal mode.
	ideal := make([]Action, len(mids))
	actual := make([]Action, len(mids))
	if reveal {
		id, err := os.Open(path.Join(dataDir, "out", "keeper-ideal.csv"))
		if err != nil {
			return nil, err
		}
		defer id.Close()

		idrecords, err := csv.NewReader(id).ReadAll()
		if err != nil {
			return nil, fmt.Errorf("keeper ideal: %s", err)
		}

		for _, record := range idrecords[1:] { // skip header
			managerName := record[0]
			gridderName := record[1]
			roundPick := record[2]

			mid, ok := mids[managerName]
			if !ok {
				log.Fatalf("No manager with name %q", managerName)
			}

			gid, ok := gids[gridderName]
			if !ok {
				log.Fatalf("No gridder with name %q", gridderName)
			}

			dash := strings.Index(roundPick, "-")
			pick, err := strconv.Atoi(roundPick[dash+1:])
			if err != nil {
				log.Fatal(err)
			}

			ideal[mid] = append(ideal[mid], &Keep{pick - 1, gid})
		}

		a, err := os.Open(path.Join(dataDir, "managers", "keeper-selections.csv"))
		if err != nil {
			return nil, err
		}
		defer a.Close()

		arecords, err := csv.NewReader(a).ReadAll()
		if err != nil {
			return nil, fmt.Errorf("keeper selections: %s", err)
		}

		for _, record := range arecords[1:] { // skip header
			managerName := record[0]
			gridderName := record[1]
			roundPick := record[2]

			mid, ok := mids[managerName]
			if !ok {
				log.Fatalf("No manager with name %q", managerName)
			}

			gid, ok := gids[gridderName]
			if !ok {
				log.Fatalf("No gridder with name %q", gridderName)
			}

			dash := strings.Index(roundPick, "-")
			pick, err := strconv.Atoi(roundPick[dash+1:])
			if err != nil {
				log.Fatal(err)
			}

			// TODO: Check that this manager owns this pick.
			round, err := strconv.Atoi(roundPick[:dash])
			if err != nil {
				log.Fatal(err)
			}
			if (pick-1)/12 != round-1 {
				log.Fatal(roundPick)
			}

			actual[mid] = append(actual[mid], &Keep{pick - 1, gid})
		}
	}

	return newConstants(gridders, managers, picks, picksViaTrade, ideal, actual), nil
}
