package fantasy

type Scorer struct {
	Schema []byte
	Bench  bool
}

var (
	posMultiplier = map[byte]float64{
		'Q': 0.893,
		'R': 0.861,
		'W': 0.799,
		'T': 0.848,
		'K': 0.512,
		'D': 0.487,
	}

	benchWeights = map[byte][]float64{
		'D': {0.1},
		'K': {},
		'Q': {0.1},
		'R': {0.4, 0.2, 0.1},
		'T': {0.2},
		'W': {0.4, 0.2, 0.1},
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
	for _, player := range team.PlayersByPoints() {
		ch := player.Pos[0]
		if start[ch] > 0 {
			start[ch]--
			result += player.Points * posMultiplier[ch]
			continue
		}
		switch ch {
		case 'W', 'R', 'T':
			if start['X'] > 0 {
				start['X']--
				result += player.Points * posMultiplier[ch]
				continue
			}
		}
		if s.Bench {
			if bench[ch] < len(benchWeights[ch]) {
				result += player.Points * benchWeights[ch][bench[ch]] * posMultiplier[ch]
				bench[ch]++
				continue
			}
		}
	}
	return result
}
