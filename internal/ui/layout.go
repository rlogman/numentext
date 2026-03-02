package ui

import (
	"github.com/rivo/tview"
)

// Layout holds the main application layout
type Layout struct {
	Root     *tview.Flex
	MainGrid *tview.Flex

	MenuBar   *MenuBar
	FileTree  tview.Primitive
	Editor    tview.Primitive
	Output    tview.Primitive
	StatusBar *StatusBar

	Pages *tview.Pages

	outputHeight int
}

// NewLayout creates the main application layout
func NewLayout(menuBar *MenuBar, fileTree tview.Primitive, editor tview.Primitive, output tview.Primitive, statusBar *StatusBar) *Layout {
	l := &Layout{
		MenuBar:   menuBar,
		FileTree:  fileTree,
		Editor:    editor,
		Output:    output,
		StatusBar: statusBar,
	}

	// Middle section: file tree + editor
	middle := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(fileTree, 20, 0, false).
		AddItem(editor, 0, 1, true)

	// Main vertical layout: menu, middle, output, status
	// Output starts hidden (0 height), shown when content is added
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(menuBar, 1, 0, false).
		AddItem(middle, 0, 3, true).
		AddItem(output, 0, 0, false).
		AddItem(statusBar, 1, 0, false)

	mainFlex.SetBackgroundColor(ColorBg)

	l.MainGrid = mainFlex

	// Pages for dialog overlays
	l.Pages = tview.NewPages()
	l.Pages.AddPage("main", mainFlex, true, true)

	l.Root = tview.NewFlex().AddItem(l.Pages, 0, 1, true)

	return l
}

// ShowDialog shows a modal dialog overlay
func (l *Layout) ShowDialog(name string, dialog tview.Primitive) {
	l.Pages.AddPage(name, dialog, true, true)
}

// HideDialog removes a dialog overlay
func (l *Layout) HideDialog(name string) {
	l.Pages.RemovePage(name)
}

// HasDialog checks if a dialog is showing
func (l *Layout) HasDialog(name string) bool {
	return l.Pages.HasPage(name)
}

// SetOutputVisible shows or hides the output/terminal panel
func (l *Layout) SetOutputVisible(visible bool, height int) {
	if visible && height > 0 {
		if l.outputHeight == height {
			return
		}
		l.outputHeight = height
	} else {
		if l.outputHeight == 0 {
			return
		}
		l.outputHeight = 0
		height = 0
	}
	l.rebuildMainFlex()
}

// OutputVisible returns whether the output panel is visible
func (l *Layout) OutputVisible() bool {
	return l.outputHeight > 0
}

func (l *Layout) rebuildMainFlex() {
	l.MainGrid.Clear()

	middle := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(l.FileTree, 20, 0, false).
		AddItem(l.Editor, 0, 1, true)

	l.MainGrid.AddItem(l.MenuBar, 1, 0, false)
	l.MainGrid.AddItem(middle, 0, 3, true)
	l.MainGrid.AddItem(l.Output, l.outputHeight, 0, false)
	l.MainGrid.AddItem(l.StatusBar, 1, 0, false)
}
