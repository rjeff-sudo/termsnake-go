package ui

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rjeff-sudo/termsnake-go/internal/game"
)

func newSim(t *testing.T, w, h int) tcell.SimulationScreen {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("init screen: %v", err)
	}
	screen.SetSize(w, h)
	t.Cleanup(screen.Fini)
	return screen
}

func screenText(screen tcell.SimulationScreen) string {
	cells, w, h := screen.GetContents()
	var b strings.Builder
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := cells[y*w+x]
			if len(cell.Runes) == 0 {
				b.WriteRune(' ')
				continue
			}
			b.WriteRune(cell.Runes[0])
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func TestScreensRenderAtMinimumSize(t *testing.T) {
	screen := newSim(t, MinWidth, MinHeight)
	u := New(screen, false)

	u.DrawMenu(0, 3)
	u.Show()
	text := screenText(screen)
	for _, want := range []string{"PLAY", "HIGH SCORES", "QUIT", "TERMINAL SNAKE IN GO"} {
		if !strings.Contains(text, want) {
			t.Errorf("menu missing %q\n%s", want, text)
		}
	}

	u.DrawDifficulty(1)
	u.Show()
	text = screenText(screen)
	for _, want := range []string{"SELECT DIFFICULTY", "EASY", "NORMAL", "HARD", "Speed increases"} {
		if !strings.Contains(text, want) {
			t.Errorf("difficulty screen missing %q\n%s", want, text)
		}
	}

	g := game.New(BoardCols, BoardRows, game.Normal, rand.New(rand.NewSource(1)))
	u.DrawGame(g, 120, false, true)
	u.DrawControls()
	u.DrawPopup(game.Point{X: 5, Y: 5}, "+10")
	u.Show()
	text = screenText(screen)
	for _, want := range []string{"SNAKE", "Score: 0", "High: 120", "ARROWS/WASD move", "┌", "►", "█", "●"} {
		if !strings.Contains(text, want) {
			t.Errorf("game screen missing %q\n%s", want, text)
		}
	}

	u.DrawPaused(0)
	u.Show()
	text = screenText(screen)
	for _, want := range []string{"PAUSED", "RESUME", "RESTART", "QUIT TO MENU"} {
		if !strings.Contains(text, want) {
			t.Errorf("pause screen missing %q\n%s", want, text)
		}
	}

	u.DrawGameOver(false, 40, 120, true, "PLAYER", true, false, "")
	u.Show()
	text = screenText(screen)
	for _, want := range []string{"GAME OVER", "Score: 40", "High: 120", "NEW HIGH SCORE!", "Name: PLAYER_"} {
		if !strings.Contains(text, want) {
			t.Errorf("game over screen missing %q\n%s", want, text)
		}
	}

	u.DrawScores([]ScoreRow{
		{Rank: 1, Name: "ALICE", Score: 90, Difficulty: "normal", Date: "2026-11-19"},
	}, "")
	u.Show()
	text = screenText(screen)
	for _, want := range []string{"HIGH SCORES", "ALICE", "2026-11-19", "SCORE"} {
		if !strings.Contains(text, want) {
			t.Errorf("scores screen missing %q\n%s", want, text)
		}
	}

	u.DrawTooSmall(30, 10)
	u.Show()
	if text := screenText(screen); !strings.Contains(text, "TERMINAL TOO SMALL") {
		t.Errorf("small screen missing warning\n%s", text)
	}
}

func TestScreensRenderWideWithoutColor(t *testing.T) {
	screen := newSim(t, 120, 40)
	u := New(screen, true)
	u.DrawMenu(2, 50)
	u.Show()
	if text := screenText(screen); !strings.Contains(text, "QUIT") {
		t.Errorf("wide menu missing quit\n%s", text)
	}

	g := game.New(BoardCols, BoardRows, game.Hard, rand.New(rand.NewSource(2)))
	u.DrawGame(g, 0, false, false)
	u.DrawControls()
	u.Show()
	if text := screenText(screen); !strings.Contains(text, "└") {
		t.Errorf("wide board missing frame\n%s", text)
	}
}
