package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestServer(t *testing.T) *Server {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			score INTEGER NOT NULL,
			time INTEGER NOT NULL,
			victory BOOLEAN NOT NULL,
			kills INTEGER NOT NULL,
			gamemode INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return &Server{db: db}
}

func TestPostAndGetScores(t *testing.T) {
	srv := setupTestServer(t)

	// Post a score
	entry := ScoreEntry{
		Name:     "TestPlayer",
		Score:    5000,
		Time:     120,
		Victory:  true,
		Kills:    42,
		GameMode: 0,
	}

	body, _ := json.Marshal(entry)
	req := httptest.NewRequest(http.MethodPost, "/api/scores", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	srv.handleScores(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %d", rec.Code)
	}

	var created ScoreEntry
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if created.Name != "TestPlayer" || created.Score != 5000 {
		t.Errorf("Unexpected created entry: %+v", created)
	}

	// Get scores
	getReq := httptest.NewRequest(http.MethodGet, "/api/scores?gamemode=0&limit=10", nil)
	getRec := httptest.NewRecorder()

	srv.handleScores(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", getRec.Code)
	}

	var list []ScoreEntry
	if err := json.NewDecoder(getRec.Body).Decode(&list); err != nil {
		t.Fatalf("Failed to decode score list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Expected 1 score, got %d", len(list))
	}
	if list[0].Name != "TestPlayer" {
		t.Errorf("Expected TestPlayer, got %s", list[0].Name)
	}
}

func TestGetRank(t *testing.T) {
	srv := setupTestServer(t)

	// Insert scores directly
	_, _ = srv.db.Exec("INSERT INTO scores (name, score, time, victory, kills, gamemode) VALUES ('A', 1000, 60, 0, 10, 0)")
	_, _ = srv.db.Exec("INSERT INTO scores (name, score, time, victory, kills, gamemode) VALUES ('B', 2000, 60, 0, 10, 0)")

	// Score 1500 should be rank #2 (behind 2000)
	req := httptest.NewRequest(http.MethodGet, "/api/rank?gamemode=0&score=1500", nil)
	rec := httptest.NewRecorder()

	srv.handleRank(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", rec.Code)
	}

	var res struct {
		Rank int `json:"rank"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode rank: %v", err)
	}

	if res.Rank != 2 {
		t.Errorf("Expected rank 2, got %d", res.Rank)
	}
}
