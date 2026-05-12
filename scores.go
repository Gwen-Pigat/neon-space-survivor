package main

import (
	"encoding/json"
	"os"
	"sort"
)

type ScoreEntry struct {
	Name    string `json:"name"`
	Score   int    `json:"score"`
	Time    int    `json:"time"` // Seconds
	Victory bool   `json:"victory"`
	Kills   int    `json:"kills"`
}

type HighscoreManager struct {
	Scores []ScoreEntry `json:"scores"`
}

func NewHighscoreManager() *HighscoreManager {
	hm := &HighscoreManager{Scores: []ScoreEntry{}}
	hm.Load()
	return hm
}

func (hm *HighscoreManager) Load() {
	data, err := os.ReadFile("highscores.json")
	if err != nil {
		return
	}
	json.Unmarshal(data, hm)
}

func (hm *HighscoreManager) Save() {
	data, _ := json.Marshal(hm)
	os.WriteFile("highscores.json", data, 0644)
}

func (hm *HighscoreManager) AddScore(name string, score int, time int, victory bool, kills int) {
	hm.Scores = append(hm.Scores, ScoreEntry{Name: name, Score: score, Time: time, Victory: victory, Kills: kills})
	sort.Slice(hm.Scores, func(i, j int) bool {
		return hm.Scores[i].Score > hm.Scores[j].Score
	})
	if len(hm.Scores) > 10 {
		hm.Scores = hm.Scores[:10]
	}
	hm.Save()
}

func (hm *HighscoreManager) GetTop(n int) []ScoreEntry {
	if len(hm.Scores) < n {
		return hm.Scores
	}
	return hm.Scores[:n]
}
