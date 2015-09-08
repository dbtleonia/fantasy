package fantasy

import (
	"encoding/csv"
	"io"
	"os"
)

type Rules struct {
	AutopickMap map[string]map[byte]bool
	HumanoidMap map[string]map[byte]bool
}

func ReadRules(rulesCsv string) (*Rules, error) {
	f, err := os.Open(rulesCsv)
	if err != nil {
		return nil, err
	}
	autopickMap := make(map[string]map[byte]bool)
	humanoidMap := make(map[string]map[byte]bool)
	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		var (
			roster   = record[0]
			autopick = record[1]
			humanoid = record[2]
		)
		autopickMap[roster] = make(map[byte]bool)
		for _, ch := range []byte(autopick) {
			autopickMap[roster][ch] = true
		}
		humanoidMap[roster] = make(map[byte]bool)
		for _, ch := range []byte(humanoid) {
			humanoidMap[roster][ch] = true
		}
	}
	return &Rules{autopickMap, humanoidMap}, nil
}
