package fantasy

// RunDraft simulates a draft starting from the given input state.  It
// modifies the state during the simulation.  The order and strategies
// are read-only.
func RunDraft(state *State, order []int, strategies []Strategy) {
	for ; state.Pick < len(order); state.Pick++ {
		i := order[state.Pick]
		player, justification := strategies[i].Select(state, order)
		state.Teams[i].Add(player, state.Pick, justification)
		state.Undrafted = removePlayer(state.Undrafted, player.ID)
	}
}
