package fantasy

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

func ReadOrder(orderCsv string) ([]int, error) {
	f, err := os.Open(orderCsv)
	if err != nil {
		return nil, err
	}
	order := []int{8888} // dummy as first element; picks start at 1
	r := csv.NewReader(f)
	if _, err := r.Read(); err != nil { // discard header
		return nil, err
	}
	teams := make(map[string]int)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		pick, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, err
		}
		if pick != len(order) {
			return nil, fmt.Errorf("got pick %d, want %d", pick, len(order))
		}
		team, ok := teams[record[1]]
		if !ok {
			team = len(teams)
			teams[record[1]] = team
		}
		order = append(order, team)
	}
	return order, nil
}
