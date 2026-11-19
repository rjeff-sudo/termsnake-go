// Package game implements the rules of Snake.
//
// It deliberately imports nothing terminal-related and performs no I/O:
// callers feed directions in and call Tick to advance the game one step.
// That keeps the rules deterministic and easy to unit test.
package game

import (
	"math/rand"
	"time"
)

const (
	// PointsPerFood is added to the score for every pellet eaten.
	PointsPerFood = 10
	// SpeedUpEvery is the number of pellets eaten between speed-ups.
	SpeedUpEvery = 5
	// SpeedUpStep is how much time each speed-up removes from the tick.
	SpeedUpStep = 5 * time.Millisecond
	// MinInterval is the fastest the snake is allowed to move.
	MinInterval = 40 * time.Millisecond
	// QueueLimit is how many direction changes may wait for the next tick.
	QueueLimit = 2
)

// Point is a cell coordinate on the board. X grows to the right, Y grows down.
type Point struct {
	X, Y int
}

// Add returns the point offset by q.
func (p Point) Add(q Point) Point {
	return Point{X: p.X + q.X, Y: p.Y + q.Y}
}

// Direction is the direction the snake is travelling in.
type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

// Delta is the grid step for the direction.
func (d Direction) Delta() Point {
	switch d {
	case Up:
		return Point{X: 0, Y: -1}
	case Down:
		return Point{X: 0, Y: 1}
	case Left:
		return Point{X: -1, Y: 0}
	default:
		return Point{X: 1, Y: 0}
	}
}

// Opposite returns the direction that would reverse the snake.
func (d Direction) Opposite() Direction {
	switch d {
	case Up:
		return Down
	case Down:
		return Up
	case Left:
		return Right
	default:
		return Left
	}
}

// String returns the direction name used in logs and tests.
func (d Direction) String() string {
	switch d {
	case Up:
		return "up"
	case Down:
		return "down"
	case Left:
		return "left"
	default:
		return "right"
	}
}

// Difficulty controls the starting tick rate.
type Difficulty int

const (
	Easy Difficulty = iota
	Normal
	Hard
)

// String returns the database-friendly name of the difficulty.
func (d Difficulty) String() string {
	switch d {
	case Easy:
		return "easy"
	case Hard:
		return "hard"
	default:
		return "normal"
	}
}

// ParseDifficulty converts a stored name back into a Difficulty.
func ParseDifficulty(s string) (Difficulty, bool) {
	switch s {
	case "easy":
		return Easy, true
	case "normal":
		return Normal, true
	case "hard":
		return Hard, true
	default:
		return Normal, false
	}
}

// BaseInterval is the tick length at the start of a game.
func (d Difficulty) BaseInterval() time.Duration {
	switch d {
	case Easy:
		return 150 * time.Millisecond
	case Hard:
		return 60 * time.Millisecond
	default:
		return 100 * time.Millisecond
	}
}

// TickInterval returns the tick length after eaten pellets, speeding up every
// SpeedUpEvery pellets until it reaches MinInterval.
func TickInterval(d Difficulty, eaten int) time.Duration {
	if eaten < 0 {
		eaten = 0
	}
	interval := d.BaseInterval() - time.Duration(eaten/SpeedUpEvery)*SpeedUpStep
	if interval < MinInterval {
		return MinInterval
	}
	return interval
}

// Status describes how a game ended.
type Status int

const (
	Running Status = iota
	Dead
	Won
)

// Game is the complete state of one round of Snake.
type Game struct {
	Width, Height int
	Difficulty    Difficulty
	Snake         []Point // head first
	Dir           Direction
	Food          Point
	Score         int
	Eaten         int
	Status        Status

	rng   *rand.Rand
	queue []Direction
}

// New creates a game on a width x height board. Pass a seeded rand.Rand from
// tests for deterministic food placement; nil uses a time-seeded source.
func New(width, height int, difficulty Difficulty, rng *rand.Rand) *Game {
	if width < 4 {
		width = 4
	}
	if height < 4 {
		height = 4
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	cx, cy := width/2, height/2
	head := Point{X: cx, Y: cy}
	g := &Game{
		Width:      width,
		Height:     height,
		Difficulty: difficulty,
		Snake: []Point{
			head,
			{X: cx - 1, Y: cy},
			{X: cx - 2, Y: cy},
		},
		Dir:    Right,
		Status: Running,
		rng:    rng,
	}
	g.spawnFood()
	return g
}

// Head returns the position of the snake's head.
func (g *Game) Head() Point {
	return g.Snake[0]
}

// Interval returns the current tick length for this game.
func (g *Game) Interval() time.Duration {
	return TickInterval(g.Difficulty, g.Eaten)
}

// SetDirection queues a direction change for the next tick.
//
// Reversing straight back onto the neck is ignored, as is changing direction
// more than QueueLimit times before the snake has moved. Each queued direction
// is validated against the previous queued one, not just the current heading.
func (g *Game) SetDirection(d Direction) {
	if g.Status != Running || d < Up || d > Right || len(g.queue) >= QueueLimit {
		return
	}
	last := g.Dir
	if len(g.queue) > 0 {
		last = g.queue[len(g.queue)-1]
	}
	if d == last || d == last.Opposite() {
		return
	}
	g.queue = append(g.queue, d)
}

// Tick advances the snake by one cell, applying the queued direction, eating
// food if the head lands on it, and ending the game on collision or a full
// board.
func (g *Game) Tick() {
	if g.Status != Running {
		return
	}
	if len(g.queue) > 0 {
		g.Dir = g.queue[0]
		g.queue = g.queue[1:]
	}

	next := g.Head().Add(g.Dir.Delta())
	if next.X < 0 || next.X >= g.Width || next.Y < 0 || next.Y >= g.Height {
		g.Status = Dead
		return
	}

	eating := next == g.Food
	// When the snake is not growing the tail cell is vacated this tick, so
	// moving into it is not a collision.
	check := len(g.Snake)
	if !eating {
		check--
	}
	for i := 0; i < check; i++ {
		if g.Snake[i] == next {
			g.Status = Dead
			return
		}
	}

	g.Snake = append([]Point{next}, g.Snake...)
	if !eating {
		g.Snake = g.Snake[:len(g.Snake)-1]
		return
	}

	g.Score += PointsPerFood
	g.Eaten++
	if len(g.Snake) == g.Width*g.Height {
		g.Status = Won
		g.Food = Point{X: -1, Y: -1}
		return
	}
	g.spawnFood()
}

// spawnFood picks uniformly from the empty cells, so it always terminates and
// never places food under the snake.
func (g *Game) spawnFood() {
	occupied := make(map[Point]struct{}, len(g.Snake))
	for _, p := range g.Snake {
		occupied[p] = struct{}{}
	}
	empty := make([]Point, 0, g.Width*g.Height-len(g.Snake))
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			p := Point{X: x, Y: y}
			if _, taken := occupied[p]; !taken {
				empty = append(empty, p)
			}
		}
	}
	if len(empty) == 0 {
		g.Food = Point{X: -1, Y: -1}
		return
	}
	g.Food = empty[g.rng.Intn(len(empty))]
}
