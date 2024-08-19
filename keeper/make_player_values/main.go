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
	// out/keeper-options.csv
	//   field 1  = <player-canon>
	// fantasypros/raw-player-values.tsv   -- TSV, not CSV
	//   field 1  = <player-raw>
	//   field 2  = $<value>
	// fantasypros/extra-renames.csv
	//   field 0  = <player-raw>
	//   field 1  = <player-canon>
	// fantasypros/extra-player-values.csv
	//   field 0  = <player-canon>
	//   field 1  = <value>
	//
	// ===== Output under <data_dir> =====
	//
	// out/player-values.csv
	//   field 0  = <player-canon>
	//   field 1  = <value>
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
	// Did include team in player name.
	if false {
		n := strings.Index(playerRaw, "(")
		return playerRaw[:n-1], nil
	}

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
	for _, record := range mustReadAll(path.Join(dir, "fantasypros", "extra-renames.csv")) {
		extraRenames[record[0]] = record[1]
	}

	// Read extra player values.
	extraProjections := make(map[string]string) // player -> value
	extraProjectionsOrder := []string{}
	for _, record := range mustReadAll(path.Join(dir, "fantasypros", "extra-player-values.csv")) {
		extraProjections[record[0]] = record[1]
		extraProjectionsOrder = append(extraProjectionsOrder, record[0])
	}

	// Read raw player values.
	projections := make(map[string]string) // player -> value
	projectionsOrder := []string{}
	for _, record := range mustReadAll(path.Join(dir, "fantasypros", "raw-player-values.tsv")) {
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

	// Read keeper_options and check constraints.
	var problems []string
	krecords := mustReadAll(path.Join(dir, "out", "keeper-options.csv"))
	for _, record := range krecords[1:] { // skip header
		name := record[1]
		_, ok1 := projections[name]
		_, ok2 := extraProjections[name]
		if !ok1 && !ok2 {
			problems = append(problems, fmt.Sprintf("no value for %q; add to extra-renames.csv or extra-player-values.csv", name))
		}
		if ok1 && ok2 {
			problems = append(problems, fmt.Sprintf("multiple values for %q; remove from extra-player-values.csv", name))
		}
	}
	if len(problems) > 0 {
		log.Fatalf("Checks failed:\n  %s\n", strings.Join(problems, "\n  "))
	}

	// Write output file.
	out := [][]string{{"player", "value"}}
	for _, name := range projectionsOrder {
		out = append(out, []string{name, projections[name]})
	}
	for _, name := range extraProjectionsOrder {
		out = append(out, []string{name, extraProjections[name]})
	}
	f, err := os.Create(path.Join(dir, "out", "player-values.csv"))
	if err != nil {
		log.Fatal(err)
	}
	if err := csv.NewWriter(f).WriteAll(out); err != nil {
		log.Fatal(err)
	}
}
