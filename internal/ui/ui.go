// Package ui renders the game with tcell. It knows nothing about the game
// loop; the app package decides what to draw and when.
package ui

import (
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

const (
	// BoardCols and BoardRows are the size of the play field in cells.
	BoardCols = 20
	BoardRows = 20
	// MinWidth and MinHeight are the smallest terminal the game draws in.
	MinWidth  = 44
	MinHeight = 24

	// Every cell is two columns wide so it looks roughly square.
	cellWidth  = 2
	frameW     = BoardCols*cellWidth + 2
	frameH     = BoardRows + 2
	contentH   = frameH + 2 // HUD row above, controls row below
	contentW   = frameW
	modalWidth = 40
)

// Layout is where the board and its furniture sit on screen. All coordinates
// are absolute terminal cells.
type Layout struct {
	FrameX, FrameY int // top-left corner of the border
	BoardX, BoardY int // top-left interior cell
	HUDY           int // row of the score line
	HintY          int // row of the controls hint
}

// ComputeLayout centres the board on a w x h terminal.
func ComputeLayout(w, h int) Layout {
	left := (w - contentW) / 2
	if left < 0 {
		left = 0
	}
	top := (h - contentH) / 2
	if top < 0 {
		top = 0
	}
	return Layout{
		FrameX: left,
		FrameY: top + 1,
		BoardX: left + 1,
		BoardY: top + 2,
		HUDY:   top,
		HintY:  top + 1 + frameH,
	}
}

// UI draws screens on a tcell screen.
type UI struct {
	screen tcell.Screen
	theme  Theme
}

// New returns a UI bound to an initialised screen.
func New(screen tcell.Screen, noColor bool) *UI {
	return &UI{screen: screen, theme: NewTheme(noColor)}
}

// Size reports the current terminal size.
func (u *UI) Size() (int, int) {
	return u.screen.Size()
}

// Clear blanks the screen buffer.
func (u *UI) Clear() {
	u.screen.Clear()
}

// Show flushes the buffer to the terminal.
func (u *UI) Show() {
	u.screen.Show()
}

// Sync forces a full redraw, which clears artifacts after a resize.
func (u *UI) Sync() {
	u.screen.Sync()
}

// text writes s starting at (x, y).
func (u *UI) text(x, y int, s string, style tcell.Style) {
	for _, r := range s {
		u.screen.SetContent(x, y, r, nil, style)
		x++
	}
}

// centerText writes s horizontally centred on the terminal.
func (u *UI) centerText(y int, s string, style tcell.Style) {
	w, _ := u.screen.Size()
	u.text((w-runeLen(s))/2, y, s, style)
}

// centerIn writes s centred inside the box that starts at x and is width wide.
func (u *UI) centerIn(x, width, y int, s string, style tcell.Style) {
	u.text(x+(width-runeLen(s))/2, y, s, style)
}

// fill paints a rectangle with spaces in the given style.
func (u *UI) fill(x, y, width, height int, style tcell.Style) {
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			u.screen.SetContent(x+col, y+row, ' ', nil, style)
		}
	}
}

// box draws a border and blanks the inside.
func (u *UI) box(x, y, width, height int, style tcell.Style) {
	u.fill(x, y, width, height, u.theme.Base)
	for col := 1; col < width-1; col++ {
		u.screen.SetContent(x+col, y, '─', nil, style)
		u.screen.SetContent(x+col, y+height-1, '─', nil, style)
	}
	for row := 1; row < height-1; row++ {
		u.screen.SetContent(x, y+row, '│', nil, style)
		u.screen.SetContent(x+width-1, y+row, '│', nil, style)
	}
	u.screen.SetContent(x, y, '┌', nil, style)
	u.screen.SetContent(x+width-1, y, '┐', nil, style)
	u.screen.SetContent(x, y+height-1, '└', nil, style)
	u.screen.SetContent(x+width-1, y+height-1, '┘', nil, style)
}

// menuBar renders one selectable line, padded to width. The selected line is
// reversed so it is visible even without color support.
func (u *UI) menuBar(x, width, y int, label string, selected bool) {
	style := u.theme.Base
	prefix := "  "
	if selected {
		style = u.theme.Selected
		prefix = "▶ "
	}
	line := prefix + label
	if runeLen(line) < width {
		line += strings.Repeat(" ", width-runeLen(line))
	}
	u.text(x, y, line, style)
}

// runeLen counts display runes. The board glyphs are all single-width.
func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}
