package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

type ScoreEntry struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name"`
	Score    int    `json:"score"`
	Time     int    `json:"time"`
	Victory  bool   `json:"victory"`
	Kills    int    `json:"kills"`
	GameMode int    `json:"gamemode"`
}

type HighscoreManager struct {
	serverURL  string
	httpClient *http.Client
	mu         sync.RWMutex
	TopScores  map[int][]ScoreEntry
}

func NewHighscoreManager() *HighscoreManager {
	serverURL := os.Getenv("SCORE_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:3000"
	}

	hm := &HighscoreManager{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		TopScores: make(map[int][]ScoreEntry),
	}

	hm.ReloadTopScores()
	return hm
}

func (hm *HighscoreManager) ReloadTopScores() {
	for mode := 0; mode <= 2; mode++ {
		url := fmt.Sprintf("%s/api/scores?gamemode=%d&limit=10", hm.serverURL, mode)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err == nil {
			resp, err := hm.httpClient.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				var scores []ScoreEntry
				if json.NewDecoder(resp.Body).Decode(&scores) == nil {
					hm.mu.Lock()
					hm.TopScores[mode] = scores
					hm.mu.Unlock()
				}
				resp.Body.Close()
			}
		}
		cancel()
	}
}

func (hm *HighscoreManager) AddScore(name string, score int, timeSpent int, victory bool, kills int, gamemode int) {
	entry := ScoreEntry{
		Name:     name,
		Score:    score,
		Time:     timeSpent,
		Victory:  victory,
		Kills:    kills,
		GameMode: gamemode,
	}

	// Update in-memory cache immediately for responsive UI
	hm.mu.Lock()
	hm.TopScores[gamemode] = append(hm.TopScores[gamemode], entry)
	sort.Slice(hm.TopScores[gamemode], func(i, j int) bool {
		return hm.TopScores[gamemode][i].Score > hm.TopScores[gamemode][j].Score
	})
	if len(hm.TopScores[gamemode]) > 10 {
		hm.TopScores[gamemode] = hm.TopScores[gamemode][:10]
	}
	hm.mu.Unlock()

	// Asynchronously submit to remote central server
	go func() {
		data, err := json.Marshal(entry)
		if err != nil {
			return
		}
		url := fmt.Sprintf("%s/api/scores", hm.serverURL)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := hm.httpClient.Do(req)
		if err == nil {
			resp.Body.Close()
			// Refresh top scores from central DB after post completes
			hm.ReloadTopScores()
		}
	}()
}

func (hm *HighscoreManager) GetTop(n int, gamemode int) []ScoreEntry {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	scores := hm.TopScores[gamemode]
	if len(scores) < n {
		return scores
	}
	return scores[:n]
}

func (hm *HighscoreManager) GetRank(score int, gamemode int) int {
	// Query remote server for rank
	url := fmt.Sprintf("%s/api/rank?gamemode=%d&score=%d", hm.serverURL, gamemode, score)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err == nil {
		resp, err := hm.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			var result struct {
				Rank int `json:"rank"`
			}
			if json.NewDecoder(resp.Body).Decode(&result) == nil {
				resp.Body.Close()
				return result.Rank
			}
			resp.Body.Close()
		}
	}

	// Fallback using in-memory cached top scores
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	rank := 1
	for _, entry := range hm.TopScores[gamemode] {
		if entry.Score > score {
			rank++
		}
	}
	return rank
}

func (hm *HighscoreManager) Close() {
	// No persistent resources to close in pure client mode
}
