package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	charts  = flag.Bool("charts", true, "output HTML with charts of results")
	dataDir = flag.String("data_dir", "", "directory for data files; empty string means $HOME/data")
)

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
		round, err := strconv.Atoi(record[2])
		if err != nil {
			log.Fatal(err)
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

	return Constants(gridders, managers), nil
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
