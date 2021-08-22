package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	// Inputs:
	//   <data_dir>/merged.csv
	dataDir = flag.String("data_dir", "", "directory containing data sources")

	normalizePoints = flag.Bool("normalize_points", false, "normalize points")
	printValues     = flag.Bool("print_values", false, "print values")
)

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	f, err := os.Open(path.Join(*dataDir, "merged.csv"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	reader := csv.NewReader(f)
	if _, err = reader.Read(); err != nil { // ignore header
		log.Fatal(err)
	}
	var gridders []*gridder
	var managers []*manager
	managerids := make(map[string]managerid)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		gridderName := record[0]
		projection, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Fatal(err)
		}
		var value float64
		if *normalizePoints {
			switch {
			case strings.Contains(record[0], "- DEF"):
				value = projection - 110.0
			case strings.Contains(record[0], "- K"):
				value = projection - 135.0
			case strings.Contains(record[0], "- QB"):
				value = projection - 255.0
			case strings.Contains(record[0], "- RB"):
				value = projection - 110.0
			case strings.Contains(record[0], "- TE"):
				value = projection - 70.0
			case strings.Contains(record[0], "- WR"):
				value = projection - 85.0
			default:
				log.Fatal("Unknown pos in %v", record[0])
			}
		} else {
			value = projection
		}
		managerName := record[2]
		mid := managerid(-1)
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
			gid := gridderid(len(gridders)) // this will be gid for new gridder
			managers[mid].gids = append(managers[mid].gids, gid)
		}

		var round int
		if record[3] == "" {
			round = 0
		} else {
			round, err = strconv.Atoi(record[3])
			if err != nil {
				log.Fatal(err)
			}
		}

		gridders = append(gridders, &gridder{
			name:  gridderName,
			value: value,
			mid:   mid,
			round: round,
		})
	}

	// Call the library function.
	profiles, err := Run(gridders, managers)
	if err != nil {
		log.Fatal(err)
	}

	// Output the results.  This currently loops through the entire
	// response for each gridder.  We could make it more efficient if
	// necessary.
	for g, gridder := range gridders {
		if gridder.mid == -1 {
			continue
		}
		managerName := managers[gridder.mid].name
		if len(managerName) > 20 {
			managerName = managerName[:20]
		}
		fmt.Printf("%-20s %35s (%6.1f) @ %2d ", managerName, gridder.name, gridder.value, gridder.round)
		for _, profile := range profiles {
			var round string
			var value float64
			for _, keeps := range profile {
				for _, k := range keeps {
					if gridderid(g) == k.gid {
						round = fmt.Sprintf("%X", (k.pick/len(profile) + 1))
						value = k.value
					}
				}
			}
			if *printValues {
				if len(round) > 0 {
					fmt.Printf("%4.1f ", value)
				} else {
					fmt.Printf(".... ")
				}
			} else {
				if len(round) > 0 {
					fmt.Printf("X")
				} else {
					fmt.Printf(".")
				}
			}
		}
		fmt.Printf("\n")
	}
}
