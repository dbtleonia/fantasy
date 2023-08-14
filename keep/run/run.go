package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

var (
	reveal  = flag.Bool("reveal", true, "make report for reveal")
	charts  = flag.Bool("charts", false, "output HTML with charts of results")
	dataDir = flag.String("data_dir", "", "directory for data files; empty string means $HOME/data")
)

func boolString(b bool, s string) string {
	if b {
		return s
	}
	return strings.Repeat(" ", len(s))
}

func numBetterThan(utils []float64, mid managerid, util float64) int {
	n := 0
	for otherMid, otherUtil := range utils {
		if managerid(otherMid) != mid && otherUtil > util {
			n++
		}
	}
	return n
}

func makeRanks(n int) []int {
	ranks := make([]int, n)
	for i := 0; i < n; i++ {
		ranks[i] = i
	}
	return ranks
}

func ord(n int) string {
	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	}
	return fmt.Sprintf("%dth", n)
}

func printCharts(c *constants, profile []action) {
	nokeep := make([]action, len(profile))

	// TODO: Translate to short manager names.
	managerNames := make([]string, len(c.managers))
	for i, m := range c.managers {
		managerNames[i] = m.name[:9]
	}
	managersPre := utilityAll(c, nokeep)
	managersPost := utilityAll(c, profile)

	var pickNames []string
	utilityAccum(c, nokeep, func(pick int, mid managerid, gid gridderid, iskeep bool) {
		round := pick/len(profile) + 1
		pickNames = append(pickNames, fmt.Sprintf("%s %d-%d", managerNames[mid], round, pick+1))
	})
	picksPre := pickValues(c, nokeep)
	picksPost := pickValues(c, profile)
	keepers := pickKeepers(c, profile)

	tmpl := template.Must(template.ParseFiles("tmpl.html"))
	data := map[string]interface{}{
		"managerNames": managerNames,
		"managersPre":  managersPre,
		"managersPost": managersPost,
		"pickNames":    pickNames,
		"picksPre":     picksPre,
		"picksPost":    picksPost,
		"keepers":      keepers,
	}
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		log.Fatal(err)
	}
}

