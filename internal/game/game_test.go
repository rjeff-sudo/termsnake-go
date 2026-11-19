package game

import (
	"math/rand"
	"testing"
)

// fixedGame builds a deterministic game with a known snake and food position.
func fixedGame(t *testing.T, snake []Point, dir Direction, food Point, width, height int) *Game {
	t.Helper()
	g := New(width, height, Normal, rand.New(rand.NewSource(1)))
	g.Snake = append([]Point(nil), snake...)
	g.Dir = dir
	g.Food = food
	g.Status = Running
	g.Score = 0
	g.Eaten = 0
	return g
}

func TestTickMovesInDirection(t *testing.T) {
	tests := []struct {
		name  string
		snake []Point
		dir   Direction
		want  Point
	}{
		{"up", []Point{{5, 5}, {5, 6}, {5, 7}}, Up, Point{5, 4}},
		{"down", []Point{{5, 5}, {5, 4}, {5, 3}}, Down, Point{5, 6}},
		{"left", []Point{{5, 5}, {6, 5}, {7, 5}}, Left, Point{4, 5}},
		{"right", []Point{{5, 5}, {4, 5}, {3, 5}}, Right, Point{6, 5}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			start := tc.snake
			g := fixedGame(t, start, tc.dir, Point{0, 0}, 20, 20)
			g.Tick()
			if g.Status != Running {
				t.Fatalf("status = %v, want Running", g.Status)
			}
			if got := g.Head(); got != tc.want {
				t.Fatalf("head = %v, want %v", got, tc.want)
			}
			if len(g.Snake) != len(start) {
				t.Fatalf("length = %d, want %d", len(g.Snake), len(start))
			}
		})
	}
}

func TestTickGrowsAndScoresOnFood(t *testing.T) {
	snake := []Point{{5, 5}, {4, 5}, {3, 5}}
	g := fixedGame(t, snake, Right, Point{6, 5}, 20, 20)

	g.Tick()

	if g.Status != Running {
		t.Fatalf("status = %v, want Running", g.Status)
	}
	if len(g.Snake) != 4 {
		t.Fatalf("length = %d, want 4", len(g.Snake))
	}
	if g.Score != PointsPerFood {
		t.Fatalf("score = %d, want %d", g.Score, PointsPerFood)
	}
	if g.Eaten != 1 {
		t.Fatalf("eaten = %d, want 1", g.Eaten)
	}
	if g.Food == (Point{-1, -1}) {
		t.Fatal("food was not respawned")
	}
	for _, p := range g.Snake {
		if g.Food == p {
			t.Fatalf("food spawned on snake at %v", p)
		}
	}
}

func TestTickWallCollision(t *testing.T) {
	tests := []struct {
		name  string
		snake []Point
		dir   Direction
	}{
		{"left wall", []Point{{0, 5}, {1, 5}, {2, 5}}, Left},
		{"right wall", []Point{{19, 5}, {18, 5}, {17, 5}}, Right},
		{"top wall", []Point{{5, 0}, {5, 1}, {5, 2}}, Up},
		{"bottom wall", []Point{{5, 19}, {5, 18}, {5, 17}}, Down},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := fixedGame(t, tc.snake, tc.dir, Point{10, 10}, 20, 20)
			before := append([]Point(nil), g.Snake...)
			g.Tick()
			if g.Status != Dead {
				t.Fatalf("status = %v, want Dead", g.Status)
			}
			if len(g.Snake) != len(before) || g.Snake[0] != before[0] {
				t.Fatal("snake moved after a fatal collision")
			}
		})
	}
}

func TestTickSelfCollision(t *testing.T) {
	// The head at (2,1) turns down into (2,2), which is body but not the tail.
	snake := []Point{{2, 1}, {1, 1}, {1, 2}, {2, 2}, {3, 2}}
	g := fixedGame(t, snake, Down, Point{10, 10}, 20, 20)

	g.Tick()

	if g.Status != Dead {
		t.Fatalf("status = %v, want Dead", g.Status)
	}
}

func TestMovingIntoVacatingTailIsAllowed(t *testing.T) {
	// Head at (2,1) moves down into (2,2), which is the tail and will move away.
	snake := []Point{{2, 1}, {1, 1}, {1, 2}, {2, 2}}
	g := fixedGame(t, snake, Down, Point{10, 10}, 20, 20)

	g.Tick()

	if g.Status != Running {
		t.Fatalf("status = %v, want Running", g.Status)
	}
	if g.Head() != (Point{2, 2}) {
		t.Fatalf("head = %v, want {2 2}", g.Head())
	}
}

func TestReversalIsIgnored(t *testing.T) {
	g := fixedGame(t, []Point{{5, 5}, {4, 5}, {3, 5}}, Right, Point{10, 10}, 20, 20)

	g.SetDirection(Left)
	g.Tick()

	if g.Dir != Right {
		t.Fatalf("dir = %v, want right", g.Dir)
	}
	if got := g.Head(); got != (Point{6, 5}) {
		t.Fatalf("head = %v, want {6 5}", got)
	}
}

