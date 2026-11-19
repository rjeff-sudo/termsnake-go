package ui

import "github.com/gdamore/tcell/v2"

// Theme is the set of styles used to draw the game. It only uses the 16 ANSI
// colors so it works on limited terminals.
type Theme struct {
	Base      tcell.Style
	Title     tcell.Style
	Head      tcell.Style
	Body      tcell.Style
	Food      tcell.Style
	FoodDim   tcell.Style
	Wall      tcell.Style
	HUD       tcell.Style
	Score     tcell.Style
	Danger    tcell.Style
	Highlight tcell.Style
	Muted     tcell.Style
	Selected  tcell.Style
}

// NewTheme returns the default palette, or an attribute-only palette when
// noColor is set.
func NewTheme(noColor bool) Theme {
	if noColor {
		return Theme{
			Base:      tcell.StyleDefault,
			Title:     tcell.StyleDefault.Bold(true),
			Head:      tcell.StyleDefault.Bold(true),
			Body:      tcell.StyleDefault,
			Food:      tcell.StyleDefault.Bold(true),
			FoodDim:   tcell.StyleDefault.Dim(true),
			Wall:      tcell.StyleDefault.Dim(true),
			HUD:       tcell.StyleDefault.Bold(true),
			Score:     tcell.StyleDefault.Bold(true),
			Danger:    tcell.StyleDefault.Bold(true),
			Highlight: tcell.StyleDefault.Bold(true),
			Muted:     tcell.StyleDefault.Dim(true),
			Selected:  tcell.StyleDefault.Reverse(true),
		}
	}
	return Theme{
		Base:      tcell.StyleDefault.Foreground(tcell.ColorSilver),
		Title:     tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true),
		Head:      tcell.StyleDefault.Foreground(tcell.ColorLime).Bold(true),
		Body:      tcell.StyleDefault.Foreground(tcell.ColorGreen),
		Food:      tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true),
		FoodDim:   tcell.StyleDefault.Foreground(tcell.ColorMaroon),
		Wall:      tcell.StyleDefault.Foreground(tcell.ColorGray),
		HUD:       tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true),
		Score:     tcell.StyleDefault.Foreground(tcell.ColorYellow),
		Danger:    tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true),
		Highlight: tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
		Muted:     tcell.StyleDefault.Foreground(tcell.ColorGray),
		Selected:  tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorSilver),
	}
}
