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
	VOR     float64
	Pos     string // QB RB WR TE K DST
	Points  float64
	Rank    int
	PosRank int
	ADP     float64
	Ceiling float64
}

type ByPick []*Player

func (x ByPick) Len() int           { return len(x) }
func (x ByPick) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ByPick) Less(i, j int) bool { return x[i].Pick < x[j].Pick }

type ByVOR []*Player

func (x ByVOR) Len() int           { return len(x) }
func (x ByVOR) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ByVOR) Less(i, j int) bool { return x[i].VOR < x[j].VOR }

type ByADP []*Player

func (x ByADP) Len() int           { return len(x) }
func (x ByADP) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ByADP) Less(i, j int) bool { return x[i].ADP < x[j].ADP }

func (p *Player) String() string {
	// TODO: Don't hardcode %-25s.
	return fmt.Sprintf("%3d %07d %3d %5.1f %3s %3d %8.4f %8.4f %9.4f %-3s %-25s # %s", p.Pick, p.ID, p.Rank, p.ADP, p.Pos, p.PosRank, p.Points, p.Ceiling, p.VOR, p.Team, p.Name, p.Justification)
}

func ReadPlayers(filename string) ([]*Player, error) {
	const (
		colPick    = 0
		colID      = 1
		colName    = 2
		colPos     = 3
		colTeam    = 4
		colVOR     = 5
		colPoints  = 6
		colRank    = 7
		colPosRank = 8
		colADP     = 9
		colCeiling = 10
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
		vor, err := strconv.ParseFloat(record[colVOR], 64)
		if err != nil {
			return nil, err
		}
		points, err := strconv.ParseFloat(record[colPoints], 64)
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
		ceiling, err := strconv.ParseFloat(record[colCeiling], 64)
		if err != nil {
			return nil, err
		}
		players = append(players, &Player{
			Pick:    pick,
			ID:      id,
			Name:    record[colName],
			Team:    record[colTeam],
			VOR:     vor,
			Pos:     record[colPos],
			Points:  points,
			Rank:    rank,
			PosRank: posRank,
			ADP:     adp,
			Ceiling: ceiling,
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
