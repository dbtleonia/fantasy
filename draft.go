package fantasy

// RunDraft simulates a draft starting from the given input state.  It
// modifies the state during the simulation.  The order and strategies
// are read-only.
func RunDraft(state *State, order []int, strategies []Strategy) {
	for state.Pick < len(order) {
		i := order[state.Pick]
		if i == -1 { // this pick is a keeper, skip it
			state.Pick++
			continue
		}
		player, justification := strategies[i].Select(state)
		state.Update(i, player, justification)
		state.Pick++
	}
}
