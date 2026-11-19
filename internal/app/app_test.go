package app

import (
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/rjeff-sudo/termsnake-go/internal/game"
	"github.com/rjeff-sudo/termsnake-go/internal/input"
	"github.com/rjeff-sudo/termsnake-go/internal/storage"
	"github.com/rjeff-sudo/termsnake-go/internal/ui"
)

func newTestApp(t *testing.T) (*app, *storage.Store) {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("init screen: %v", err)
	}
	screen.SetSize(ui.MinWidth, ui.MinHeight)
	t.Cleanup(screen.Fini)

	store, err := storage.Open(filepath.Join(t.TempDir(), "scores.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	a := &app{
		screen: screen,
		ui:     ui.New(screen, false),
		store:  store,
		rng:    rand.New(rand.NewSource(1)),
		state:  stateMenu,
		name:   defaultName,
	}
	t.Cleanup(a.stopTicker)
	return a, store
}

func key(cmd input.Command) input.Event {
	return input.Event{Kind: input.Key, Cmd: cmd}
}

func TestMenuStartsAndPausesAGame(t *testing.T) {
	a, _ := newTestApp(t)

	if !a.handle(key(input.Confirm)) {
		t.Fatal("program quit from the menu")
	}
	if a.state != stateDifficulty {
		t.Fatalf("state = %v, want difficulty", a.state)
	}
	if !a.handle(key(input.Confirm)) {
		t.Fatal("program quit starting a game")
	}
	if a.state != statePlaying || a.game == nil || a.ticker == nil {
		t.Fatalf("game did not start: state=%v game=%v ticker=%v", a.state, a.game, a.ticker)
	}

	a.handle(key(input.Pause))
	if a.state != statePaused || a.ticker != nil {
		t.Fatalf("pause failed: state=%v ticker=%v", a.state, a.ticker)
	}
	a.handle(key(input.Pause))
	if a.state != statePlaying || a.ticker == nil {
		t.Fatalf("resume failed: state=%v ticker=%v", a.state, a.ticker)
	}

	a.handle(key(input.Back))
	if a.state != stateMenu {
		t.Fatalf("state = %v, want menu", a.state)
	}
}

func TestHighScoresShortcut(t *testing.T) {
	a, _ := newTestApp(t)
	a.handle(key(input.Down))
	if a.menuSel != menuHighScores {
		t.Fatalf("menuSel = %d, want %d", a.menuSel, menuHighScores)
	}
	a.handle(key(input.Confirm))
	if a.state != stateScores {
		t.Fatalf("state = %v, want scores", a.state)
	}
	a.handle(key(input.Back))
	if a.state != stateMenu {
		t.Fatalf("state = %v, want menu", a.state)
	}
}

func TestNameEntrySavesQualifyingScore(t *testing.T) {
	a, store := newTestApp(t)
	a.game = game.New(ui.BoardCols, ui.BoardRows, game.Normal, a.rng)
	a.game.Score = 40
	a.startedAt = time.Now()
	a.name = ""

	a.enterGameOver(false)
	if !a.overQualifies || a.overSaved {
		t.Fatalf("qualifying score not in name-entry state: qualifies=%v saved=%v",
			a.overQualifies, a.overSaved)
	}

	for _, r := range "ABC" {
		a.handle(input.Event{Kind: input.Key, Cmd: input.Left, Rune: r})
	}
	a.handle(key(input.Confirm))
	if !a.overSaved {
		t.Fatal("score was not saved after Enter")
	}

	scores, err := store.TopScores("normal", 10)
	if err != nil {
		t.Fatalf("TopScores: %v", err)
	}
	if len(scores) != 1 {
		t.Fatalf("saved %d scores, want 1", len(scores))
	}
	if scores[0].PlayerName != "ABC" || scores[0].Score != 40 {
		t.Fatalf("unexpected score row: %+v", scores[0])
	}
}
