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
	nullADPCutoff = flag.Int("null", 400, "fatal error if player with null ADP has rank less than cutoff")
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <custom-rankings-csv>", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
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
		})
	}
	out.Flush()
}
