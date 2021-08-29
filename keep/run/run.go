package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	dataDir = flag.String("data_dir", "", "directory for data files; empty string means $HOME/data")
)

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
