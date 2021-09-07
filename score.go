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
		'R': {0.5, 0.2},
		'T': {0.2},
		'W': {0.5, 0.2},
	}
)

func (s *Scorer) Score(team *Team) float64 {
	bye := make(map[byte]int)
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
			bye[ch] = player.Bye
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
				if (ch != 'D' && ch != 'T') || player.Bye != bye[ch] {
					result += player.Value*benchWeights[ch][bench[ch]] + 0.5
					bench[ch]++
					continue
				}
			}
		}
	}
	return result
}
