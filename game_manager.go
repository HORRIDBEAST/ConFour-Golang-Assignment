package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	matchmakingTimeout = 10 * time.Second
	reconnectTimeout   = 30 * time.Second
)

// GameManager manages all active games and players.
type GameManager struct {
	players       map[string]*Player // Keyed by username
	games         map[string]*Game   // Keyed by game ID
	waitingPlayer *Player
	mutex         sync.RWMutex
}

// NewGameManager creates a new game manager.
func NewGameManager() *GameManager {
	return &GameManager{
		players: make(map[string]*Player),
		games:   make(map[string]*Game),
	}
}

// AddPlayer adds a new player to the manager.
func (gm *GameManager) AddPlayer(player *Player) {
	player.SendMessage("waiting", nil) // Tell client we're waiting for join message
}

// UnregisterPlayer removes a player from the manager.
func (gm *GameManager) UnregisterPlayer(player *Player) {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	if player.Username == "" {
		return // Player never fully joined
	}

	delete(gm.players, player.Username)

	// If player was in the waiting lobby
	if gm.waitingPlayer == player {
		gm.waitingPlayer = nil
		log.Printf("Waiting player %s disconnected.", player.Username)
	}

	// If player was in a game
	if player.Game != nil {
		log.Printf("Player %s disconnected from game %s.", player.Username, player.Game.ID)
		player.Game.HandleDisconnect(player)
	}

	close(player.Send)
}

// HandleMessage routes messages from players to the correct handler.
func (gm *GameManager) HandleMessage(player *Player, rawMsg []byte) {
	var msg Message
	if err := json.Unmarshal(rawMsg, &msg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		player.SendError("Invalid message format.")
		return
	}

	switch msg.Type {
	case "join":
		gm.handleJoin(player, msg.Username)
	case "move":
		gm.handleMove(player, msg.Column)
	case "reconnect":
		gm.handleReconnect(player, msg.Username)
	default:
		player.SendError("Unknown message type.")
	}
}

// handleJoin processes a new player's request to join a game.
func (gm *GameManager) handleJoin(player *Player, username string) {
	if username == "" {
		player.SendError("Username cannot be empty.")
		return
	}

	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	if _, exists := gm.players[username]; exists {
		// This could be a reconnect attempt, but "join" is for new games
		player.SendError("Username already taken. Try 'reconnect' if you were in a game.")
		return
	}

	log.Printf("Player %s joining.", username)
	player.Username = username
	gm.players[username] = player

	if gm.waitingPlayer == nil {
		// This is the first player, make them wait
		gm.waitingPlayer = player
		player.SendMessage("waiting", nil)

		// Start the 10-second bot timer
		time.AfterFunc(matchmakingTimeout, func() {
			gm.startBotGame(player)
		})
	} else {
		// A waiting player exists, start a game
		if gm.waitingPlayer == player {
			return // Should not happen, but safeguard
		}
		opponent := gm.waitingPlayer
		gm.waitingPlayer = nil
		gm.startGame(opponent, player)
	}
}

// handleMove passes a move to the player's active game.
func (gm *GameManager) handleMove(player *Player, col int) {
	if player.Game == nil {
		player.SendError("You are not in a game.")
		return
	}
	player.Game.HandleMove(player, col)
}

// handleReconnect attempts to rejoin a player to their disconnected game.
func (gm *GameManager) handleReconnect(player *Player, username string) {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	// Find the *old* player struct to see if they were in a game
	oldPlayer, exists := gm.players[username]
	if !exists || oldPlayer.Game == nil {
		player.SendError("No active game found to reconnect to.")
		return
	}

	// Game found, perform the reconnect
	log.Printf("Player %s attempting to reconnect to game %s.", username, oldPlayer.Game.ID)

	// Close the new player's connection and channels, as we're replacing the old one
	go func() {
		// Drain the send channel before closing
		for len(player.Send) > 0 {
			<-player.Send
		}
		close(player.Send)
		player.Conn.Close()
	}()

	// Update the old player struct with the new connection
	oldPlayer.Game.HandleReconnect(oldPlayer, player.Conn)
}

// startGame creates and starts a new 1v1 game.
func (gm *GameManager) startGame(p1, p2 *Player) {
	gameID := uuid.New().String()
	game := NewGame(gameID, gm, p1, p2)
	gm.games[gameID] = game

	p1.Game = game
	p2.Game = game

	log.Printf("Starting game %s between %s and %s", game.ID, p1.Username, p2.Username)
	game.BroadcastState()

	// Produce analytics event
	go ProduceEvent("game_started", map[string]interface{}{
		"gameId":   game.ID,
		"player1":  p1.Username,
		"player2":  p2.Username,
		"isBot":    false,
		"gameTime": game.StartTime.Unix(),
	})
}

// startBotGame is called by the timer if no opponent joins.
func (gm *GameManager) startBotGame(player *Player) {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	// Check if the player is still the waiting player
	if gm.waitingPlayer != player {
		return // Player already got matched, do nothing
	}

	gm.waitingPlayer = nil
	gameID := uuid.New().String()
	game := NewBotGame(gameID, gm, player)
	gm.games[gameID] = game
	player.Game = game

	log.Printf("Starting bot game %s for %s", game.ID, player.Username)
	game.BroadcastState()

	// Produce analytics event
	go ProduceEvent("game_started", map[string]interface{}{
		"gameId":   game.ID,
		"player1":  player.Username,
		"player2":  "Bot",
		"isBot":    true,
		"gameTime": game.StartTime.Unix(),
	})
}
