package fantasy

type Scorer struct {
	Schema []byte
}

func (s *Scorer) Score(team *Team) float64 {
	open := make(map[byte]int)
	for _, ch := range s.Schema {
		if ch != 'B' {
			open[ch]++
		}
	}
	result := 0.0
	for _, player := range team.PlayersByPoints() {
		ch := player.Pos[0]
		if open[ch] > 0 {
			open[ch]--
			result += player.Points
			continue
		}
		switch ch {
		case 'W', 'R', 'T':
			if open['X'] > 0 {
				open['X']--
				result += player.Points
			}
		}
	}
	return result
}
