package keymode

// viOperatorMotion composes an operator (d, c, y) with a motion to produce actions.
// Returns a list of actions to execute, or nil if the motion is invalid.
func viOperatorMotion(op rune, motion rune, count int, ctx KeyContext) []int {
	if count < 1 {
		count = 1
	}

	switch op {
	case 'd':
		return viDeleteMotion(motion, count)
	case 'c':
		return viChangeMotion(motion, count)
	case 'y':
		return viYankMotion(motion, count)
	}
	return nil
}

func viDeleteMotion(motion rune, count int) []int {
	sel := selectMotion(motion, count)
	if sel == nil {
		return nil
	}
	return append(sel, int(ActionCut))
}

func viChangeMotion(motion rune, count int) []int {
	sel := selectMotion(motion, count)
	if sel == nil {
		return nil
	}
	// Cut then the vi mode will switch to insert
	return append(sel, int(ActionCut))
}

func viYankMotion(motion rune, count int) []int {
	sel := selectMotion(motion, count)
	if sel == nil {
		return nil
	}
	return append(sel, int(ActionCopy))
}

// selectMotion returns a sequence of select actions corresponding to a vi motion
func selectMotion(motion rune, count int) []int {
	var actions []int
	for i := 0; i < count; i++ {
		switch motion {
		case 'w':
			actions = append(actions, int(ActionSelectWordRight))
		case 'b':
			actions = append(actions, int(ActionSelectWordLeft))
		case 'e':
			actions = append(actions, int(ActionSelectWordRight))
		case '$':
			actions = append(actions, int(ActionSelectEnd))
		case '0':
			actions = append(actions, int(ActionSelectHome))
		case '^':
			actions = append(actions, int(ActionSelectHome))
		case 'j':
			actions = append(actions, int(ActionSelectDown))
		case 'k':
			actions = append(actions, int(ActionSelectUp))
		case 'h':
			actions = append(actions, int(ActionSelectLeft))
		case 'l':
			actions = append(actions, int(ActionSelectRight))
		default:
			return nil
		}
	}
	return actions
}
