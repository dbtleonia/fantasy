package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

var (
	// Inputs:
	//   <data_dir>/managers/<manager-name>.csv
	//     field 0   = <gridder-name> (<team> - <pos>)
	//     field 1   = <round>
	//   <data_dir>/projections/<pos>.csv
	//     field 0   = <gridder-name>
	//     field 1   = <team>  # uppercase; empty for DEF.csv
	//     field n-1 = <projection>
	//   <data_dir>/missing.csv
	//     field 0   = <gridder-name> (<team> - <pos>)
	//   <data_dir>/renames.csv
	//     field 0   = <gridder-name-projections>
	//     field 1   = <gridder-name-managers>
	// Output:
	//   <data_dir>/merged.csv
	dataDir = flag.String("data_dir", "", "directory containing data sources")
)

var (
	teams = map[string]string{
		"ARI": "Ari",
		"ATL": "Atl",
		"BAL": "Bal",
		"BUF": "Buf",
		"CAR": "Car",
		"CHI": "Chi",
		"CIN": "Cin",
		"CLE": "Cle",
		"DAL": "Dal",
		"DEN": "Den",
		"DET": "Det",
		"GB":  "GB",
		"HOU": "Hou",
		"IND": "Ind",
		"JAC": "Jax",
		"KC":  "KC",
		"LAC": "LAC",
		"LAR": "LAR",
		"LV":  "LV",
		"MIA": "Mia",
		"MIN": "Min",
		"NE":  "NE",
		"NO":  "NO",
		"NYG": "NYG",
		"NYJ": "NYJ",
		"PHI": "Phi",
		"PIT": "Pit",
		"SEA": "Sea",
		"SF":  "SF",
		"TB":  "TB",
		"TEN": "Ten",
		"WAS": "Was",
	}
	defnames = map[string]string{
		"Arizona Cardinals":    "Arizona (Ari - DEF)",
		"Atlanta Falcons":      "Atlanta (Atl - DEF)",
		"Baltimore Ravens":     "Baltimore (Bal - DEF)",
		"Buffalo Bills":        "Buffalo (Buf - DEF)",
		"Carolina Panthers":    "Carolina (Car - DEF)",
		"Chicago Bears":        "Chicago (Chi - DEF)",
		"Cincinnati Bengals":   "Cincinnati (Cin - DEF)",
		"Cleveland Browns":     "Cleveland (Cle - DEF)",
		"Dallas Cowboys":       "Dallas (Dal - DEF)",
		"Denver Broncos":       "Denver (Den - DEF)",
		"Detroit Lions":        "Detroit (Det - DEF)",
		"Green Bay Packers":    "Green Bay (GB - DEF)",
		"Houston Texans":       "Houston (Hou - DEF)",
		"Indianapolis Colts":   "Indianapolis (Ind - DEF)",
		"Jacksonville Jaguars": "Jacksonville (Jax - DEF)",
		"Kansas City Chiefs":   "Kansas City (KC - DEF)",
		"Las Vegas Raiders":    "Las Vegas (LV - DEF)",
		"Los Angeles Chargers": "Los Angeles (LAC - DEF)",
		"Los Angeles Rams":     "Los Angeles (LAR - DEF)",
		"Miami Dolphins":       "Miami (Mia - DEF)",
		"Minnesota Vikings":    "Minnesota (Min - DEF)",
		"New England Patriots": "New England (NE - DEF)",
		"New Orleans Saints":   "New Orleans (NO - DEF)",
		"New York Giants":      "New York (NYG - DEF)",
		"New York Jets":        "New York (NYJ - DEF)",
		"Philadelphia Eagles":  "Philadelphia (Phi - DEF)",
		"Pittsburgh Steelers":  "Pittsburgh (Pit - DEF)",
		"San Francisco 49ers":  "San Francisco (SF - DEF)",
		"Seattle Seahawks":     "Seattle (Sea - DEF)",
		"Tampa Bay Buccaneers": "Tampa Bay (TB - DEF)",
		"Tennessee Titans":     "Tennessee (Ten - DEF)",
		"Washington Redskins":  "Washington (Was - DEF)",
	}
)

func main() {
	flag.Parse()

	// Open output file & write header.
	out, err := os.Create(path.Join(*dataDir, "merged.csv"))
	if err != nil {
		log.Fatal(err)
	}
	writer := csv.NewWriter(out)
	header := []string{"player", "proj", "manager", "rd"}
	if err := writer.Write(header); err != nil {
		log.Fatal(err)
	}

	// Read renames.
	renames := make(map[string]string)
	rfile, err := os.Open(path.Join(*dataDir, "renames.csv"))
	if err != nil {
		log.Fatal(err)
	}
	defer rfile.Close()
	rreader := csv.NewReader(rfile)
	for {
		record, err := rreader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		renames[record[0]] = record[1]
	}

	// Read missing.
	missing := make(map[string]bool)
	mfile, err := os.Open(path.Join(*dataDir, "missing.csv"))
	if err != nil {
		log.Fatal(err)
	}
	defer mfile.Close()
	mreader := csv.NewReader(mfile)
	for {
		record, err := mreader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		missing[record[0]] = true
	}

	// Read projections.
	projections := make(map[string]string) // gridder name -> projected points
	pfiles, err := ioutil.ReadDir(path.Join(*dataDir, "projections"))
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range pfiles {
		pos := strings.TrimSuffix(file.Name(), ".csv")
		f, err := os.Open(path.Join(*dataDir, "projections", file.Name()))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		r := csv.NewReader(f)
		r.Read() // skip header
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			var bigname string
			if record[1] == "" {
				var ok bool
				bigname, ok = defnames[record[0]]
				if !ok {
					log.Fatalf("Unknown name: %q in record %v", record[0], record)
				}
			} else {
				name := record[0]
				if rename, ok := renames[name]; ok {
					name = rename
				}
				team, ok := teams[record[1]]
				if !ok {
					log.Fatalf("Unknown team: %q in record %v", record[1], record)
				}
				bigname = fmt.Sprintf("%s (%s - %s)", name, team, pos)
			}
			projections[bigname] = record[len(record)-1]
		}
	}

	// Read managers & write owned gridders.
	mfiles, err := ioutil.ReadDir(path.Join(*dataDir, "managers"))
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range mfiles {
		managerName := strings.TrimSuffix(file.Name(), ".csv")
		f, err := os.Open(path.Join(*dataDir, "managers", file.Name()))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		r := csv.NewReader(f)
		r.Read() // skip header

		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			var (
				name  = record[0]
				round = record[1]
			)

			proj, ok := projections[name]
			if !ok {
				if !missing[name] {
					log.Fatalf("No projection for %q\n", name)
				}
				proj = "0.0"
			} else {
				delete(projections, name)
			}

			r := []string{name, proj, managerName, round}
			if err := writer.Write(r); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Write un-owned gridders.
	for name, proj := range projections {
		r := []string{name, proj, "", ""}
		if err := writer.Write(r); err != nil {
			log.Fatal(err)
		}
	}

	// Close output file.
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Fatal(err)
	}
}
