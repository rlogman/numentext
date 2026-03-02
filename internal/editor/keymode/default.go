package keymode

import (
	"github.com/gdamore/tcell/v2"
)

// DefaultMode implements the standard non-modal keymap with Insert/Overwrite toggle
type DefaultMode struct {
	overwrite bool
}

func NewDefaultMode() *DefaultMode {
	return &DefaultMode{}
}

func (d *DefaultMode) ProcessKey(ev *tcell.EventKey, ctx KeyContext) KeyResult {
	// Insert key toggles overwrite mode
	if ev.Key() == tcell.KeyInsert {
		d.overwrite = !d.overwrite
		return KeyResult{Handled: true}
	}

	action := MapKey(ev)
	if action == ActionNone {
		return KeyResult{Handled: false}
	}

	// In overwrite mode, replace ActionInsertChar with ActionOverwriteChar
	if d.overwrite && action == ActionInsertChar {
		return KeyResult{
			Action:  int(ActionOverwriteChar),
			Char:    ev.Rune(),
			Handled: true,
		}
	}

	return KeyResult{
		Action:  int(action),
		Char:    ev.Rune(),
		Handled: true,
	}
}

func (d *DefaultMode) Mode() string { return "Default" }

func (d *DefaultMode) SubMode() SubMode {
	if d.overwrite {
		return SubModeOverwrite
	}
	return SubModeInsert
}

func (d *DefaultMode) SubModeLabel() string {
	if d.overwrite {
		return "OVR"
	}
	return "INS"
}

func (d *DefaultMode) PendingDisplay() string { return "" }

func (d *DefaultMode) CursorStyle() CursorStyle {
	if d.overwrite {
		return CursorBlock
	}
	return CursorLine
}

func (d *DefaultMode) Reset() {}
