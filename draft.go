package fantasy

// RunDraft simulates a draft starting from the given input state.  It
// modifies the state during the simulation.  The order and strategies
// are read-only.
func RunDraft(state *State, order []int, strategies []Strategy) {
	for ; state.Pick < len(order); state.Pick++ {
		i := order[state.Pick]
		j := strategies[i].Select(state, order)
		player := state.Undrafted[j]
		state.Teams[i].Add(player, state.Pick)
		state.Undrafted = append(state.Undrafted[:j], state.Undrafted[j+1:]...)
	}
}
