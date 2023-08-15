package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	dummy   = flag.Int("dummy", 10, "number of dummy players to generate for each position")
	keepers = flag.String("keepers", "", "keepers file named keepers-<nteams>.csv")
)

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
		for _, record := range records {
			nteams, err := strconv.Atoi(strings.TrimSuffix(strings.Split(*keepers, "-")[1], ".csv"))
			if err != nil {
				log.Fatal(err)
			}
			parts := strings.Split(record[1], "-")
			round, err := strconv.Atoi(parts[0])
			if err != nil {
				log.Fatal(err)
			}
			slot, err := strconv.Atoi(parts[1])
			if err != nil {
				log.Fatal(err)
			}
			keeperPicks[record[2]] = nteams*(round-1) + slot
		}
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	in := csv.NewReader(f)
	in.Comma = '\t'
	records, err := in.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	const (
		colName  = 1
		colValue = 2
	)
	var out [][]string
	var problems []string
	poscount := make(map[string]int)
	for i, record := range records {
		var pos string
		switch {
		case strings.Contains(record[colName], "- DST"):
			pos = "DST"
		case strings.Contains(record[colName], "- K"):
			pos = "K"
		case strings.Contains(record[colName], "- QB"):
			pos = "QB"
		case strings.Contains(record[colName], "- RB"):
			pos = "RB"
		case strings.Contains(record[colName], "- TE"):
			pos = "TE"
		case strings.Contains(record[colName], "- WR"):
			pos = "WR"
		default:
			problems = append(problems, record[colName])
		}
		poscount[pos]++

		// TODO: Remove dummy values once sim/opt no longer need them.
		out = append(out, []string{
			strconv.Itoa(keeperPicks[record[colName]]), // pick
			strconv.Itoa(10000 + i),                    // id
			record[colName],                            // name
			pos,                                        // pos
			"XXX",                                      // team
			strings.TrimPrefix(record[colValue], "$"), // value
			"999",                       // points
			strconv.Itoa(i + 1),         // rank
			strconv.Itoa(poscount[pos]), // posrank
			strconv.Itoa(i + 1),         // adp
			strings.TrimPrefix(record[colValue], "$"), // ceiling
			"", // bye
		})
	}
	if len(problems) > 0 {
		log.Fatalf("Unknown positions:  \n  %s\n", strings.Join(problems, "\n  "))
	}
	for j, pos := range []string{"DST", "K", "QB", "RB", "TE", "WR"} {
		for i := 0; i < *dummy; i++ {
			poscount[pos]++
			n := len(out)
			out = append(out, []string{
				"0",
				strconv.Itoa(20000 + 10000*j + i), // id
				fmt.Sprintf("%sdummy <%s> #%d", pos[:1], pos, i), // name
				pos,                         // pos
				"XXX",                       // team
				"0",                         // value
				"0",                         // points
				strconv.Itoa(n + 1),         // rank
				strconv.Itoa(poscount[pos]), // posrank
				strconv.Itoa(n + 1),         // adp
				"0",                         // ceiling
				"",                          // bye
			})
		}
	}
	if err = csv.NewWriter(os.Stdout).WriteAll(out); err != nil {
		log.Fatal(err)
	}
}
