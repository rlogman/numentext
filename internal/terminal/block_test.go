package terminal

import (
	"testing"
)

func TestBlockTrackerOSC133(t *testing.T) {
	bt := NewBlockTracker()

	// Simulate OSC 133 sequence: A (prompt) -> B (command start) -> C (execution) -> D (done)
	bt.HandleOSC133('A') // Prompt start
	bt.HandleOSC133('B') // Command start
	bt.FeedChar('l')
	bt.FeedChar('s')
	bt.HandleOSC133('C') // Command executed

	if bt.BlockCount() != 1 {
		t.Fatalf("expected 1 block, got %d", bt.BlockCount())
	}
	if bt.Blocks[0].Command != "ls" {
		t.Errorf("expected command 'ls', got %q", bt.Blocks[0].Command)
	}
	if bt.Blocks[0].Finished {
		t.Error("block should not be finished yet")
	}

	// Add output
	bt.FeedChar('f')
	bt.FeedChar('o')
	bt.FeedChar('o')
	bt.FeedNewline()
	bt.FeedChar('b')
	bt.FeedChar('a')
	bt.FeedChar('r')
	bt.FeedNewline()

	bt.HandleOSC133('D') // Command finished

	if !bt.Blocks[0].Finished {
		t.Error("block should be finished")
	}
	if len(bt.Blocks[0].Output) != 2 {
		t.Errorf("expected 2 output lines, got %d", len(bt.Blocks[0].Output))
	}
	if bt.Blocks[0].Output[0] != "foo" {
		t.Errorf("expected output line 'foo', got %q", bt.Blocks[0].Output[0])
	}
	if bt.Blocks[0].Output[1] != "bar" {
		t.Errorf("expected output line 'bar', got %q", bt.Blocks[0].Output[1])
	}
}

func TestBlockTrackerMultipleBlocks(t *testing.T) {
	bt := NewBlockTracker()

	// First command
	bt.HandleOSC133('A')
	bt.HandleOSC133('B')
	bt.FeedChar('l')
	bt.FeedChar('s')
	bt.HandleOSC133('C')
	bt.FeedChar('a')
	bt.FeedNewline()
	bt.HandleOSC133('D')

	// Second command
	bt.HandleOSC133('A')
	bt.HandleOSC133('B')
	bt.FeedChar('p')
	bt.FeedChar('w')
	bt.FeedChar('d')
	bt.HandleOSC133('C')
	bt.FeedChar('/')
	bt.FeedChar('h')
	bt.FeedNewline()
	bt.HandleOSC133('D')

	if bt.BlockCount() != 2 {
		t.Fatalf("expected 2 blocks, got %d", bt.BlockCount())
	}
	if bt.Blocks[0].Command != "ls" {
		t.Errorf("first command: expected 'ls', got %q", bt.Blocks[0].Command)
	}
	if bt.Blocks[1].Command != "pwd" {
		t.Errorf("second command: expected 'pwd', got %q", bt.Blocks[1].Command)
	}
}

func TestBlockTrackerAltScreenSuspends(t *testing.T) {
	bt := NewBlockTracker()

	bt.SetAltScreen(true)
	bt.HandleOSC133('A')
	bt.HandleOSC133('B')
	bt.FeedChar('v')
	bt.FeedChar('i')
	bt.HandleOSC133('C')

	if bt.BlockCount() != 0 {
		t.Error("blocks should not be created during alt screen")
	}

	bt.SetAltScreen(false)
	bt.HandleOSC133('A')
	bt.HandleOSC133('B')
	bt.FeedChar('l')
	bt.FeedChar('s')
	bt.HandleOSC133('C')
	bt.HandleOSC133('D')

	if bt.BlockCount() != 1 {
		t.Fatalf("expected 1 block after alt screen off, got %d", bt.BlockCount())
	}
}

func TestBlockPlainText(t *testing.T) {
	blk := &CommandBlock{
		Command: "echo hello",
		Output:  []string{"hello"},
	}
	expected := "echo hello\nhello"
	if blk.PlainText() != expected {
		t.Errorf("expected %q, got %q", expected, blk.PlainText())
	}
}

