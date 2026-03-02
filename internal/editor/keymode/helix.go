package keymode

import (
	"github.com/gdamore/tcell/v2"
)

// HelixMode implements Helix-style selection-first editing.
// In Normal mode, all movement creates/extends selection.
// In Insert mode, behaves like standard editing.
type HelixMode struct {
	subMode  SubMode
	countBuf string
}

func NewHelixMode() *HelixMode {
	return &HelixMode{
		subMode: SubModeNormal,
	}
}

func (h *HelixMode) ProcessKey(ev *tcell.EventKey, ctx KeyContext) KeyResult {
	switch h.subMode {
	case SubModeInsert:
		return h.processInsert(ev, ctx)
	case SubModeNormal:
		return h.processNormal(ev, ctx)
	}
	return KeyResult{Handled: false}
}

func (h *HelixMode) processInsert(ev *tcell.EventKey, ctx KeyContext) KeyResult {
	if ev.Key() == tcell.KeyEscape {
		h.subMode = SubModeNormal
		return KeyResult{Handled: true}
	}

	// In insert mode, use standard key mappings
	action := MapKey(ev)
	if action != ActionNone {
		return KeyResult{
			Action:  int(action),
			Char:    ev.Rune(),
			Handled: true,
		}
	}
	return KeyResult{Handled: false}
}

func (h *HelixMode) processNormal(ev *tcell.EventKey, ctx KeyContext) KeyResult {
	key := ev.Key()
	mod := ev.Modifiers()
	ctrl := mod&tcell.ModCtrl != 0

	if key == tcell.KeyEscape {
		h.countBuf = ""
		// Clear selection
		return h.result(ActionCursorRight)
	}

	// Ctrl combos
	if ctrl {
		switch key {
		case tcell.KeyCtrlU:
			return h.result(ActionSelectPageUp)
		case tcell.KeyCtrlD:
			return h.result(ActionSelectPageDown)
		}
		return KeyResult{Handled: false}
	}

	if key != tcell.KeyRune {
		return KeyResult{Handled: false}
	}

	ch := ev.Rune()

	// Count prefix
	if (ch >= '1' && ch <= '9') || (ch == '0' && h.countBuf != "") {
		h.countBuf += string(ch)
		return KeyResult{Handled: true}
	}

	count := h.getCount()

	// In Helix Normal, movement keys create/extend selection
	switch ch {
	// Selection movement (Helix: all movement selects)
	case 'h':
		return h.repeatAction(ActionSelectLeft, count)
	case 'j':
		return h.repeatAction(ActionSelectDown, count)
	case 'k':
		return h.repeatAction(ActionSelectUp, count)
	case 'l':
		return h.repeatAction(ActionSelectRight, count)
	case 'w':
		return h.repeatAction(ActionSelectWordRight, count)
	case 'b':
		return h.repeatAction(ActionSelectWordLeft, count)
	case 'e':
		return h.repeatAction(ActionSelectWordRight, count)
	case '0':
		return h.result(ActionSelectHome)
	case '$':
		return h.result(ActionSelectEnd)

	// Line selection
	case 'x':
		return h.repeatAction(ActionSelectLine, count)
	case 'X':
		return h.repeatAction(ActionExtendLineSelect, count)

	// Actions on selection
	case 'd':
		return h.result(ActionCut)
	case 'c':
		h.subMode = SubModeInsert
		return h.result(ActionCut)
	case 'y':
		return h.result(ActionCopy)

	// Enter insert mode
	case 'i':
		h.subMode = SubModeInsert
		return KeyResult{Handled: true}
	case 'a':
		h.subMode = SubModeInsert
		// Move past selection end
		return h.result(ActionCursorRight)
	case 'I':
		h.subMode = SubModeInsert
		return h.result(ActionCursorHome)
	case 'A':
		h.subMode = SubModeInsert
		return h.result(ActionCursorEnd)
	case 'o':
		h.subMode = SubModeInsert
		return h.result(ActionOpenLineBelow)
	case 'O':
		h.subMode = SubModeInsert
		return h.result(ActionOpenLineAbove)

	// Other
	case 'u':
		return h.result(ActionUndo)
	case 'U':
		return h.result(ActionRedo)
	case 'p':
		return h.result(ActionPasteAfter)
	case 'P':
		return h.result(ActionPasteBefore)
	case 'g':
		// Simplified: g goes to doc start
		return h.result(ActionCursorDocStart)
	case 'G':
		return h.result(ActionCursorDocEnd)

	// Search
	case '/':
		return h.result(ActionSearchForward)
	case 'n':
		return h.result(ActionSearchNext)
	case 'N':
		return h.result(ActionSearchPrev)
	}

	return KeyResult{Handled: false}
}

func (h *HelixMode) result(action Action) KeyResult {
	h.countBuf = ""
	return KeyResult{
		Action:  int(action),
		Handled: true,
	}
}

func (h *HelixMode) repeatAction(action Action, count int) KeyResult {
	h.countBuf = ""
	if count <= 1 {
		return KeyResult{
			Action:  int(action),
			Handled: true,
		}
	}
	actions := make([]int, count)
	for i := range actions {
		actions[i] = int(action)
	}
	return KeyResult{
		Actions: actions,
		Handled: true,
	}
}

func (h *HelixMode) getCount() int {
	if h.countBuf == "" {
		return 1
	}
	n := 0
	for _, ch := range h.countBuf {
		n = n*10 + int(ch-'0')
	}
	h.countBuf = ""
	if n < 1 {
		return 1
	}
	return n
}

func (h *HelixMode) Mode() string { return "Helix" }

func (h *HelixMode) SubMode() SubMode { return h.subMode }

func (h *HelixMode) SubModeLabel() string {
	switch h.subMode {
	case SubModeInsert:
		return "INSERT"
	case SubModeNormal:
		return "NORMAL"
	}
	return "NORMAL"
}

func (h *HelixMode) PendingDisplay() string {
	return h.countBuf
}

func (h *HelixMode) CursorStyle() CursorStyle {
	if h.subMode == SubModeInsert {
		return CursorLine
	}
	return CursorBlock
}

func (h *HelixMode) Reset() {
	h.countBuf = ""
}
