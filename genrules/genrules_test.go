package main

import (
	"testing"
)

func TestAllowedPos(t *testing.T) {
	schema := []byte("QRRWWWTXDKBBBBBBBB")

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
		autopick := allowedPos(schema, []byte("DKQRTWX"), []byte(tt.roster))
		if autopick != tt.autopick {
			t.Errorf("allowedPos(_, %s) = autopick %s; want %s", tt.roster, autopick, tt.autopick)
		}
		humanoid := allowedPos(schema, []byte("QRWX"), []byte(tt.roster))
		if humanoid != tt.humanoid {
			t.Errorf("allowedPos(_, %s) = humanoid %s; want %s", tt.roster, humanoid, tt.humanoid)
		}
	}
}
