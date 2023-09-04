package fantasy

// RunDraft simulates a draft starting from the given input state.  It
// modifies the state during the simulation.  The order and strategies
// are read-only.
func RunDraft(state *State, order []int, strategies []Strategy) {
	for ; state.Pick < len(order); state.Pick++ {
		i := order[state.Pick]
		if i == -1 { // this pick is a keeper, skip it
			continue
		}
		player, justification := strategies[state.Pick].Select(state)
		state.Teams[i].Add(player, state.Pick, justification)
		state.UndraftedByPoints = removePlayer(state.UndraftedByPoints, player.ID)
		state.UndraftedByADP = removePlayer(state.UndraftedByADP, player.ID)
	}
}
