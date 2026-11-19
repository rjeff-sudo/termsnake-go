package ui

import (
	"fmt"

	"github.com/rjeff-sudo/termsnake-go/internal/game"
)

var titleArt = []string{
	"███████ ██   ██   ███   ██   ██ ███████",
	"██      ███  ██  ██ ██  ██  ██  ██     ",
	"███████ ██ █ ██ ███████ █████   █████  ",
	"     ██ ██  ███ ██   ██ ██  ██  ██     ",
	"███████ ██   ██ ██   ██ ██   ██ ███████",
}

var menuLabels = []string{"PLAY", "HIGH SCORES", "QUIT"}

// DrawMenu renders the title screen. crawl is an animation frame counter for
// the little snake under the title.
func (u *UI) DrawMenu(sel, crawl int) {
	w, h := u.Size()
	top := max(h/2-9, 1)
	for i, line := range titleArt {
		u.centerText(top+i, line, u.theme.Title)
	}
	u.centerText(top+6, "TERMINAL SNAKE IN GO", u.theme.HUD)

	if w > 12 {
		x := crawl%(w+10) - 5
		for i := 0; i < 6; i++ {
			if x+i >= 0 && x+i < w {
				u.text(x+i, top+8, "█", u.theme.Body)
			}
		}
		if x+6 >= 0 && x+6 < w {
			u.text(x+6, top+8, "►", u.theme.Head)
		}
	}

	barW := 24
	barX := (w - barW) / 2
	itemsY := top + 10
	for i, label := range menuLabels {
		u.menuBar(barX, barW, itemsY+i*2, label, i == sel)
	}
	u.centerText(h-2, "↑↓ navigate · Enter select · Q quit", u.theme.Muted)
}

// DrawDifficulty renders the difficulty picker.
func (u *UI) DrawDifficulty(sel int) {
	w, h := u.Size()
	u.centerText(3, "SELECT DIFFICULTY", u.theme.Title)

	items := []struct {
		name string
		desc string
	}{
		{"EASY", "150 ms per step · relaxed"},
		{"NORMAL", "100 ms per step · classic"},
		{"HARD", "60 ms per step · twitchy"},
	}
	barW := 30
	barX := (w - barW) / 2
	startY := h/2 - 3
	for i, item := range items {
		y := startY + i*3
		u.menuBar(barX, barW, y, item.name, i == sel)
		descStyle := u.theme.Muted
		if i == sel {
			descStyle = u.theme.Base
		}
		u.centerText(y+1, item.desc, descStyle)
	}
	u.centerText(h/2+6, fmt.Sprintf("Speed increases every %d foods.", game.SpeedUpEvery), u.theme.Muted)
	u.centerText(h-2, "↑↓ choose · Enter start · Q back", u.theme.Muted)
}

// ScoreRow is one rendered line on the high-score table.
type ScoreRow struct {
	Rank       int
	Name       string
	Score      int
	Difficulty string
	Date       string
}

// DrawScores renders the top-ten table, or note when scores are unavailable.
func (u *UI) DrawScores(rows []ScoreRow, note string) {
	w, h := u.Size()
	u.centerText(2, "HIGH SCORES", u.theme.Title)
	u.centerText(h-2, "Q back", u.theme.Muted)

	if note != "" {
		u.centerText(h/2, note, u.theme.Muted)
		return
	}
	if len(rows) == 0 {
		u.centerText(h/2, "No scores yet — play a game!", u.theme.Muted)
		return
	}

	tableW := 39
	x := (w - tableW) / 2
	header := fmt.Sprintf("%2s  %5s  %-10s %-6s %s", "#", "SCORE", "PLAYER", "MODE", "DATE")
	u.text(x, 5, header, u.theme.HUD)
	for i, row := range rows {
		if i >= 10 {
			break
		}
		name := truncate(row.Name, 10)
		line := fmt.Sprintf("%2d  %5d  %-10s %-6s %s",
			row.Rank, row.Score, name, truncate(row.Difficulty, 6), truncate(row.Date, 10))
		u.text(x, 6+i, line, u.theme.Base)
	}
}

// DrawPaused renders the pause menu over the board.
func (u *UI) DrawPaused(sel int) {
	w, h := u.Size()
	bw, bh := 30, 9
	x, y := (w-bw)/2, (h-bh)/2
	u.box(x, y, bw, bh, u.theme.Wall)
	u.centerIn(x, bw, y+1, "PAUSED", u.theme.Title)

	labels := []string{"RESUME", "RESTART", "QUIT TO MENU"}
	for i, label := range labels {
		u.menuBar(x+3, bw-6, y+3+i, label, i == sel)
	}
	u.centerIn(x, bw, y+bh-2, "P resume · Q menu", u.theme.Muted)
}

// DrawGameOver renders the end-of-round modal, including optional name entry.
func (u *UI) DrawGameOver(won bool, score, high int, newHigh bool, name string, editing, saved bool, note string) {
	w, h := u.Size()
	bw := min(modalWidth, w-2)
	bh := 13
	x, y := (w-bw)/2, (h-bh)/2
	u.box(x, y, bw, bh, u.theme.Wall)

	title, titleStyle := "GAME OVER", u.theme.Danger
	if won {
		title, titleStyle = "YOU WIN!", u.theme.Highlight
	}
	u.centerIn(x, bw, y+1, title, titleStyle)
	u.centerIn(x, bw, y+3, fmt.Sprintf("Score: %d    High: %d", score, high), u.theme.HUD)
	if newHigh {
		u.centerIn(x, bw, y+5, "NEW HIGH SCORE!", u.theme.Highlight)
	}

	switch {
	case editing:
		u.centerIn(x, bw, y+7, "Name: "+name+"_", u.theme.Score)
		u.centerIn(x, bw, y+8, "3-12 characters · Enter saves", u.theme.Muted)
	case saved:
		u.centerIn(x, bw, y+7, "Score saved.", u.theme.Muted)
	case note != "":
		u.centerIn(x, bw, y+7, note, u.theme.Muted)
	}
	u.centerIn(x, bw, y+bh-2, "R restart · Q menu", u.theme.Muted)
}

// DrawTooSmall is the friendly blocker shown when the terminal cannot fit the
// board.
func (u *UI) DrawTooSmall(w, h int) {
	u.centerText(max(h/2-2, 0), "TERMINAL TOO SMALL", u.theme.Danger)
	u.centerText(h/2, fmt.Sprintf("Need at least %dx%d, have %dx%d.", MinWidth, MinHeight, w, h), u.theme.Base)
	u.centerText(h/2+1, "Resize the window to continue.", u.theme.Muted)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
