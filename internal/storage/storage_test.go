package storage

import (
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestSaveAndTopScores(t *testing.T) {
	store := openTestStore(t)

	runs := []Score{
		{PlayerName: "ALICE", Score: 30, Difficulty: "easy", SnakeLen: 6, DurationMS: 4000},
		{PlayerName: "BOB", Score: 90, Difficulty: "normal", SnakeLen: 12, DurationMS: 9000},
		{PlayerName: "CAROL", Score: 60, Difficulty: "normal", SnakeLen: 9, DurationMS: 7000},
	}
	for _, run := range runs {
		if err := store.SaveScore(run); err != nil {
			t.Fatalf("SaveScore: %v", err)
		}
	}

	all, err := store.TopScores("", 10)
	if err != nil {
		t.Fatalf("TopScores: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("got %d scores, want 3", len(all))
	}
	if all[0].Score != 90 || all[1].Score != 60 || all[2].Score != 30 {
		t.Fatalf("scores out of order: %+v", all)
	}
	if all[0].PlayerName != "BOB" || all[0].Difficulty != "normal" {
		t.Fatalf("wrong row contents: %+v", all[0])
	}
	if all[0].ID == 0 || all[0].CreatedAt == "" {
		t.Fatalf("id/created_at not filled in: %+v", all[0])
	}

	normal, err := store.TopScores("normal", 10)
	if err != nil {
		t.Fatalf("TopScores(normal): %v", err)
	}
	if len(normal) != 2 {
		t.Fatalf("got %d normal scores, want 2", len(normal))
	}

	limited, err := store.TopScores("", 2)
	if err != nil {
		t.Fatalf("TopScores(limit 2): %v", err)
	}
	if len(limited) != 2 || limited[0].Score != 90 || limited[1].Score != 60 {
		t.Fatalf("limit not applied: %+v", limited)
	}
}

func TestTieScoresUseEarliestFirst(t *testing.T) {
	store := openTestStore(t)

	// Insert raw rows so the timestamps are deterministic.
	for _, createdAt := range []string{"2026-02-02 10:00:00", "2026-02-01 10:00:00"} {
		if _, err := store.db.Exec(
			`INSERT INTO scores (player_name, score, difficulty, snake_len, duration_ms, created_at)
			 VALUES ('TIE', 50, 'easy', 8, 1000, ?)`, createdAt,
		); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	scores, err := store.TopScores("easy", 10)
	if err != nil {
		t.Fatalf("TopScores: %v", err)
	}
	if len(scores) != 2 {
		t.Fatalf("got %d scores, want 2", len(scores))
	}
	if scores[0].CreatedAt != "2026-02-01 10:00:00" {
		t.Fatalf("tie not ordered by earliest date: %+v", scores)
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	store := openTestStore(t)

	if got, err := store.GetSetting("player_name"); err != nil || got != "" {
		t.Fatalf("missing setting = %q, %v; want empty, nil", got, err)
	}
	if err := store.SetSetting("player_name", "JEFF"); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	if err := store.SetSetting("player_name", "JEFF2"); err != nil {
		t.Fatalf("SetSetting overwrite: %v", err)
	}
	if got, err := store.GetSetting("player_name"); err != nil || got != "JEFF2" {
		t.Fatalf("setting = %q, %v; want JEFF2, nil", got, err)
	}
}

func TestOpenRejectsEmptyPath(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("Open(\"\") succeeded, want error")
	}
}
