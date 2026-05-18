package main

import (
	"database/sql"
	"fmt"
	"sort"

	_ "modernc.org/sqlite"
)

type ScoreEntry struct {
	Name     string
	Score    int
	Time     int // Seconds
	Victory  bool
	Kills    int
	GameMode int
}

type HighscoreManager struct {
	db        *sql.DB
	TopScores map[int][]ScoreEntry
}

func NewHighscoreManager() *HighscoreManager {
	hm := &HighscoreManager{
		TopScores: make(map[int][]ScoreEntry),
	}

	db, err := sql.Open("sqlite", "scores.db")
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		return hm
	}
	hm.db = db

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			score INTEGER,
			time INTEGER,
			victory BOOLEAN,
			kills INTEGER,
			gamemode INTEGER
		)
	`)
	if err != nil {
		fmt.Printf("Failed to create scores table: %v\n", err)
	}

	hm.ReloadTopScores()
	return hm
}

func (hm *HighscoreManager) ReloadTopScores() {
	if hm.db == nil {
		return
	}
	// Preload for modes 0, 1, 2
	for mode := 0; mode <= 2; mode++ {
		rows, err := hm.db.Query("SELECT name, score, time, victory, kills, gamemode FROM scores WHERE gamemode = ? ORDER BY score DESC LIMIT 10", mode)
		if err != nil {
			fmt.Printf("Query error for mode %d: %v\n", mode, err)
			continue
		}
		
		var scores []ScoreEntry
		for rows.Next() {
			var entry ScoreEntry
			err = rows.Scan(&entry.Name, &entry.Score, &entry.Time, &entry.Victory, &entry.Kills, &entry.GameMode)
			if err == nil {
				scores = append(scores, entry)
			}
		}
		rows.Close()
		hm.TopScores[mode] = scores
	}
}

func (hm *HighscoreManager) AddScore(name string, score int, time int, victory bool, kills int, gamemode int) {
	if hm.db != nil {
		_, err := hm.db.Exec("INSERT INTO scores (name, score, time, victory, kills, gamemode) VALUES (?, ?, ?, ?, ?, ?)", name, score, time, victory, kills, gamemode)
		if err != nil {
			fmt.Printf("Failed to insert score: %v\n", err)
		}
		hm.ReloadTopScores()
	} else {
		// Fallback for memory-only
		hm.TopScores[gamemode] = append(hm.TopScores[gamemode], ScoreEntry{Name: name, Score: score, Time: time, Victory: victory, Kills: kills, GameMode: gamemode})
		sort.Slice(hm.TopScores[gamemode], func(i, j int) bool {
			return hm.TopScores[gamemode][i].Score > hm.TopScores[gamemode][j].Score
		})
		if len(hm.TopScores[gamemode]) > 10 {
			hm.TopScores[gamemode] = hm.TopScores[gamemode][:10]
		}
	}
}

func (hm *HighscoreManager) GetTop(n int, gamemode int) []ScoreEntry {
	scores := hm.TopScores[gamemode]
	if len(scores) < n {
		return scores
	}
	return scores[:n]
}

func (hm *HighscoreManager) Close() {
	if hm.db != nil {
		hm.db.Close()
	}
}
