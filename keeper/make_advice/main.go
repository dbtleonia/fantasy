package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/dbtleonia/fantasy/keeper"
)

var (
	ideal   = flag.Bool("ideal", false, "make ideal CSV")
	dataDir = flag.String("data_dir", "", "directory for data files; empty string means $HOME/data")
)

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	dir := *dataDir
	if dir == "" {
		if home := os.Getenv("HOME"); home != "" {
			dir = path.Join(home, "data")
		}
	}

	consts, err := keeper.ReadConstants(dir, false)
	if err != nil {
		log.Fatal(err)
	}

	profiles := keeper.IteratedBestResponse(consts)

	// Output the results.  This currently loops through the entire
	// response for each gridder.  We could make it more efficient if
	// necessary.
	if *ideal {
		// TODO: Use CSV writer.
		fmt.Printf("manager,player,pick\n")
	}
	for g, gridder := range consts.Gridders {
		if gridder.MID == -1 {
			continue
		}
		managerName := consts.Managers[gridder.MID].Name
		if len(managerName) > 20 {
			managerName = managerName[:20]
		}
		if !*ideal {
			fmt.Printf("%-20s %35s (%6.1f) @ %2d ", managerName, gridder.Name, gridder.Value, gridder.Round)
		}
		for i, profile := range profiles {
			pick := -1
			for _, keeps := range profile {
				for _, k := range keeps {
					if keeper.GridderID(g) == k.GID {
						pick = k.Pick
					}
				}
			}
			if *ideal {
				if i == len(profiles)-1 && pick >= 0 {
					fmt.Printf("%s,%s,%d-%d\n", managerName, gridder.Name, pick/12+1, pick+1)
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
