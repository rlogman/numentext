package keymode

import (
	"strconv"
	"strings"
)

// viCommandParse parses a Vi : command string and invokes the appropriate callback
func viCommandParse(input string, callbacks *ViCommandCallback) {
	if callbacks == nil {
		return
	}
	cmd := strings.TrimSpace(input)
	if cmd == "" {
		return
	}

	switch cmd {
	case "w":
		if callbacks.OnSave != nil {
			callbacks.OnSave()
		}
	case "q":
		if callbacks.OnQuit != nil {
			callbacks.OnQuit()
		}
	case "wq", "x":
		if callbacks.OnSaveQuit != nil {
			callbacks.OnSaveQuit()
		}
	case "q!":
		// Force quit — same as quit for now
		if callbacks.OnQuit != nil {
			callbacks.OnQuit()
		}
	default:
		// Check for :{number} — go to line
		if n, err := strconv.Atoi(cmd); err == nil && n > 0 {
			if callbacks.OnGoToLine != nil {
				callbacks.OnGoToLine(n)
			}
		}
	}
}
