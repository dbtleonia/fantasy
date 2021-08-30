package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	// ===== Inputs under <data_dir> =====
	//
	// keeper_rounds.csv
	//   field 0  = <manager>
	//   field 1  = <player-name-canon>
	//   field 2  = <round>
	// raw3_projections.tsv  -- TSV, not CSV
	//   field 1  = <player-name-format3>
	//   field 2  = $<value>
	// raw3_extra_renames.csv
	//   field 0  = <player-name-format3>
	//   field 1  = <player-name-canon>
	// raw3_extra_projections.csv
	//   field 0  = <player-name-canon>
	//
	// ===== Output under <data_dir> =====
	//
	// projections.csv
	//   field 0  = <player-name-canon>
	//   field 1  = <value>
	//   field 2  = <stddev>  # empty, future formats may populate
	//
	dataDir = flag.String("data_dir", "", "directory for data files; empty string means $HOME/data")

	// Mapping of team formats from raw -> canon.
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
		"":    "",
	}

	teamRE = regexp.MustCompile(`\([^ ]*`)
)

func translateName(playerRaw string) (string, error) {
	// Remove extra chars to the right of ')'.
	playerRaw = strings.TrimRightFunc(playerRaw, unicode.IsLetter)

	// Replace '(<team-raw>' with '(<team-canon'.
	var err error
	result := teamRE.ReplaceAllStringFunc(playerRaw, func(teamRaw string) string {
		teamRaw = teamRaw[1:] // strip '('
		teamCanon, ok := teams[teamRaw]
		if !ok {
			err = fmt.Errorf("unknown team: %q in player %q", teamRaw, playerRaw)
			return ""
		}
		return "(" + teamCanon
	})

	return result, err
}

func mustReadAll(filename string) [][]string {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	if filepath.Ext(filename) == ".tsv" {
		r.Comma = '\t'
	}
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	return records
}

func main() {
	flag.Parse()

	dir := *dataDir
	if dir == "" {
		if home := os.Getenv("HOME"); home != "" {
			dir = path.Join(home, "data")
		}
	}

	// Read extra renames.
	extraRenames := make(map[string]string)
	for _, record := range mustReadAll(path.Join(dir, "raw3_extra_renames.csv")) {
		extraRenames[record[0]] = record[1]
	}

	// Read extra projections.
	extraProjections := make(map[string]string) // player -> value
	extraProjectionsOrder := []string{}
	for _, record := range mustReadAll(path.Join(dir, "raw3_extra_projections.csv")) {
		extraProjections[record[0]] = record[1]
		extraProjectionsOrder = append(extraProjectionsOrder, record[0])
	}

	// Read raw projections.
	projections := make(map[string]string) // player -> value
	projectionsOrder := []string{}
	for _, record := range mustReadAll(path.Join(dir, "raw3_projections.tsv")) {
		playerRaw := record[1]

		var player string
		if rename, ok := extraRenames[playerRaw]; ok {
			player = rename
		} else {
			var err error
			player, err = translateName(playerRaw)
			if err != nil {
				log.Fatal(err)
			}
		}

		value := strings.TrimPrefix(record[2], "$")

		projections[player] = value
		projectionsOrder = append(projectionsOrder, player)
	}

	// Read keeper_rounds and check constraints.
	krecords := mustReadAll(path.Join(dir, "keeper_rounds.csv"))
	for _, record := range krecords[1:] { // skip header
		name := record[1]
		_, ok1 := projections[name]
		_, ok2 := extraProjections[name]
		if !ok1 && !ok2 {
			log.Fatalf("No projection for %q; add to extra renames or projections\n", name)
		}
		if ok1 && ok2 {
			log.Fatalf("Multiple projections for %q; remove from extra projections\n", name)
		}
	}

	// Write output file.
	out := [][]string{{"player", "value", "stddev"}}
	for _, name := range projectionsOrder {
		out = append(out, []string{name, projections[name], ""})
	}
	for _, name := range extraProjectionsOrder {
		out = append(out, []string{name, extraProjections[name], ""})
	}
	f, err := os.Create(path.Join(dir, "projections.csv"))
	if err != nil {
		log.Fatal(err)
	}
	if err := csv.NewWriter(f).WriteAll(out); err != nil {
		log.Fatal(err)
	}
}
