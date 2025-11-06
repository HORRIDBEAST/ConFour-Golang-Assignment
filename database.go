package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/connect4?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Database connection error: %v", err)
		return
	}

	if err = db.Ping(); err != nil {
		log.Printf("Database ping error: %v", err)
		return
	}

	createTables()
	log.Println("Database connected successfully")
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS games (
			id VARCHAR(255) PRIMARY KEY,
			player1 VARCHAR(255) NOT NULL,
			player2 VARCHAR(255) NOT NULL,
			winner VARCHAR(255),
			board JSONB,
			is_bot BOOLEAN DEFAULT FALSE,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			duration FLOAT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS players (
			username VARCHAR(255) PRIMARY KEY,
			games_played INT DEFAULT 0,
			games_won INT DEFAULT 0,
			games_lost INT DEFAULT 0,
			games_drawn INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_games_start_time ON games(start_time)`,
		`CREATE INDEX IF NOT EXISTS idx_players_games_won ON players(games_won DESC)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			log.Printf("Create table error: %v", err)
		}
	}
}

func SaveGame(game *Game) {
	if db == nil {
		return
	}

	boardJSON, _ := json.Marshal(game.Board)
	duration := game.EndTime.Sub(game.StartTime).Seconds()

	winner := ""
	if game.Winner == Player1 {
		winner = game.Player1.Username
	} else if game.Winner == Player2 {
		winner = game.getPlayerName(game.Player2)
	} else {
		winner = "Draw"
	}

	_, err := db.Exec(`
		INSERT INTO games (id, player1, player2, winner, board, is_bot, start_time, end_time, duration)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, game.ID, game.Player1.Username, game.getPlayerName(game.Player2), winner, boardJSON, game.IsBot, game.StartTime, game.EndTime, duration)

	if err != nil {
		log.Printf("Save game error: %v", err)
	}
}

func UpdatePlayerStats(username string, won bool) {
	if db == nil {
		return
	}

	// Insert or update player
	_, err := db.Exec(`
		INSERT INTO players (username, games_played, games_won, games_lost)
		VALUES ($1, 1, $2, $3)
		ON CONFLICT (username) DO UPDATE SET
			games_played = players.games_played + 1,
			games_won = players.games_won + $2,
			games_lost = players.games_lost + $3,
			updated_at = CURRENT_TIMESTAMP
	`, username, boolToInt(won), boolToInt(!won))

	if err != nil {
		log.Printf("Update player stats error: %v", err)
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

type LeaderboardEntry struct {
	Username    string  `json:"username"`
	GamesPlayed int     `json:"gamesPlayed"`
	GamesWon    int     `json:"gamesWon"`
	GamesLost   int     `json:"gamesLost"`
	WinRate     float64 `json:"winRate"`
}

func GetLeaderboard() []LeaderboardEntry {
	if db == nil {
		return []LeaderboardEntry{}
	}

	rows, err := db.Query(`
        SELECT username, games_played, games_won, games_lost,
               CASE WHEN games_played > 0 THEN ROUND(((games_won::float / games_played::float) * 100)::numeric, 2) ELSE 0 END as win_rate
        FROM players
        ORDER BY games_won DESC, win_rate DESC
        LIMIT 10
    `)
	if err != nil {
		log.Printf("Get leaderboard error: %v", err)
		return []LeaderboardEntry{}
	}
	defer rows.Close()

	leaderboard := []LeaderboardEntry{}
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.Username, &entry.GamesPlayed, &entry.GamesWon, &entry.GamesLost, &entry.WinRate); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		leaderboard = append(leaderboard, entry)
	}

	return leaderboard
}

type Analytics struct {
	TotalGames         int     `json:"totalGames"`
	AvgDuration        float64 `json:"avgDuration"`
	BotGames           int     `json:"botGames"`
	PlayerGames        int     `json:"playerGames"`
	GamesToday         int     `json:"gamesToday"`
	MostFrequentWinner string  `json:"mostFrequentWinner"`
}

func GetAnalytics() Analytics {
	if db == nil {
		return Analytics{}
	}

	var analytics Analytics

	// Total games and average duration
	db.QueryRow(`
		SELECT COUNT(*), COALESCE(AVG(duration), 0)
		FROM games
	`).Scan(&analytics.TotalGames, &analytics.AvgDuration)

	// Bot vs Player games
	db.QueryRow(`SELECT COUNT(*) FROM games WHERE is_bot = true`).Scan(&analytics.BotGames)
	db.QueryRow(`SELECT COUNT(*) FROM games WHERE is_bot = false`).Scan(&analytics.PlayerGames)

	// Games today
	today := time.Now().Format("2006-01-02")
	db.QueryRow(`
		SELECT COUNT(*) FROM games
		WHERE DATE(start_time) = $1
	`, today).Scan(&analytics.GamesToday)

	// Most frequent winner
	db.QueryRow(`
		SELECT winner FROM games
		WHERE winner != 'Draw'
		GROUP BY winner
		ORDER BY COUNT(*) DESC
		LIMIT 1
	`).Scan(&analytics.MostFrequentWinner)

	return analytics
}
