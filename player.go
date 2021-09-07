package fantasy

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
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
	return fmt.Sprintf("%3d %07d %3d %5.1f %3s %3d %6.2f %-3s %-25s # %s", p.Pick, p.ID, p.Rank, p.ADP, p.Pos, p.PosRank, p.Value, p.Team, p.Name, p.Justification)
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
