package fantasy

import (
	"encoding/csv"
	"io"
	"os"
)

type Rules struct {
	AutopickRaw map[string]string
	HumanoidRaw map[string]string
	AutopickMap map[string]map[byte]bool
	HumanoidMap map[string]map[byte]bool
}

func ReadRules(rulesCsv string) (*Rules, error) {
	f, err := os.Open(rulesCsv)
	if err != nil {
		return nil, err
	}
	autopickRaw := make(map[string]string)
	humanoidRaw := make(map[string]string)
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
		autopickRaw[roster] = autopick
		humanoidRaw[roster] = humanoid
		autopickMap[roster] = make(map[byte]bool)
		for _, ch := range []byte(autopick) {
			autopickMap[roster][ch] = true
		}
		humanoidMap[roster] = make(map[byte]bool)
		for _, ch := range []byte(humanoid) {
			humanoidMap[roster][ch] = true
		}
	}
	return &Rules{autopickRaw, humanoidRaw, autopickMap, humanoidMap}, nil
}
