package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rjeff-sudo/termsnake-go/internal/game"
)

// DrawGame renders the HUD, board border, food and snake.
func (u *UI) DrawGame(g *game.Game, high int, flash, foodVisible bool) {
	l := ComputeLayout(u.Size())
	u.drawHUD(g, high, l)
	u.drawFrame(l)

	if g.Status == game.Running {
		style := u.theme.Food
		if !foodVisible {
			style = u.theme.FoodDim
		}
		u.boardCell(l, g.Food, "●", style)
	}

	headStyle, bodyStyle := u.theme.Head, u.theme.Body
	if flash {
		headStyle, bodyStyle = u.theme.Danger, u.theme.Danger
	}
	for i, p := range g.Snake {
		if i == 0 {
			u.drawHead(l, p, g.Dir, headStyle)
			continue
		}
		u.boardCell(l, p, "██", bodyStyle)
	}
}

// DrawControls writes the one-line controls hint under the board.
func (u *UI) DrawControls() {
	l := ComputeLayout(u.Size())
	hint := "ARROWS/WASD move · P pause · Q quit"
	u.centerIn(l.FrameX, frameW, l.HintY, hint, u.theme.Muted)
}

// DrawPopup shows a short score message near a board cell, e.g. "+10".
func (u *UI) DrawPopup(at game.Point, text string) {
	l := ComputeLayout(u.Size())
	x := l.BoardX + at.X*cellWidth
	y := l.BoardY + at.Y
	if at.Y > 0 {
		y--
	}
	if maxX := l.FrameX + frameW - 1 - runeLen(text); x > maxX {
		x = maxX
	}
	u.text(x, y, text, u.theme.Score)
}

func (u *UI) drawHUD(g *game.Game, high int, l Layout) {
	u.text(l.FrameX, l.HUDY, "SNAKE", u.theme.HUD)

	if g.Score > high {
		high = g.Score
	}
	score := fmt.Sprintf("Score: %d", g.Score)
	best := fmt.Sprintf("High: %d", high)
	right := l.FrameX + frameW
	u.text(right-runeLen(best), l.HUDY, best, u.theme.Score)
	u.text(right-runeLen(best)-runeLen(score)-3, l.HUDY, score, u.theme.Score)
}

func (u *UI) drawFrame(l Layout) {
	style := u.theme.Wall
	for col := 1; col < frameW-1; col++ {
		u.screen.SetContent(l.FrameX+col, l.FrameY, '─', nil, style)
		u.screen.SetContent(l.FrameX+col, l.FrameY+frameH-1, '─', nil, style)
	}
	for row := 1; row < frameH-1; row++ {
		u.screen.SetContent(l.FrameX, l.FrameY+row, '│', nil, style)
		u.screen.SetContent(l.FrameX+frameW-1, l.FrameY+row, '│', nil, style)
	}
	u.screen.SetContent(l.FrameX, l.FrameY, '┌', nil, style)
	u.screen.SetContent(l.FrameX+frameW-1, l.FrameY, '┐', nil, style)
	u.screen.SetContent(l.FrameX, l.FrameY+frameH-1, '└', nil, style)
	u.screen.SetContent(l.FrameX+frameW-1, l.FrameY+frameH-1, '┘', nil, style)
}

func (u *UI) boardCell(l Layout, p game.Point, glyph string, style tcell.Style) {
	u.text(l.BoardX+p.X*cellWidth, l.BoardY+p.Y, glyph, style)
}

// drawHead places the arrow so it touches the neck: on the right half when
// moving left, on the left half otherwise.
func (u *UI) drawHead(l Layout, p game.Point, dir game.Direction, style tcell.Style) {
	x := l.BoardX + p.X*cellWidth
	if dir == game.Left {
		x++
	}
	u.text(x, l.BoardY+p.Y, headGlyph(dir), style)
}

func headGlyph(d game.Direction) string {
	switch d {
	case game.Up:
		return "▲"
	case game.Down:
		return "▼"
	case game.Left:
		return "◄"
	default:
		return "►"
	}
}
