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

	ID     int
	Name   string
	Pos    string // QB RB WR TE K DST
	Team   string
	Points float64
	ADP    float64
	Stddev float64 // ADP stddev
}

type ByPick []*Player

func (x ByPick) Len() int           { return len(x) }
func (x ByPick) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ByPick) Less(i, j int) bool { return x[i].Pick < x[j].Pick }

type ByPoints []*Player

func (x ByPoints) Len() int           { return len(x) }
func (x ByPoints) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ByPoints) Less(i, j int) bool { return x[i].Points < x[j].Points }

func (p *Player) String() string {
	// TODO: Don't hardcode %-25s.
	return fmt.Sprintf("%3d %07d %5.1f %3s %7.2f %-3s %-30s # %s", p.Pick, p.ID, p.ADP, p.Pos, p.Points, p.Team, p.Name, p.Justification)
}

func ReadPlayers(filename string) ([]*Player, error) {
	const (
		colPick   = 0
		colID     = 1
		colName   = 2
		colPos    = 3
		colTeam   = 4
		colPoints = 5
		colADP    = 6
		colStddev = 7
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
		points, err := strconv.ParseFloat(record[colPoints], 64)
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
			Pick:   pick,
			ID:     id,
			Name:   record[colName],
			Team:   record[colTeam],
			Points: points,
			Pos:    record[colPos],
			ADP:    adp,
			Stddev: stddev,
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
