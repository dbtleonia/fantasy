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

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <projections-tsv>", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	in := csv.NewReader(f)
	in.Comma = '\t'
	records, err := in.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	const (
		colName   = 1
		colPoints = 2
		colValue  = 3
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
			"0",
			strconv.Itoa(10000 + i), // id
			record[colName],         // name
			pos,                     // pos
			"XXX",                   // team
			strings.TrimPrefix(record[colValue], "$"), // value
			record[colPoints],                         // points
			strconv.Itoa(i + 1),                       // rank
			strconv.Itoa(poscount[pos]),               // posrank
			strconv.Itoa(i + 1),                       // adp
			strings.TrimPrefix(record[colValue], "$"), // ceiling
			"", // bye
		})
	}
	if len(problems) > 0 {
		log.Fatalf("Unknown positions:  \n  %s\n", strings.Join(problems, "\n  "))
	}
	if err = csv.NewWriter(os.Stdout).WriteAll(out); err != nil {
		log.Fatal(err)
	}
}