func ReadConstants(dataDir string) (*constants, error) {
	g, err := os.Open(path.Join(dataDir, "projections.csv"))
	if err != nil {
		return nil, err
	}
	defer g.Close()

	grecords, err := csv.NewReader(g).ReadAll()
	if err != nil {
		return nil, err
	}

	var gridders []*gridder
	gids := make(map[string]gridderid)
	for _, record := range grecords[1:] { // skip header
		gridderName := record[0]
		value, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, err
		}
		gids[gridderName] = gridderid(len(gridders))
		gridders = append(gridders, &gridder{
			name:  gridderName,
			value: value,
			mid:   -1,
		})
	}

	k, err := os.Open(path.Join(dataDir, "keeper_rounds.csv"))
	if err != nil {
		return nil, err
	}
	defer k.Close()

	krecords, err := csv.NewReader(k).ReadAll()
	if err != nil {
		return nil, err
	}

	var managers []*manager
	mids := make(map[string]managerid)
	for _, record := range krecords[1:] { // skip header
		managerName := record[0]
		playerName := record[1]

		round := 0
		if record[6] != "n/a" {
			round, err = strconv.Atoi(record[6])
			if err != nil {
				log.Fatal(err)
			}
		}

		mid, ok := mids[managerName]
		if !ok {
			mid = managerid(len(managers))
			mids[managerName] = mid
			managers = append(managers, &manager{
				name: managerName,
			})
		}
		gid := gids[playerName]
		managers[mid].gids = append(managers[mid].gids, gid)
		gridders[gid].mid = mid
		gridders[gid].round = round
	}

	o, err := os.Open(path.Join(dataDir, "draft_order.csv"))
	if err != nil {
		return nil, err
	}
	defer o.Close()

	orecords, err := csv.NewReader(o).ReadAll()
	if err != nil {
		return nil, err
	}

	var picks []managerid
	var picksViaTrade []bool
	for _, record := range orecords[1:] { // skip header
		pick, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatal(err)
		}
		managerName := record[1]
		viaTrade := strings.TrimSpace(record[2]) != ""

		if pick != len(picks)+1 {
			log.Fatal("Out of order pick: %d\n", pick)
		}

		mid, ok := mids[managerName]
		if !ok {
			log.Fatalf("No manager with name %q", managerName)
		}

		picks = append(picks, mid)
		picksViaTrade = append(picksViaTrade, viaTrade)
	}

	id, err := os.Open(path.Join(dataDir, "keeper_ideal.csv"))
	if err != nil {
		return nil, err
	}
	defer id.Close()

	idrecords, err := csv.NewReader(id).ReadAll()
	if err != nil {
		return nil, err
	}

	ideal := make([]action, len(mids))
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
			log.Fatal("No gridder with name %q", gridderName)
		}

		dash := strings.Index(roundPick, "-")
		pick, err := strconv.Atoi(roundPick[dash+1:])
		if err != nil {
			log.Fatal(err)
		}

		ideal[mid] = append(ideal[mid], &keep{pick - 1, gid})
	}

	a, err := os.Open(path.Join(dataDir, "keeper_selections.csv"))
	if err != nil {
		return nil, err
	}
	defer a.Close()

	arecords, err := csv.NewReader(a).ReadAll()
	if err != nil {
		return nil, err
	}

	actual := make([]action, len(mids))
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
			log.Fatal("No gridder with name %q", gridderName)
		}

		dash := strings.Index(roundPick, "-")
		pick, err := strconv.Atoi(roundPick[dash+1:])
		if err != nil {
			log.Fatal(err)
		}

		actual[mid] = append(actual[mid], &keep{pick - 1, gid})
	}

	return Constants(gridders, managers, picks, picksViaTrade, ideal, actual), nil
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	dir := *dataDir
	if dir == "" {
		if home := os.Getenv("HOME"); home != "" {
			dir = path.Join(home, "data")
		}
	}

	consts, err := ReadConstants(dir)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Pull this out into a separate command.
	if *reveal {
		utilsStart := utilityAll(consts, make([]action, len(consts.managers)))
		utilsIdeal := utilityAll(consts, consts.ideal)
		utilsActual := utilityAll(consts, consts.actual)

		utilsBest := make([]float64, len(consts.managers))
		scores := make([]int, len(consts.managers))
		gridderDiffs := make([]float64, len(consts.gridders))
		gridderBest := make([]bool, len(consts.gridders))

		for mid, _ := range consts.managers {
			all := allResponses(consts, managerid(mid), consts.actual)

			// Find utility for null action.
			var u0 float64
			for _, au := range all {
				if len(au.act) == 0 {
					u0 = au.utility
					break
				}
			}

			// Find best action.
			bestI := 0
			for i := 1; i < len(all); i++ {
				if all[i].utility > all[bestI].utility {
					bestI = i
				}
			}
			bestAU := all[bestI]

			// For single-gridder actions, record stats.
			for _, au := range all {
				if len(au.act) != 1 {
					continue
				}
				gid := au.act[0].gid // only 1 player in this action
				gridderDiffs[gid] = au.utility - u0
				gridderBest[gid] = bestAU.act.hasGID(gid)
			}
			scores[mid] = int(math.Round(100 * (utilsActual[mid] - u0) / (bestAU.utility - u0)))
			utilsBest[mid] = bestAU.utility
		}

		for m, manager := range consts.managers {
			mid := managerid(m)
			rankStart := 1 + numBetterThan(utilsStart, mid, utilsStart[mid])
			rankIdeal := 1 + numBetterThan(utilsIdeal, mid, utilsIdeal[mid])
			rankActual := 1 + numBetterThan(utilsActual, mid, utilsActual[mid])
			// Yes, utilsActual.  This is not truly a rank.
			rankBest := 1 + numBetterThan(utilsActual, mid, utilsBest[mid])

			fmt.Printf("+++++++++++++++ %s ++++++++++++++\n\n", manager.name)
			for _, g := range manager.gids {
				gid := gridderid(g)
				if consts.gridders[gid].round == 0 {
					// Gridder was round 1 keeper last year; not eligible this year.
					fmt.Printf("%30s n/a\n",
						consts.gridders[gid].name)
				} else if len(consts.gridders[gid].picks) == 0 {
					// Gridder can't be kept because manager has no pick that round or earlier.
					fmt.Printf("%30s R%2d  n/a\n",
						consts.gridders[gid].name,
						consts.gridders[gid].round)
				} else {
					fmt.Printf("%30s R%2d %4d %s %s\n",
						consts.gridders[gid].name,
						consts.gridders[gid].round,
						int(gridderDiffs[gid]),
						boolString(gridderBest[gid], "BEST"),
						boolString(consts.actual[mid].hasGID(gid), "ACTUAL"))
				}
			}
			fmt.Printf("\n")
			fmt.Printf("SCORE = %d\n", scores[mid])
			fmt.Printf("\n")
			fmt.Printf("Start rank  = %s\n", ord(rankStart))
			fmt.Printf("Ideal rank  = %s\n", ord(rankIdeal))
			fmt.Printf("Actual rank = %s\n", ord(rankActual))
			fmt.Printf("Best rank   = %s\n", ord(rankBest))
			fmt.Printf("\n")
		}

		fmt.Printf("================= REPORT CARD ==============\n\n")

		ranksStart := makeRanks(len(consts.managers))  // []int{0, .., 11}
		ranksIdeal := makeRanks(len(consts.managers))  // []int{0, .., 11}
		ranksActual := makeRanks(len(consts.managers)) // []int{0, .., 11}
		sort.Slice(ranksStart, func(i, j int) bool {
			return utilsStart[ranksStart[i]] > utilsStart[ranksStart[j]]
		})
		sort.Slice(ranksIdeal, func(i, j int) bool {
			return utilsIdeal[ranksIdeal[i]] > utilsIdeal[ranksIdeal[j]]
		})
		sort.Slice(ranksActual, func(i, j int) bool {
			return utilsActual[ranksActual[i]] > utilsActual[ranksActual[j]]
		})

		fmt.Printf("%4s %15s %15s %15s\n", "rank", "start", "ideal", "actual")
		fmt.Printf("%4s %15s %15s %15s\n", "====", "=====", "=====", "======")
		for i := 0; i < len(consts.managers); i++ {
			fmt.Printf("%4s %15s %15s %15s\n",
				ord(i+1),
				consts.managers[ranksStart[i]].name,
				consts.managers[ranksIdeal[i]].name,
				consts.managers[ranksActual[i]].name)
		}

		return
	}

	profiles := iteratedProfiles(consts)

	// TODO: Pull this out into a separate command.
	if *charts {
		printCharts(consts, profiles[len(profiles)-1])
		return
	}

	// Output the results.  This currently loops through the entire
	// response for each gridder.  We could make it more efficient if
	// necessary.
	for g, gridder := range consts.gridders {
		if gridder.mid == -1 {
			continue
		}
		managerName := consts.managers[gridder.mid].name
		if len(managerName) > 20 {
			managerName = managerName[:20]
		}
		fmt.Printf("%-20s %35s (%6.1f) @ %2d ", managerName, gridder.name, gridder.value, gridder.round)
		for _, profile := range profiles {
			var round string
			for _, keeps := range profile {
				for _, k := range keeps {
					if gridderid(g) == k.gid {
						round = strings.ToUpper(strconv.FormatInt(int64(k.pick/len(profile)+1), 36))
					}
				}
			}
			if len(round) > 0 {
				fmt.Printf("%s", round)
			} else {
				fmt.Printf(".")
			}
		}
		fmt.Printf("\n")
	}
}
