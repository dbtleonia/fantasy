package main

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <projections-csv>", os.Args[0])
	}
	const (
		colID     = 0
		colName   = 2
		colPos    = 3
		colPoints = 8
	)
	f, err := os.Open(os.Args[1])
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
		out.Write([]string{
			"0",
			record[colID],
			record[colName],
			record[colPos],
			record[colPoints],
		})
	}
	out.Flush()
}
