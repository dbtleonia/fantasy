package fantasy

type Scorer struct {
	Schema []byte
	Bench  bool
}

var (
	// TODO: Make these configurable.
	benchWeights = map[byte]float64{
		'D': 1.0 / 16.0,
		'K': 0.0 / 16.0,
		'Q': 2.0 / 16.0,
		'R': 2.0 / 16.0,
		'T': 2.0 / 16.0,
		'W': 2.0 / 16.0,
		'X': 0.0 / 16.0,
	}
)

func (s *Scorer) Score(team *Team) float64 {
	start := make(map[byte]int)
	bench := make(map[byte]int)
	for _, ch := range s.Schema {
		if ch != 'B' {
			start[ch]++
			bench[ch]++
		}
	}
	result := 0.0
	for _, player := range team.PlayersByPoints() {
		ch := player.Pos[0]
		if start[ch] > 0 {
			start[ch]--
			result += player.Points
			continue
		}
		switch ch {
		case 'W', 'R', 'T':
			if start['X'] > 0 {
				start['X']--
				result += player.Points
				continue
			}
		}
		if s.Bench {
			if bench[ch] > 0 {
				bench[ch]--
				result += player.Points * benchWeights[ch]
				continue
			}
			switch ch {
			case 'W', 'R', 'T':
				if bench['X'] > 0 {
					bench['X']--
					result += player.Points * benchWeights['X']
					continue
				}
			}
		}
	}
	return result
}
