package main

import (
	"testing"
)

var (
	schema           = []byte("QRRWWWTXDKBBBBBBBB")
	autopickPriority = []byte("DKQRTWX")
	autopickMin      = []byte("DKQQTTRRRRWWWW")
	autopickMax      = []byte("DDKKQQTTTRRRRRRWWWWWW")
	humanoidPriority = []byte("QRWX")
	humanoidMin      = []byte("DKQQTTRRRRWWWW")
	humanoidMax      = []byte("DDKQQTTTRRRRRRWWWWWW")

	noMin = []byte("")
	noMax = []byte("" +
		"DDDDDDDDDDDDDDDDDD" +
		"KKKKKKKKKKKKKKKKKK" +
		"QQQQQQQQQQQQQQQQQQ" +
		"RRRRRRRRRRRRRRRRRR" +
		"TTTTTTTTTTTTTTTTTT" +
		"WWWWWWWWWWWWWWWWWW")
)

func TestAllowedPosNoMinMax(t *testing.T) {
	tests := []struct {
		roster   string
		autopick string
		humanoid string
	}{
		{"", "DKQRTW", "DKQRTW"},
		{"RR", "DKQRTW", "DKQRTW"},          // still an X roster free
		{"RRR", "DKQTW", "DKQTW"},           // no more R
		{"QRW", "DKRTW", "DKRTW"},           // no more Q
		{"RRRWWWT", "DKQ", "DKQ"},           // must fill Q
		{"DRRRWWWT", "KQ", "KQ"},            // must fill Q
		{"QRRRTWWW", "DK", "DKQRTW"},        // autopick requires DK before others
		{"DKQRRRWWW", "T", "DKQRTW"},        // autopick requires T before others
		{"QRRRTWWWWWWWWWW", "DK", "DKQRTW"}, // almost end
		{"QRRRTWWWWWWWWWWW", "DK", "DK"},    // at end must pick DK
		{"KQRRRTWWWWWWWWWWW", "D", "D"},     // at end must pick D
		{"DQRRRTWWWWWWWWWWW", "K", "K"},     // at end must pick K
	}
	for _, tt := range tests {
		autopick := allowedPos(schema, autopickPriority, noMin, noMax, []byte(tt.roster))
		if autopick != tt.autopick {
			t.Errorf("allowedPos(_, %s) = autopick %s; want %s", tt.roster, autopick, tt.autopick)
		}
		humanoid := allowedPos(schema, humanoidPriority, noMin, noMax, []byte(tt.roster))
		if humanoid != tt.humanoid {
			t.Errorf("allowedPos(_, %s) = humanoid %s; want %s", tt.roster, humanoid, tt.humanoid)
		}
	}
}

func TestAllowedPosWithMinMax(t *testing.T) {
	tests := []struct {
		roster   string
		autopick string
		humanoid string
	}{
		{"", "DKQRTW", "DKQRTW"},
		{"RR", "DKQRTW", "DKQRTW"},         // still an X roster free
		{"RRR", "DKQTW", "DKQTW"},          // no more R
		{"QRW", "DKRTW", "DKRTW"},          // no more Q
		{"RRRWWWT", "DKQ", "DKQ"},          // must fill Q
		{"DRRRWWWT", "KQ", "KQ"},           // must fill Q
		{"QRRRTWWW", "DK", "DKQRTW"},       // autopick requires DK before others
		{"DKQRRRWWW", "T", "DQRTW"},        // autopick requires T before others; humanoid one K
		{"QRRRTWWWWWWWWWW", "DK", "DKQRT"}, // almost end; reached min W
		{"QRRRTWWWWWWWWWWW", "DK", "DK"},   // at end must pick DK
		{"KQRRRTWWWWWWWWWWW", "D", "D"},    // at end must pick D
		{"DQRRRTWWWWWWWWWWW", "K", "K"},    // at end must pick K
		{"DKQRRRRTTWWWWWWWW", "Q", "Q"},    // at end need min QQ
		{"DDKQQRRRRTTWWWWWW", "KRT", "RT"}, // maxed out DQW and humanoid K
	}
	for _, tt := range tests {
		autopick := allowedPos(schema, autopickPriority, autopickMin, autopickMax, []byte(tt.roster))
		if autopick != tt.autopick {
			t.Errorf("allowedPos(_, %s) = autopick %s; want %s", tt.roster, autopick, tt.autopick)
		}
		humanoid := allowedPos(schema, humanoidPriority, humanoidMin, humanoidMax, []byte(tt.roster))
		if humanoid != tt.humanoid {
			t.Errorf("allowedPos(_, %s) = humanoid %s; want %s", tt.roster, humanoid, tt.humanoid)
		}
	}
}
