package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	gameManager *GameManager
	upgrader    = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}
)

func main() {
	// Initialize game manager
	gameManager = NewGameManager()

	// Initialize database
	InitDB()
	defer CloseDB()

	// Initialize Kafka producer (optional)
	InitKafka()
	defer CloseKafka()

	// Setup routes
	r := mux.NewRouter()

	// WebSocket endpoint
	r.HandleFunc("/ws", handleWebSocket)

	// REST endpoints
	r.HandleFunc("/api/leaderboard", getLeaderboard).Methods("GET")
	r.HandleFunc("/api/analytics", getAnalytics).Methods("GET")

	// Serve static files (frontend)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	player := NewPlayer(conn, gameManager)
	gameManager.AddPlayer(player)

	go player.ReadMessages()
	go player.WriteMessages()
}

func getLeaderboard(w http.ResponseWriter, r *http.Request) {
	leaderboard := GetLeaderboard()
	respondJSON(w, leaderboard)
}

func getAnalytics(w http.ResponseWriter, r *http.Request) {
	analytics := GetAnalytics()
	respondJSON(w, analytics)
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
