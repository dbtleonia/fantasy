package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

var (
	nullADPCutoff    = flag.Int("null_adp_cutoff", 400, "fatal error if player with null ADP has rank less than cutoff")
	missingByeCutoff = flag.Int("missing_bye_cutoff", 10, "fatal error if more than this number of players is missing a bye")
)

func readByes(filename string) (map[string]string, error) {
	const (
		colName = 0
		colPos  = 1
		colTeam = 2
		colBye  = 3
	)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(f)
	if _, err := r.ReadString('\n'); err != nil { // discard the first line
		return nil, err
	}
	in := csv.NewReader(r)
	byes := make(map[string]string) // key is <name><pos><team>
	for {
		record, err := in.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		key := record[colName] + record[colPos] + record[colTeam]
		byes[key] = record[colBye]
	}
	return byes, nil
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 || flag.NArg() > 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <custom-rankings-csv> [<raw-stat-projections-csv>]", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	var byes map[string]string // key is <name><pos><team>
	if flag.NArg() == 2 {
		var err error
		byes, err = readByes(flag.Arg(1))
		if err != nil {
			log.Fatal(err)
		}
	}
	const (
		colID      = 0
		colName    = 2
		colPos     = 3
		colTeam    = 4
		colVOR     = 7
		colPoints  = 8
		colRank    = 11
		colPosRank = 12
		colADP     = 16
		colCeiling = 19
	)
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	r := bufio.NewReader(f)
	if _, err := r.ReadString('\n'); err != nil { // discard the first line
		log.Fatal(err)
	}
	in := csv.NewReader(r)
	out := csv.NewWriter(os.Stdout)
	missingByeCount := 0
	for {
		record, err := in.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if record[colADP] == "null" {
			rank, err := strconv.Atoi(record[colRank])
			if err != nil {
				log.Fatal(err)
			}
			if rank < *nullADPCutoff {
				log.Fatalf("Player with rank %d (below cutoff %d) has null ADP: %v", rank, *nullADPCutoff, record)
			}
			continue // these players ain't gonna be drafted anyhow
		}
		var bye string
		if byes != nil {
			key := record[colName] + record[colPos] + record[colTeam]
			var ok bool
			bye, ok = byes[key]
			if !ok {
				missingByeCount++
			}
		}
		out.Write([]string{
			"0",
			record[colID],
			record[colName],
			record[colPos],
			record[colTeam],
			record[colVOR],
			record[colPoints],
			record[colRank],
			record[colPosRank],
			record[colADP],
			record[colCeiling],
			bye,
		})
	}
	out.Flush()
	if missingByeCount > *missingByeCutoff {
		log.Fatalf("Found %d players missing byes, more than cutoff %d", missingByeCount, *missingByeCutoff)
	}
}
