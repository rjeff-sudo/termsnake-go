// Package app wires the game rules, terminal UI, input and storage into the
// running program. Only the loop in Run mutates game state, so no locks are
// needed between input and the ticker.
package app

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"

	"github.com/rjeff-sudo/termsnake-go/internal/game"
	"github.com/rjeff-sudo/termsnake-go/internal/input"
	"github.com/rjeff-sudo/termsnake-go/internal/storage"
	"github.com/rjeff-sudo/termsnake-go/internal/ui"
)

const (
	animInterval  = 100 * time.Millisecond
	blinkInterval = 500 * time.Millisecond
	flashDuration = 250 * time.Millisecond
	popupDuration = 600 * time.Millisecond

	minNameLen  = 3
	maxNameLen  = 12
	defaultName = "PLAYER"
)

const (
	menuPlay = iota
	menuHighScores
	menuQuit
	menuCount
)

const (
	pauseResume = iota
	pauseRestart
	pauseQuit
)

type state int

const (
	stateMenu state = iota
	stateDifficulty
	statePlaying
	statePaused
	stateDeath
	stateGameOver
	stateWin
	stateScores
)

type popup struct {
	at    game.Point
	text  string
	until time.Time
}

type app struct {
	screen tcell.Screen
	ui     *ui.UI
	store  *storage.Store
	events <-chan input.Event
	rng    *rand.Rand

	state    state
	menuSel  int
	diffSel  int
	pauseSel int
	crawl    int

	game         *game.Game
	high         int
	ticker       *time.Ticker
	startedAt    time.Time
	pausedFor    time.Duration
	pauseStarted time.Time
	popup        *popup
	flashUntil   time.Time

	overScore     int
	overHigh      int
	overQualifies bool
	overWon       bool
	overSaved     bool
	scoreMessage  string
	name          string

	scoreRows  []ui.ScoreRow
	scoresNote string
}