func TestDirectionQueue(t *testing.T) {
	g := fixedGame(t, []Point{{5, 5}, {4, 5}, {3, 5}}, Right, Point{10, 10}, 20, 20)

	g.SetDirection(Up)
	g.SetDirection(Left)
	g.SetDirection(Down) // queue is full, must be ignored

	g.Tick()
	if g.Dir != Up || g.Head() != (Point{5, 4}) {
		t.Fatalf("after first tick dir = %v head = %v, want up {5 4}", g.Dir, g.Head())
	}

	g.Tick()
	if g.Dir != Left || g.Head() != (Point{4, 4}) {
		t.Fatalf("after second tick dir = %v head = %v, want left {4 4}", g.Dir, g.Head())
	}
}

func TestQueuedOppositeRejected(t *testing.T) {
	g := fixedGame(t, []Point{{5, 5}, {4, 5}, {3, 5}}, Right, Point{10, 10}, 20, 20)

	g.SetDirection(Up)
	g.SetDirection(Down) // opposite of the queued Up, not merely of Right

	g.Tick()

	if g.Dir != Up {
		t.Fatalf("dir = %v, want up", g.Dir)
	}
}

func TestFoodNeverSpawnsOnSnake(t *testing.T) {
	for seed := int64(0); seed < 200; seed++ {
		rng := rand.New(rand.NewSource(seed))
		g := New(20, 20, Normal, rng)
		if g.Food.X < 0 || g.Food.X >= g.Width || g.Food.Y < 0 || g.Food.Y >= g.Height {
			t.Fatalf("seed %d: food %v outside board", seed, g.Food)
		}
		for _, p := range g.Snake {
			if g.Food == p {
				t.Fatalf("seed %d: food %v spawned on snake", seed, g.Food)
			}
		}
	}
}

func TestFoodUsesLastEmptyCell(t *testing.T) {
	g := fixedGame(t, nil, Right, Point{0, 0}, 4, 4)
	// Fill the board except for (2, 2).
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if x == 2 && y == 2 {
				continue
			}
			g.Snake = append(g.Snake, Point{X: x, Y: y})
		}
	}

	g.spawnFood()

	if g.Food != (Point{2, 2}) {
		t.Fatalf("food = %v, want {2 2}", g.Food)
	}
}

func TestFillingBoardWins(t *testing.T) {
	// A 4x4 board with one free cell, (3,3), right in front of the head.
	snake := []Point{
		{2, 3}, {1, 3}, {0, 3},
		{0, 2}, {1, 2}, {2, 2}, {3, 2},
		{3, 1}, {2, 1}, {1, 1}, {0, 1},
		{0, 0}, {1, 0}, {2, 0}, {3, 0},
	}
	g := fixedGame(t, snake, Right, Point{3, 3}, 4, 4)

	g.Tick()

	if g.Status != Won {
		t.Fatalf("status = %v, want Won", g.Status)
	}
	if len(g.Snake) != 16 {
		t.Fatalf("length = %d, want 16", len(g.Snake))
	}
}

func TestTickAfterGameOverIsNoop(t *testing.T) {
	g := fixedGame(t, []Point{{0, 5}, {1, 5}, {2, 5}}, Left, Point{10, 10}, 20, 20)
	g.Tick()
	head := g.Head()
	g.Tick()
	if g.Head() != head {
		t.Fatal("game advanced after it was over")
	}
}

func TestTickInterval(t *testing.T) {
	tests := []struct {
		difficulty Difficulty
		eaten      int
		wantMS     int
	}{
		{Easy, 0, 150},
		{Normal, 0, 100},
		{Hard, 0, 60},
		{Normal, 4, 100},
		{Normal, 5, 95},
		{Normal, 12, 90},
		{Easy, 30, 120},
		{Hard, 100, 40},
		{Normal, -3, 100},
	}
	for _, tc := range tests {
		got := TickInterval(tc.difficulty, tc.eaten)
		if got.Milliseconds() != int64(tc.wantMS) {
			t.Errorf("TickInterval(%v, %d) = %v, want %dms", tc.difficulty, tc.eaten, got, tc.wantMS)
		}
	}
}

func TestParseDifficulty(t *testing.T) {
	tests := []struct {
		in    string
		want  Difficulty
		valid bool
	}{
		{"easy", Easy, true},
		{"normal", Normal, true},
		{"hard", Hard, true},
		{"nope", Normal, false},
	}
	for _, tc := range tests {
		got, ok := ParseDifficulty(tc.in)
		if ok != tc.valid || got != tc.want {
			t.Errorf("ParseDifficulty(%q) = %v, %v; want %v, %v", tc.in, got, ok, tc.want, tc.valid)
		}
	}
}
