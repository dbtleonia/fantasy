package fantasy

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Player struct {
	Pick   int // 0 if undrafted
	ID     int
	Name   string
	Pos    string // QB RB WR TE K DST
	Points float64
}

type ByPick []*Player

func (x ByPick) Len() int           { return len(x) }
func (x ByPick) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x ByPick) Less(i, j int) bool { return x[i].Pick < x[j].Pick }

func (p *Player) String() string {
	return fmt.Sprintf("%3d %07d %3s %8.4f %s", p.Pick, p.ID, p.Pos, p.Points, p.Name)
}

func ReadPlayers(filename string) ([]*Player, error) {
	const (
		colPick   = 0
		colID     = 1
		colName   = 2
		colPos    = 3
		colPoints = 4
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
		players = append(players, &Player{
			Pick:   pick,
			ID:     id,
			Name:   record[colName],
			Pos:    record[colPos],
			Points: points,
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
