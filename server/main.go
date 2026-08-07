package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type ScoreEntry struct {
	ID        int       `json:"id,omitempty"`
	Name      string    `json:"name"`
	Score     int       `json:"score"`
	Time      int       `json:"time"`
	Victory   bool      `json:"victory"`
	Kills     int       `json:"kills"`
	GameMode  int       `json:"gamemode"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type Server struct {
	db *sql.DB
	mu sync.RWMutex
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "server_scores.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database at %s: %v", dbPath, err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			score INTEGER NOT NULL,
			time INTEGER NOT NULL,
			victory BOOLEAN NOT NULL,
			kills INTEGER NOT NULL,
			gamemode INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_mode_score ON scores(gamemode, score DESC);
	`)
	if err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	srv := &Server{db: db}

	http.HandleFunc("/api/scores", srv.handleScores)
	http.HandleFunc("/api/rank", srv.handleRank)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fmt.Printf("Neon Space Survivor Leaderboard Server running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func (s *Server) enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *Server) handleScores(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getScores(w, r)
	case http.MethodPost:
		s.postScore(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getScores(w http.ResponseWriter, r *http.Request) {
	modeStr := r.URL.Query().Get("gamemode")
	limitStr := r.URL.Query().Get("limit")

	mode := 0
	if modeStr != "" {
		parsedMode, err := strconv.Atoi(modeStr)
		if err == nil && parsedMode >= 0 && parsedMode <= 2 {
			mode = parsedMode
		}
	}

	limit := 10
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		"SELECT id, name, score, time, victory, kills, gamemode FROM scores WHERE gamemode = ? ORDER BY score DESC, time ASC LIMIT ?",
		mode, limit,
	)
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		log.Printf("Query error: %v", err)
		return
	}
	defer rows.Close()

	scores := make([]ScoreEntry, 0)
	for rows.Next() {
		var entry ScoreEntry
		if err := rows.Scan(&entry.ID, &entry.Name, &entry.Score, &entry.Time, &entry.Victory, &entry.Kills, &entry.GameMode); err == nil {
			scores = append(scores, entry)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

func (s *Server) postScore(w http.ResponseWriter, r *http.Request) {
	var entry ScoreEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Input sanitization & validation
	entry.Name = strings.TrimSpace(entry.Name)
	if entry.Name == "" {
		entry.Name = "Anonymous"
	}
	if len(entry.Name) > 20 {
		entry.Name = entry.Name[:20]
	}
	if entry.Score < 0 {
		entry.Score = 0
	}
	if entry.GameMode < 0 || entry.GameMode > 2 {
		entry.GameMode = 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	res, err := s.db.Exec(
		"INSERT INTO scores (name, score, time, victory, kills, gamemode) VALUES (?, ?, ?, ?, ?, ?)",
		entry.Name, entry.Score, entry.Time, entry.Victory, entry.Kills, entry.GameMode,
	)
	if err != nil {
		http.Error(w, "Failed to save score", http.StatusInternalServerError)
		log.Printf("Insert error: %v", err)
		return
	}

	id, _ := res.LastInsertId()
	entry.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

func (s *Server) handleRank(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	modeStr := r.URL.Query().Get("gamemode")
	scoreStr := r.URL.Query().Get("score")

	mode := 0
	if modeStr != "" {
		if parsedMode, err := strconv.Atoi(modeStr); err == nil && parsedMode >= 0 && parsedMode <= 2 {
			mode = parsedMode
		}
	}

	score := 0
	if scoreStr != "" {
		if parsedScore, err := strconv.Atoi(scoreStr); err == nil {
			score = parsedScore
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var rank int
	err := s.db.QueryRow("SELECT COUNT(*) + 1 FROM scores WHERE gamemode = ? AND score > ?", mode, score).Scan(&rank)
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"rank": rank})
}
