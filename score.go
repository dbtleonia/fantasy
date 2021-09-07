package fantasy

type Scorer struct {
	Schema []byte
	Bench  bool
}

var (
	benchWeights = map[byte][]float64{
		'D': {0.2},
		'K': {},
		'Q': {},
		'R': {0.5, 0.2, 0.1},
		'T': {0.2},
		'W': {0.5, 0.2, 0.1},
	}
)

func (s *Scorer) Score(team *Team) float64 {
	start := make(map[byte]int)
	bench := make(map[byte]int)
	for _, ch := range s.Schema {
		if ch != 'B' {
			start[ch]++
		}
	}
	result := 0.0
	for _, player := range team.PlayersByValue() {
		ch := player.Pos[0]
		if start[ch] > 0 {
			start[ch]--
			result += player.Value
			continue
		}
		switch ch {
		case 'R':
			if start['X'] > 0 {
				start['X']--
				result += player.Value
				continue
			}
		}
		if s.Bench {
			if bench[ch] < len(benchWeights[ch]) {
				result += player.Value * benchWeights[ch][bench[ch]]
				bench[ch]++
				continue
			}
		}
	}
	return result
}
