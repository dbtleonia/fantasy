package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/dbtleonia/fantasy/keeper"
)

var (
	dataDir = flag.String("data_dir", "", "directory for data files; empty string means $HOME/data")
)

func boolString(b bool, s string) string {
	if b {
		return s
	}
	return strings.Repeat(" ", len(s))
}

func numBetterThan(utils []float64, mid keeper.ManagerID, util float64) int {
	n := 0
	for otherMid, otherUtil := range utils {
		if keeper.ManagerID(otherMid) != mid && otherUtil > util {
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

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	dir := *dataDir
	if dir == "" {
		if home := os.Getenv("HOME"); home != "" {
			dir = path.Join(home, "data")
		}
	}

	consts, err := keeper.ReadConstants(dir, true)
	if err != nil {
		log.Fatal(err)
	}

	utilsStart := keeper.UtilityAll(consts, make([]keeper.Action, len(consts.Managers)))
	utilsIdeal := keeper.UtilityAll(consts, consts.Ideal)
	utilsActual := keeper.UtilityAll(consts, consts.Actual)

	utilsBest := make([]float64, len(consts.Managers))
	scores := make([]int, len(consts.Managers))
	gridderDiffs := make([]float64, len(consts.Gridders))
	gridderBest := make([]bool, len(consts.Gridders))

	for mid, _ := range consts.Managers {
		all := keeper.AllResponses(consts, keeper.ManagerID(mid), consts.Actual)

		// Find utility for null action.
		var u0 float64
		for _, au := range all {
			if len(au.Act) == 0 {
				u0 = au.Utility
				break
			}
		}

		// Find best action.
		bestI := 0
		for i := 1; i < len(all); i++ {
			if all[i].Utility > all[bestI].Utility {
				bestI = i
			}
		}
		bestAU := all[bestI]

		// For single-gridder actions, record stats.
		for _, au := range all {
			if len(au.Act) != 1 {
				continue
			}
			gid := au.Act[0].GID // only 1 player in this action
			gridderDiffs[gid] = au.Utility - u0
			gridderBest[gid] = bestAU.Act.HasGID(gid)
		}
		scores[mid] = int(math.Round(100 * (utilsActual[mid] - u0) / (bestAU.Utility - u0)))
		utilsBest[mid] = bestAU.Utility
	}

	for m, manager := range consts.Managers {
		mid := keeper.ManagerID(m)
		rankStart := 1 + numBetterThan(utilsStart, mid, utilsStart[mid])
		rankIdeal := 1 + numBetterThan(utilsIdeal, mid, utilsIdeal[mid])
		rankActual := 1 + numBetterThan(utilsActual, mid, utilsActual[mid])
		// Yes, utilsActual.  This is not truly a rank.
		rankBest := 1 + numBetterThan(utilsActual, mid, utilsBest[mid])

		fmt.Printf("+++++++++++++++ %s ++++++++++++++\n\n", manager.Name)
		for _, g := range manager.GIDs {
			gid := keeper.GridderID(g)
			if consts.Gridders[gid].Round == 0 {
				// Gridder was round 1 keeper last year; not eligible this year.
				fmt.Printf("%32s n/a\n",
					consts.Gridders[gid].Name)
			} else if len(consts.Gridders[gid].Picks) == 0 {
				// Gridder can't be kept because manager has no pick that round or earlier.
				fmt.Printf("%32s R%2d  n/a\n",
					consts.Gridders[gid].Name,
					consts.Gridders[gid].Round)
			} else {
				fmt.Printf("%32s R%2d %4d %s %s\n",
					consts.Gridders[gid].Name,
					consts.Gridders[gid].Round,
					int(gridderDiffs[gid]),
					boolString(gridderBest[gid], "BEST"),
					boolString(consts.Actual[mid].HasGID(gid), "ACTUAL"))
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

	ranksStart := makeRanks(len(consts.Managers))  // []int{0, .., 11}
	ranksIdeal := makeRanks(len(consts.Managers))  // []int{0, .., 11}
	ranksActual := makeRanks(len(consts.Managers)) // []int{0, .., 11}
	sort.Slice(ranksStart, func(i, j int) bool {
		return utilsStart[ranksStart[i]] > utilsStart[ranksStart[j]]
	})
	sort.Slice(ranksIdeal, func(i, j int) bool {
		return utilsIdeal[ranksIdeal[i]] > utilsIdeal[ranksIdeal[j]]
	})
	sort.Slice(ranksActual, func(i, j int) bool {
		return utilsActual[ranksActual[i]] > utilsActual[ranksActual[j]]
	})

	fmt.Printf("%4s   %-15s   %-15s   %-15s\n", "rank", "start", "ideal", "actual")
	fmt.Printf("%4s   %-15s   %-15s   %-15s\n", "====", "=====", "=====", "======")
	for i := 0; i < len(consts.Managers); i++ {
		fmt.Printf("%4s   %-15s   %-15s   %-15s\n",
			ord(i+1),
			truncate(consts.Managers[ranksStart[i]].Name, 15),
			truncate(consts.Managers[ranksIdeal[i]].Name, 15),
			truncate(consts.Managers[ranksActual[i]].Name, 15))
	}
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
