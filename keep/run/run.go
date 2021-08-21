package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/dbtleonia/fantasy/keep"
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

	// Convert the CSV to serialized JSON so that we can call the
	// library function.
	var rows [][]interface{}
	f, err := os.Open(path.Join(*dataDir, "merged.csv"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	reader := csv.NewReader(f)
	header, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}
	rows = append(rows, []interface{}{header[0], header[1], header[2], header[3]})
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		row := make([]interface{}, 4)
		row[0] = record[0]
		projection, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Fatal(err)
		}
		if *normalizePoints {
			switch {
			case strings.Contains(record[0], "- DEF"):
				row[1] = projection - 110.0
			case strings.Contains(record[0], "- K"):
				row[1] = projection - 135.0
			case strings.Contains(record[0], "- QB"):
				row[1] = projection - 255.0
			case strings.Contains(record[0], "- RB"):
				row[1] = projection - 110.0
			case strings.Contains(record[0], "- TE"):
				row[1] = projection - 70.0
			case strings.Contains(record[0], "- WR"):
				row[1] = projection - 85.0
			default:
				log.Fatal("Unknown pos in %v", record[0])
			}
		} else {
			row[1] = projection
		}
		row[2] = record[2]
		if record[3] == "" {
			row[3] = ""
		} else {
			row[3], err = strconv.ParseFloat(record[3], 64)
			if err != nil {
				log.Fatal(err)
			}
		}
		rows = append(rows, row)
	}

	// Call the library function.
	b, err := json.Marshal(rows)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	if err := keep.Run(bytes.NewReader(b), &buf); err != nil {
		log.Fatal(err)
	}
	var result [][][]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		log.Fatal(err)
	}

	// Output the results.  This currently loops through the entire
	// response for each gridder.  We could make it more efficient if
	// necessary.
	for g, row := range rows[1:] {
		if row[2] == "" {
			continue
		}
		fmt.Printf("%-20s %35s @ %2d ", row[2], row[0], int(row[3].(float64)))
		for _, keeps := range result {
			var round string
			var value float64
			for _, k := range keeps {
				if g+1 == int(k[0].(float64)) {
					round = k[1].(string)
					value = k[2].(float64)
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
