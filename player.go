package fantasy

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type Player struct {
	Pick          int    // 0 if undrafted
	Justification string // "" if undrafted

	ID      int
	Name    string
	Team    string
	Value   float64
	Pos     string // QB RB WR TE K DST
	Rank    int
	PosRank int
	ADP     float64
	Stddev  float64 // ADP stddev

	Bye int
}

var byeMap = map[string]int{
	"":    0,
	"ARI": 12,
	"ATL": 6,
	"BAL": 8,
	"BUF": 7,
	"CAR": 13,
	"CHI": 10,
	"CIN": 10,
	"CLE": 13,
	"DAL": 7,
	"DEN": 11,
	"DET": 9,
	"GB":  13,
	"HOU": 10,
	"IND": 14,
	"JAC": 7,
	"KC":  12,
	"LAC": 7,
	"LAR": 11,
	"LV":  8,
	"MIA": 14,
	"MIN": 7,
	"NE":  14,
	"NO":  6,
	"NYG": 10,
	"NYJ": 6,
	"PHI": 14,
	"PIT": 7,
	"SEA": 9,
	"SF":  6,
	"TB":  9,
	"TEN": 13,
	"WAS": 9,
}

func lookupBye(name string) int {
	if true {
		return 0 // disable byes for now
	}
	if strings.Contains(name, "dummy") {
		return 0
	}
	team := strings.Split(strings.Split(name, "(")[1], " ")[0]
	bye, ok := byeMap[team]
	if !ok {
		log.Fatal("Unknown team: %q", team)
	}
	return bye
}

type ByPick []*Player

func (x ByPick) Len() int           { return len(x) }
func (x ByPick) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ByPick) Less(i, j int) bool { return x[i].Pick < x[j].Pick }

type ByValue []*Player

func (x ByValue) Len() int           { return len(x) }
func (x ByValue) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ByValue) Less(i, j int) bool { return x[i].Value < x[j].Value }

type ByADP []*Player

func (x ByADP) Len() int           { return len(x) }
func (x ByADP) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ByADP) Less(i, j int) bool { return x[i].ADP < x[j].ADP }

func (p *Player) String() string {
	// TODO: Don't hardcode %-25s.
	return fmt.Sprintf("%3d %07d %3d %5.1f %3s %3d %7.2f %-3s b%02d %-30s # %s", p.Pick, p.ID, p.Rank, p.ADP, p.Pos, p.PosRank, p.Value, p.Team, p.Bye, p.Name, p.Justification)
}

func ReadPlayers(filename string) ([]*Player, error) {
	const (
		colPick    = 0
		colID      = 1
		colName    = 2
		colPos     = 3
		colTeam    = 4
		colValue   = 5
		colRank    = 7
		colPosRank = 8
		colADP     = 9
		colStddev  = 12
	)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	var players []*Player
	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		pick, err := strconv.Atoi(record[colPick])
		if err != nil {
			return nil, err
		}
		id, err := strconv.Atoi(record[colID])
		if err != nil {
			return nil, err
		}
		value, err := strconv.ParseFloat(record[colValue], 64)
		if err != nil {
			return nil, err
		}
		rank, err := strconv.Atoi(record[colRank])
		if err != nil {
			return nil, err
		}
		posRank, err := strconv.Atoi(record[colPosRank])
		if err != nil {
			return nil, err
		}
		adp, err := strconv.ParseFloat(record[colADP], 64)
		if err != nil {
			return nil, err
		}
		stddev, err := strconv.ParseFloat(record[colStddev], 64)
		if err != nil {
			return nil, err
		}
		players = append(players, &Player{
			Pick:    pick,
			ID:      id,
			Name:    record[colName],
			Team:    record[colTeam],
			Value:   value,
			Pos:     record[colPos],
			Rank:    rank,
			PosRank: posRank,
			ADP:     adp,
			Bye:     lookupBye(record[colName]),
			Stddev:  stddev,
		})
	}
	return players, nil
}

func clonePlayers(players []*Player) []*Player {
	result := make([]*Player, len(players))
	for i, input := range players {
		output := *input
		result[i] = &output
	}
	return result
}

func removePlayer(players []*Player, id int) []*Player {
	for j, p := range players {
		if p.ID == id {
			return append(players[:j], players[j+1:]...)
		}
	}
	return players
}
