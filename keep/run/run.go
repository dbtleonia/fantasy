package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

var (
	ideal   = flag.Bool("ideal", false, "make ideal CSV")
	reveal  = flag.Bool("reveal", true, "make report for reveal")
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

func ReadConstants(dataDir string) (*constants, error) {
	g, err := os.Open(path.Join(dataDir, "player_values.csv"))
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

	k, err := os.Open(path.Join(dataDir, "keeper_options.csv"))
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
		if record[3] != "n/a" {
			round, err = strconv.Atoi(record[3])
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
	ideal := make([]action, len(mids))
	actual := make([]action, len(mids))
	if *reveal {
		id, err := os.Open(path.Join(dataDir, "keeper_ideal.csv"))
		if err != nil {
			return nil, err
		}
		defer id.Close()

		idrecords, err := csv.NewReader(id).ReadAll()
		if err != nil {
			return nil, err
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

			actual[mid] = append(actual[mid], &keep{pick - 1, gid})
		}
	}

	return newConstants(gridders, managers, picks, picksViaTrade, ideal, actual), nil
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if flag.NArg() != 1 {
		log.Fatal("year please")
	}
	year := flag.Arg(0)

	dir := *dataDir
	if dir == "" {
		if home := os.Getenv("HOME"); home != "" {
			dir = path.Join(home, "data")
		}
	}

	consts, err := ReadConstants(path.Join(dir, "out", year))
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

	profiles := iteratedBestResponse(consts)

	// Output the results.  This currently loops through the entire
	// response for each gridder.  We could make it more efficient if
	// necessary.
	if *ideal {
		// TODO: Use CSV writer.
		fmt.Printf("manager,player,pick\n")
	}
	for g, gridder := range consts.gridders {
		if gridder.mid == -1 {
			continue
		}
		managerName := consts.managers[gridder.mid].name
		if len(managerName) > 20 {
			managerName = managerName[:20]
		}
		if !*ideal {
			fmt.Printf("%-20s %35s (%6.1f) @ %2d ", managerName, gridder.name, gridder.value, gridder.round)
		}
		for i, profile := range profiles {
			pick := -1
			for _, keeps := range profile {
				for _, k := range keeps {
					if gridderid(g) == k.gid {
						pick = k.pick
					}
				}
			}
			if *ideal {
				if i == len(profiles)-1 && pick >= 0 {
					fmt.Printf("%s,%s,%d-%d\n", managerName, gridder.name, pick/12+1, pick+1)
				}
			} else {
				if pick >= 0 {
					fmt.Print(strings.ToUpper(strconv.FormatInt(int64(pick/len(profile)+1), 36)))
				} else {
					fmt.Printf(".")
				}
			}
		}
		if !*ideal {
			fmt.Printf("\n")
		}
	}
}
