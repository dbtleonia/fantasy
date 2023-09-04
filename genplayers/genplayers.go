package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

var (
	dummy   = flag.Int("dummy", 10, "number of dummy players to generate for each position")
	keepers = flag.String("keepers", "", "keepers file")
	adpDir  = flag.String("adp", "", "directory with ADP values")
)

// TODO: Dedupe with similar function in keeper code.
func mustReadAll(filename string) [][]string {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	return records
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <projections-tsv>", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	keeperPicks := make(map[string]int) // player -> pick
	if *keepers != "" {
		k, err := os.Open(*keepers)
		if err != nil {
			log.Fatal(err)
		}
		defer k.Close()
		records, err := csv.NewReader(k).ReadAll()
		if err != nil {
			log.Fatal(err)
		}
		for _, record := range records[1:] {
			parts := strings.Split(record[2], "-")
			pick, err := strconv.Atoi(parts[1])
			if err != nil {
				log.Fatal(err)
			}
			keeperPicks[record[1]] = pick
		}
	}

	type adp struct {
		mean   float64
		stddev float64
	}
	playerADP := make(map[string]adp)
	if *adpDir != "" {
		missing := make(map[string]bool)
		for _, record := range mustReadAll(path.Join(*adpDir, "missing.csv")) {
			missing[record[0]] = true
		}

		renames := make(map[string]string)
		for _, record := range mustReadAll(path.Join(*adpDir, "renames.csv")) {
			renames[record[0]] = record[1]
		}

		records := mustReadAll(path.Join(*adpDir, "adp.csv"))
		for _, record := range records[1:] {
			name := record[2]
			mean, err := strconv.ParseFloat(record[1], 64)
			if err != nil {
				log.Fatal(err)
			}
			stddev, err := strconv.ParseFloat(record[6], 64)
			if err != nil {
				log.Fatal(err)
			}

			if _, ok := missing[name]; ok {
				continue
			}

			if n, ok := renames[name]; ok {
				name = n
			}

			playerADP[name] = adp{mean, stddev}
		}
	}

	files, err := os.ReadDir(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	j := 0
	players := make(map[string][]string)
	var problems []string
	for _, file := range files {
		f, err := os.Open(path.Join(flag.Arg(0), file.Name()))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		in := csv.NewReader(f)
		in.FieldsPerRecord = -1 // allow bad records
		records, err := in.ReadAll()
		if err != nil {
			log.Fatal(err)
		}

		const (
			colName = 0
		)
		colPoints := len(records[0]) - 1
		for _, record := range records[1:] {
			if len(record) != len(records[0]) {
				continue // skip bad records
			}
			name := record[colName]
			points, err := strconv.ParseFloat(record[colPoints], 64)
			if err != nil {
				log.Fatal(err)
			}
			// TODO: Use a struct rather than an array.
			if player, ok := players[name]; ok {
				v, _ := strconv.ParseFloat(player[5], 64)
				if v > points {
					continue
				}
			}
			var pos string
			switch {
			case strings.Contains(file.Name(), "_DST"):
				pos = "DST"
			case strings.Contains(file.Name(), "_K"):
				pos = "K"
			case strings.Contains(file.Name(), "_QB"):
				pos = "QB"
			case strings.Contains(file.Name(), "_RB"):
				pos = "RB"
			case strings.Contains(file.Name(), "_TE"):
				pos = "TE"
			case strings.Contains(file.Name(), "_WR"):
				pos = "WR"
			default:
				problems = append(problems, file.Name())
			}

			keeperPick, ok := keeperPicks[record[colName]]
			if ok {
				delete(keeperPicks, record[colName])
			}

			pADP, ok := playerADP[record[colName]]
			if ok {
				delete(playerADP, record[colName])
			} else {
				pADP = adp{300.0, 20.0}
			}

			players[name] = []string{
				strconv.Itoa(keeperPick),         // pick
				strconv.Itoa(10000 + j),          // id
				record[colName],                  // name
				pos,                              // pos
				"XXX",                            // team
				record[colPoints],                // points
				fmt.Sprintf("%.1f", pADP.mean),   // adp mean
				fmt.Sprintf("%.1f", pADP.stddev), // adp stddev
			}
			j++
		}
	}
	if len(problems) > 0 {
		log.Fatalf("Unknown positions:  \n  %s\n", strings.Join(problems, "\n  "))
	}
	if len(keeperPicks) > 0 {
		var ps []string
		for p := range keeperPicks {
			ps = append(ps, p)
		}
		sort.Strings(ps)
		log.Fatalf("Keeper not used for:  \n  %s\n", strings.Join(ps, "\n  "))
	}
	if len(playerADP) > 0 {
		var ps []string
		for p := range playerADP {
			ps = append(ps, p)
		}
		sort.Strings(ps)
		log.Fatalf("ADP not used for:  \n  %s\n", strings.Join(ps, "\n  "))
	}
	var out [][]string
	for _, player := range players {
		out = append(out, player)
	}
	sort.Slice(out, func(i, j int) bool {
		// TODO: Same here, this should use a struct and not re-parse
		// every time.
		a, _ := strconv.ParseFloat(out[i][5], 64)
		b, _ := strconv.ParseFloat(out[j][5], 64)
		return a > b
	})

	// Append dummy players.
	for j, pos := range []string{"DST", "K", "QB", "RB", "TE", "WR"} {
		for i := 0; i < *dummy; i++ {
			out = append(out, []string{
				"0",                               // pick
				strconv.Itoa(20000 + 10000*j + i), // id
				fmt.Sprintf("%sdummy <%s> #%d", pos[:1], pos, i), // name
				pos,     // pos
				"XXX",   // team
				"0",     // points
				"300.0", // adp mean
				"20.0",  // adp stddev
			})
		}
	}
	if err = csv.NewWriter(os.Stdout).WriteAll(out); err != nil {
		log.Fatal(err)
	}
}
