// Package input translates terminal events into commands the game loop
// understands.
package input

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

// Command is a semantic action requested by the player.
type Command int

const (
	None Command = iota
	Up
	Down
	Left
	Right
	Pause
	Restart
	Confirm
	Back
	Backspace
	Quit // Ctrl+C: leave immediately, restoring the terminal
)

// Kind separates key presses from terminal resizes.
type Kind int

const (
	Key Kind = iota
	Resize
)

// Event is one thing the UI thread needs to know about.
type Event struct {
	Kind          Kind
	Cmd           Command
	Rune          rune // printable rune for name entry, or 0
	Width, Height int  // board of the terminal, for Resize events
}

// Listen starts a goroutine that polls the screen and sends events on a
// buffered channel. The goroutine stops when ctx is cancelled or the screen
// closes.
func Listen(ctx context.Context, screen tcell.Screen) <-chan Event {
	events := make(chan Event, 32)
	go func() {
		defer close(events)
		for {
			switch ev := screen.PollEvent().(type) {
			case nil:
				return
			case *tcell.EventResize:
				width, height := ev.Size()
				if !send(ctx, events, Event{Kind: Resize, Width: width, Height: height}) {
					return
				}
			case *tcell.EventKey:
				if !send(ctx, events, Translate(ev)) {
					return
				}
			}
		}
	}()
	return events
}

// send delivers ev unless the context is done first.
func send(ctx context.Context, events chan<- Event, ev Event) bool {
	select {
	case events <- ev:
		return true
	case <-ctx.Done():
		return false
	}
}

// Translate maps a tcell key event to a command. Letter keys carry their rune
// as well, because the name-entry field needs raw characters.
func Translate(ev *tcell.EventKey) Event {
	out := Event{Kind: Key}
	if ev.Key() == tcell.KeyRune {
		r := ev.Rune()
		out.Rune = r
		switch r {
		case 'w', 'W':
			out.Cmd = Up
		case 's', 'S':
			out.Cmd = Down
		case 'a', 'A':
			out.Cmd = Left
		case 'd', 'D':
			out.Cmd = Right
		case 'p', 'P', ' ':
			out.Cmd = Pause
		case 'r', 'R':
			out.Cmd = Restart
		case 'q', 'Q':
			out.Cmd = Back
		}
		return out
	}

	switch ev.Key() {
	case tcell.KeyUp:
		out.Cmd = Up
	case tcell.KeyDown:
		out.Cmd = Down
	case tcell.KeyLeft:
		out.Cmd = Left
	case tcell.KeyRight:
		out.Cmd = Right
	case tcell.KeyEnter:
		out.Cmd = Confirm
	case tcell.KeyEsc:
		out.Cmd = Back
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		out.Cmd = Backspace
	case tcell.KeyCtrlC:
		out.Cmd = Quit
	}
	return out
}
