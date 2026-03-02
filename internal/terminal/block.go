package terminal

import "strings"

// CommandBlock represents a single command+output region in the terminal.
// Blocks are detected via OSC 133 shell integration sequences or prompt heuristics.
type CommandBlock struct {
	Command   string   // the command text (what was typed)
	Output    []string // output lines (plain text, stripped of ANSI)
	Collapsed bool     // whether this block is collapsed in the UI
	Finished  bool     // true once the command has completed (OSC 133;D received)
}

// BlockTracker accumulates CommandBlocks from OSC 133 events and heuristics.
type BlockTracker struct {
	Blocks        []*CommandBlock
	active        *CommandBlock // block being built (between prompt and completion)
	inPrompt      bool          // between OSC 133;A (prompt start) and OSC 133;B (command start)
	inCommand     bool          // between OSC 133;B and first newline (capturing command text)
	inOutput      bool          // between command start and OSC 133;D (command done)
	altScreen     bool          // alternate screen active — suspend block tracking
	hasSeenOSC133 bool          // true once any OSC 133 sequence has been seen
	promptLine    string        // accumulates prompt text for heuristic fallback
	commandLine   string        // accumulates command text
	curLine       string        // current line being written
}

// NewBlockTracker creates a new block tracker.
func NewBlockTracker() *BlockTracker {
	return &BlockTracker{}
}

// SetAltScreen enables or disables alternate screen mode.
// When alternate screen is active, block tracking is suspended.
func (bt *BlockTracker) SetAltScreen(on bool) {
	bt.altScreen = on
}

// AltScreen returns whether alternate screen is active.
func (bt *BlockTracker) AltScreen() bool {
	return bt.altScreen
}

// HandleOSC133 processes an OSC 133 shell integration sequence.
// The param is the single letter after "133;", e.g. 'A', 'B', 'C', 'D'.
func (bt *BlockTracker) HandleOSC133(param byte) {
	if bt.altScreen {
		return
	}
	bt.hasSeenOSC133 = true
	switch param {
	case 'A': // Prompt start — a new prompt is being drawn
		bt.inPrompt = true
		bt.inCommand = false
		bt.inOutput = false
		bt.promptLine = ""
		bt.commandLine = ""
	case 'B': // Command start — user has pressed Enter, command text follows
		bt.inPrompt = false
		bt.inCommand = true
		bt.inOutput = false
	case 'C': // Command executed — output begins
		bt.inCommand = false
		bt.inOutput = true
		cmd := bt.commandLine
		if cmd == "" {
			cmd = bt.curLine
		}
		bt.active = &CommandBlock{
			Command: cmd,
		}
		bt.Blocks = append(bt.Blocks, bt.active)
		bt.curLine = ""
	case 'D': // Command finished — mark block done
		if bt.active != nil {
			bt.active.Finished = true
		}
		bt.active = nil
		bt.inPrompt = false
		bt.inCommand = false
		bt.inOutput = false
		bt.curLine = ""
	}
}

// FeedChar is called for each printable character written to the terminal.
// It helps track command text and output lines for block assembly.
func (bt *BlockTracker) FeedChar(ch rune) {
	if bt.altScreen {
		return
	}
	if bt.inCommand {
		bt.commandLine += string(ch)
	}
	bt.curLine += string(ch)
}

// FeedNewline is called when a newline (LF) occurs.
func (bt *BlockTracker) FeedNewline() {
	if bt.altScreen {
		return
	}
	if bt.inOutput && bt.active != nil {
		bt.active.Output = append(bt.active.Output, bt.curLine)
	}
	bt.curLine = ""
}

// FeedCR is called on carriage return.
func (bt *BlockTracker) FeedCR() {
	// CR just resets the line position; we'll overwrite curLine content
	bt.curLine = ""
}

// HeuristicPrompt attempts prompt detection when OSC 133 is not available.
// Call this when a line ending in a common prompt char ($ % > #) is seen
// after a completed block or at the start.
func (bt *BlockTracker) HeuristicPrompt(line string) {
	if bt.altScreen {
		return
	}
	// If there's an active unfinished block, finish it
	if bt.active != nil && !bt.active.Finished {
		bt.active.Finished = true
		bt.active = nil
	}
}

// HeuristicCommand records a command submitted via Enter key press.
// The command text is the current line content.
func (bt *BlockTracker) HeuristicCommand(cmd string) {
	if bt.altScreen || cmd == "" {
		return
	}
	bt.active = &CommandBlock{
		Command: cmd,
	}
	bt.Blocks = append(bt.Blocks, bt.active)
}

// HeuristicEnter is called when the user presses Enter.
// If OSC 133 is not active and there's no current active block,
// it uses the accumulated current line as a command.
func (bt *BlockTracker) HeuristicEnter() {
	if bt.altScreen {
		return
	}
	// Only use heuristic if no OSC 133 sequence is in progress
	if bt.inPrompt || bt.inCommand || bt.inOutput {
		return
	}
	cmd := strings.TrimSpace(bt.curLine)
	if cmd == "" {
		return
	}
	// If there's an active unfinished block, finish it first
	if bt.active != nil && !bt.active.Finished {
		bt.active.Finished = true
	}
	bt.active = &CommandBlock{
		Command: cmd,
	}
	bt.Blocks = append(bt.Blocks, bt.active)
	bt.curLine = ""
}

// HasOSC133 returns true if at least one OSC 133 sequence has been processed.
func (bt *BlockTracker) HasOSC133() bool {
	return bt.inPrompt || bt.inCommand || bt.inOutput || bt.hasSeenOSC133
}

// ActiveBlock returns the currently building block (may be nil).
func (bt *BlockTracker) ActiveBlock() *CommandBlock {
	return bt.active
}

// SelectedBlock returns the block at the given index, or nil.
func (bt *BlockTracker) SelectedBlock(idx int) *CommandBlock {
	if idx < 0 || idx >= len(bt.Blocks) {
		return nil
	}
	return bt.Blocks[idx]
}

// BlockCount returns the number of blocks tracked.
func (bt *BlockTracker) BlockCount() int {
	return len(bt.Blocks)
}

// PlainText returns the full text of a block: command + output joined by newlines.
func (b *CommandBlock) PlainText() string {
	result := b.Command
	for _, line := range b.Output {
		result += "\n" + line
	}
	return result
}

// OutputText returns just the output lines joined by newlines.
func (b *CommandBlock) OutputText() string {
	result := ""
	for i, line := range b.Output {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}