// Run takes over the terminal until the player quits. The screen is always
// restored, including on Ctrl+C, SIGTERM and panics.
func Run(screen tcell.Screen, store *storage.Store, noColor bool) (err error) {
	if err := screen.Init(); err != nil {
		return fmt.Errorf("init screen: %w", err)
	}
	defer screen.Fini()
	defer func() {
		if r := recover(); r != nil {
			screen.Fini()
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	screen.SetStyle(tcell.StyleDefault)
	screen.HideCursor()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a := &app{
		screen: screen,
		ui:     ui.New(screen, noColor),
		store:  store,
		events: input.Listen(ctx, screen),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
		state:  stateMenu,
		name:   defaultName,
	}
	a.name = a.loadName()

	anim := time.NewTicker(animInterval)
	defer anim.Stop()

	for {
		var tick <-chan time.Time
		if a.ticker != nil {
			tick = a.ticker.C
		}
		select {
		case <-ctx.Done():
			return nil
		case <-anim.C:
			a.animate(time.Now())
		case ev, ok := <-a.events:
			if !ok {
				return nil
			}
			if !a.handle(ev) {
				return nil
			}
		case <-tick:
			a.tick(time.Now())
		}
		a.render()
		screen.Show()
	}
}

// handle processes one event. It returns false when the program should exit.
func (a *app) handle(ev input.Event) bool {
	// Resizes are handled first so a shrinking terminal always pauses play,
	// even while the friendly "too small" message is showing.
	if ev.Kind == input.Resize {
		a.screen.Sync()
		if a.state == statePlaying {
			a.pause()
		}
		return true
	}

	w, h := a.screen.Size()
	tooSmall := w < ui.MinWidth || h < ui.MinHeight
	if ev.Cmd == input.Quit {
		return false
	}
	if tooSmall {
		return true
	}

	switch a.state {
	case stateMenu:
		return a.handleMenu(ev)
	case stateDifficulty:
		return a.handleDifficulty(ev)
	case statePlaying:
		return a.handlePlaying(ev)
	case statePaused:
		return a.handlePaused(ev)
	case stateDeath:
		return true // let the death flash finish
	case stateGameOver, stateWin:
		return a.handleGameOver(ev)
	case stateScores:
		if ev.Cmd == input.Back || ev.Cmd == input.Confirm {
			a.state = stateMenu
		}
	}
	return true
}

func (a *app) handleMenu(ev input.Event) bool {
	switch ev.Cmd {
	case input.Up:
		a.menuSel = (a.menuSel + menuCount - 1) % menuCount
	case input.Down:
		a.menuSel = (a.menuSel + 1) % menuCount
	case input.Confirm:
		switch a.menuSel {
		case menuPlay:
			a.diffSel = int(game.Normal)
			a.state = stateDifficulty
		case menuHighScores:
			a.loadScores()
			a.state = stateScores
		case menuQuit:
			return false
		}
	case input.Back:
		return false
	}
	return true
}

func (a *app) handleDifficulty(ev input.Event) bool {
	switch ev.Cmd {
	case input.Up:
		a.diffSel = (a.diffSel + 2) % 3
	case input.Down:
		a.diffSel = (a.diffSel + 1) % 3
	case input.Confirm:
		a.startGame(game.Difficulty(a.diffSel))
	case input.Back:
		a.state = stateMenu
	}
	return true
}

func (a *app) handlePlaying(ev input.Event) bool {
	switch ev.Cmd {
	case input.Up:
		a.game.SetDirection(game.Up)
	case input.Down:
		a.game.SetDirection(game.Down)
	case input.Left:
		a.game.SetDirection(game.Left)
	case input.Right:
		a.game.SetDirection(game.Right)
	case input.Pause:
		a.pause()
	case input.Restart:
		a.startGame(a.game.Difficulty)
	case input.Back:
		a.toMenu()
	}
	return true
}

func (a *app) handlePaused(ev input.Event) bool {
	switch ev.Cmd {
	case input.Up:
		a.pauseSel = (a.pauseSel + 2) % 3
	case input.Down:
		a.pauseSel = (a.pauseSel + 1) % 3
	case input.Confirm:
		switch a.pauseSel {
		case pauseResume:
			a.resume()
		case pauseRestart:
			a.startGame(a.game.Difficulty)
		case pauseQuit:
			a.toMenu()
		}
	case input.Pause:
		a.resume()
	case input.Back:
		a.toMenu()
	}
	return true
}

func (a *app) handleGameOver(ev input.Event) bool {
	editing := a.overQualifies && !a.overSaved
	if editing {
		if ev.Rune != 0 && unicode.IsPrint(ev.Rune) {
			if len([]rune(a.name)) < maxNameLen {
				a.name += string(ev.Rune)
			}
			return true
		}
		switch ev.Cmd {
		case input.Backspace:
			if runes := []rune(a.name); len(runes) > 0 {
				a.name = string(runes[:len(runes)-1])
			}
		case input.Confirm:
			a.saveCurrentScore()
		case input.Restart:
			a.saveCurrentScore()
			a.startGame(a.game.Difficulty)
		case input.Back:
			a.saveCurrentScore()
			a.toMenu()
		}
		return true
	}

	switch ev.Cmd {
	case input.Restart:
		a.startGame(a.game.Difficulty)
	case input.Back:
		a.toMenu()
	}
	return true
}

// animate advances time-based visuals and releases the death flash.
func (a *app) animate(now time.Time) {
	a.crawl++
	if a.popup != nil && !now.Before(a.popup.until) {
		a.popup = nil
	}
	if a.state == stateDeath && !now.Before(a.flashUntil) {
		a.enterGameOver(false)
	}
}

// tick advances the game by one step.
func (a *app) tick(now time.Time) {
	if a.state != statePlaying || a.game == nil {
		return
	}
	food := a.game.Food
	before := a.game.Score
	a.game.Tick()

	if a.game.Score > before {
		a.popup = &popup{
			at:    food,
			text:  fmt.Sprintf("+%d", game.PointsPerFood),
			until: now.Add(popupDuration),
		}
	}

	switch a.game.Status {
	case game.Dead:
		a.state = stateDeath
		a.flashUntil = now.Add(flashDuration)
		a.stopTicker()
	case game.Won:
		a.enterGameOver(true)
	default:
		if a.ticker != nil {
			a.ticker.Reset(a.game.Interval())
		}
	}
}

func (a *app) startGame(difficulty game.Difficulty) {
	a.stopTicker()
	a.game = game.New(ui.BoardCols, ui.BoardRows, difficulty, a.rng)
	a.high = a.bestScore(difficulty)
	a.startedAt = time.Now()
	a.pausedFor = 0
	a.pauseStarted = time.Time{}
	a.popup = nil
	a.flashUntil = time.Time{}
	a.overQualifies = false
	a.overSaved = false
	a.scoreMessage = ""
	a.state = statePlaying
	a.startTicker()
}

func (a *app) pause() {
	if a.state != statePlaying {
		return
	}
	a.state = statePaused
	a.pauseSel = pauseResume
	a.stopTicker()
	a.pauseStarted = time.Now()
}

func (a *app) resume() {
	if a.state != statePaused {
		return
	}
	a.pausedFor += time.Since(a.pauseStarted)
	a.pauseStarted = time.Time{}
	a.state = statePlaying
	a.startTicker()
}

func (a *app) toMenu() {
	a.stopTicker()
	a.popup = nil
	a.state = stateMenu
}

func (a *app) startTicker() {
	a.stopTicker()
	a.ticker = time.NewTicker(a.game.Interval())
}

func (a *app) stopTicker() {
	if a.ticker != nil {
		a.ticker.Stop()
		a.ticker = nil
	}
}

func (a *app) enterGameOver(won bool) {
	a.stopTicker()
	a.overWon = won
	a.overScore = a.game.Score
	a.overHigh = max(a.high, a.overScore)
	a.overQualifies = a.qualifies(a.game.Difficulty, a.overScore)
	a.overSaved = false
	a.scoreMessage = ""
	a.state = stateGameOver
	if won {
		a.state = stateWin
	}
	if !a.overQualifies {
		a.saveCurrentScore()
	}
}

// qualifies reports whether a run makes the local top ten for its difficulty.
func (a *app) qualifies(difficulty game.Difficulty, score int) bool {
	if a.store == nil || score <= 0 {
		return false
	}
	scores, err := a.store.TopScores(difficulty.String(), 10)
	if err != nil {
		return false
	}
	return len(scores) < 10 || score > scores[len(scores)-1].Score
}

// saveCurrentScore stores the finished run once, showing a friendly message if
// the database is unavailable.
func (a *app) saveCurrentScore() {
	if a.overSaved {
		return
	}
	a.overSaved = true

	name := a.cleanName()
	if a.store == nil {
		a.scoreMessage = "scores unavailable"
		return
	}
	err := a.store.SaveScore(storage.Score{
		PlayerName: name,
		Score:      a.overScore,
		Difficulty: a.game.Difficulty.String(),
		SnakeLen:   len(a.game.Snake),
		DurationMS: a.playDuration().Milliseconds(),
	})
	if err != nil {
		a.scoreMessage = "scores unavailable"
		return
	}
	a.name = name
	a.scoreMessage = "Score saved."
	_ = a.store.SetSetting("player_name", name)
}

func (a *app) loadName() string {
	if a.store == nil {
		return defaultName
	}
	name, err := a.store.GetSetting("player_name")
	if err != nil || name == "" {
		return defaultName
	}
	return name
}

func (a *app) bestScore(difficulty game.Difficulty) int {
	if a.store == nil {
		return 0
	}
	scores, err := a.store.TopScores(difficulty.String(), 1)
	if err != nil || len(scores) == 0 {
		return 0
	}
	return scores[0].Score
}

func (a *app) loadScores() {
	a.scoreRows = nil
	a.scoresNote = ""
	if a.store == nil {
		a.scoresNote = "scores unavailable"
		return
	}
	scores, err := a.store.TopScores("", 10)
	if err != nil {
		a.scoresNote = "scores unavailable"
		return
	}
	for i, sc := range scores {
		date := sc.CreatedAt
		if len(date) > 10 {
			date = date[:10]
		}
		a.scoreRows = append(a.scoreRows, ui.ScoreRow{
			Rank:       i + 1,
			Name:       sc.PlayerName,
			Score:      sc.Score,
			Difficulty: sc.Difficulty,
			Date:       date,
		})
	}
}

func (a *app) playDuration() time.Duration {
	d := time.Since(a.startedAt) - a.pausedFor
	if a.state == statePaused && !a.pauseStarted.IsZero() {
		d -= time.Since(a.pauseStarted)
	}
	if d < 0 {
		return 0
	}
	return d
}

// cleanName enforces the 3-12 character rule, falling back to PLAYER.
func (a *app) cleanName() string {
	name := strings.TrimSpace(a.name)
	if len([]rune(name)) < minNameLen {
		return defaultName
	}
	if len([]rune(name)) > maxNameLen {
		return string([]rune(name)[:maxNameLen])
	}
	return name
}

func (a *app) render() {
	w, h := a.screen.Size()
	a.screen.Clear()

	if w < ui.MinWidth || h < ui.MinHeight {
		a.ui.DrawTooSmall(w, h)
		return
	}

	switch a.state {
	case stateMenu:
		a.ui.DrawMenu(a.menuSel, a.crawl)
	case stateDifficulty:
		a.ui.DrawDifficulty(a.diffSel)
	case statePlaying, statePaused, stateDeath, stateGameOver, stateWin:
		a.renderGame()
		switch a.state {
		case statePaused:
			a.ui.DrawPaused(a.pauseSel)
		case stateGameOver, stateWin:
			editing := a.overQualifies && !a.overSaved
			a.ui.DrawGameOver(a.overWon, a.overScore, a.overHigh, a.overQualifies,
				a.name, editing, a.overSaved, a.scoreMessage)
		}
	case stateScores:
		a.ui.DrawScores(a.scoreRows, a.scoresNote)
	}
}

func (a *app) renderGame() {
	if a.game == nil {
		return
	}
	now := time.Now()
	foodVisible := (now.UnixMilli()/blinkInterval.Milliseconds())%2 == 0
	a.ui.DrawGame(a.game, a.high, a.state == stateDeath, foodVisible)
	a.ui.DrawControls()
	if a.popup != nil {
		a.ui.DrawPopup(a.popup.at, a.popup.text)
	}
}