func TestBlockOutputText(t *testing.T) {
	blk := &CommandBlock{
		Command: "ls",
		Output:  []string{"foo", "bar", "baz"},
	}
	expected := "foo\nbar\nbaz"
	if blk.OutputText() != expected {
		t.Errorf("expected %q, got %q", expected, blk.OutputText())
	}
}

func TestBlockCollapsed(t *testing.T) {
	blk := &CommandBlock{
		Command:   "ls -la",
		Output:    []string{"total 42", "drwxr-xr-x 2 user user 4096"},
		Collapsed: false,
	}

	if blk.Collapsed {
		t.Error("block should start expanded")
	}
	blk.Collapsed = true
	if !blk.Collapsed {
		t.Error("block should be collapsed after toggle")
	}
}

func TestVTParseOSC133(t *testing.T) {
	vt := NewVT(80, 24)

	// Write OSC 133;A (prompt start), terminated by BEL
	vt.Write([]byte("\x1b]133;A\x07"))
	// Write OSC 133;B (command start)
	vt.Write([]byte("\x1b]133;B\x07"))
	// Type "ls"
	vt.Write([]byte("ls"))
	// Write OSC 133;C (command executed)
	vt.Write([]byte("\x1b]133;C\x07"))

	bt := vt.Blocks()
	if bt.BlockCount() != 1 {
		t.Fatalf("expected 1 block, got %d", bt.BlockCount())
	}
	if bt.Blocks[0].Command != "ls" {
		t.Errorf("expected command 'ls', got %q", bt.Blocks[0].Command)
	}

	// Write output
	vt.Write([]byte("file1\n"))
	vt.Write([]byte("file2\n"))
	// Write OSC 133;D (command done)
	vt.Write([]byte("\x1b]133;D\x07"))

	if !bt.Blocks[0].Finished {
		t.Error("block should be finished")
	}
	if len(bt.Blocks[0].Output) != 2 {
		t.Errorf("expected 2 output lines, got %d", len(bt.Blocks[0].Output))
	}
}

func TestVTAltScreenBlockSuspend(t *testing.T) {
	vt := NewVT(80, 24)

	// Enter alt screen: CSI ?1049h
	vt.Write([]byte("\x1b[?1049h"))

	if !vt.Blocks().AltScreen() {
		t.Error("expected alt screen on")
	}

	// OSC 133 during alt screen should be ignored
	vt.Write([]byte("\x1b]133;A\x07"))
	vt.Write([]byte("\x1b]133;B\x07"))
	vt.Write([]byte("vim\x1b]133;C\x07"))

	if vt.Blocks().BlockCount() != 0 {
		t.Error("blocks should not be created during alt screen")
	}

	// Exit alt screen: CSI ?1049l
	vt.Write([]byte("\x1b[?1049l"))

	if vt.Blocks().AltScreen() {
		t.Error("expected alt screen off")
	}
}

func TestHeuristicCommand(t *testing.T) {
	bt := NewBlockTracker()

	bt.HeuristicCommand("ls -la")

	if bt.BlockCount() != 1 {
		t.Fatalf("expected 1 block, got %d", bt.BlockCount())
	}
	if bt.Blocks[0].Command != "ls -la" {
		t.Errorf("expected 'ls -la', got %q", bt.Blocks[0].Command)
	}

	// Feed output
	bt.FeedChar('t')
	bt.FeedChar('o')
	bt.FeedChar('t')
	bt.FeedNewline()

	if len(bt.Blocks[0].Output) != 0 {
		t.Errorf("heuristic block should not capture output without inOutput flag, got %d lines", len(bt.Blocks[0].Output))
	}

	// Empty command should not create a block
	bt.HeuristicCommand("")
	if bt.BlockCount() != 1 {
		t.Errorf("empty command should not create block, got %d", bt.BlockCount())
	}
}
