package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

func init() {
	// Initialize random seed for room code generation
	rand.Seed(time.Now().UnixNano())
}

const (
	matchmakingTimeout = 10 * time.Second
	reconnectTimeout   = 30 * time.Second
	privateRoomTimeout = 40 * time.Second // Expiration time for private rooms
)

// GameManager manages all active games and players.
type GameManager struct {
	players       map[string]*Player // Keyed by username
	games         map[string]*Game   // Keyed by game ID
	waitingPlayer *Player
	PrivateRooms  map[string]*Player // Keyed by room code (6-char alphanumeric)
	mutex         sync.RWMutex
}

// NewGameManager creates a new game manager.
func NewGameManager() *GameManager {
	return &GameManager{
		players:      make(map[string]*Player),
		games:        make(map[string]*Game),
		PrivateRooms: make(map[string]*Player),
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

	// If player was hosting a private room, remove it
	for roomCode, roomPlayer := range gm.PrivateRooms {
		if roomPlayer == player {
			delete(gm.PrivateRooms, roomCode)
			log.Printf("Private room %s removed due to host %s disconnect.", roomCode, player.Username)
			break
		}
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

	log.Printf("Received message type: '%s' from player, username: '%s', roomCode: '%s'", msg.Type, msg.Username, msg.RoomCode) // DEBUG LOG

	switch msg.Type {
	case "join":
		log.Printf("Handling join for username: %s", msg.Username) // DEBUG LOG
		gm.handleJoin(player, msg.Username)
	case "move":
		gm.handleMove(player, msg.Column)
	case "reconnect":
		gm.handleReconnect(player, msg.Username)
	case "create_private_room":
		log.Printf("Handling create_private_room for username: %s", msg.Username) // DEBUG LOG
		gm.handleCreatePrivateRoom(player, msg.Username)
	case "join_private_room":
		log.Printf("Handling join_private_room for username: %s, room: %s", msg.Username, msg.RoomCode) // DEBUG LOG
		gm.handleJoinPrivateRoom(player, msg.Username, msg.RoomCode)
	default:
		log.Printf("Unknown message type: '%s'", msg.Type) // DEBUG LOG
		player.SendError("Unknown message type.")
	}
}

// handleJoin processes a new player's request to join a game.
func (gm *GameManager) handleJoin(player *Player, username string) {
	log.Printf("DEBUG handleJoin: entered, username=%s", username)

	if username == "" {
		log.Printf("DEBUG handleJoin: username empty")
		player.SendError("Username cannot be empty.")
		return
	}

	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	if _, exists := gm.players[username]; exists {
		// This could be a reconnect attempt, but "join" is for new games
		log.Printf("DEBUG handleJoin: username %s already exists", username)
		player.SendError("Username already taken. Try 'reconnect' if you were in a game.")
		return
	}

	log.Printf("Player %s joining.", username)
	player.Username = username
	gm.players[username] = player

	if gm.waitingPlayer == nil {
		// This is the first player, make them wait
		log.Printf("DEBUG handleJoin: %s becomes waiting player", username)
		gm.waitingPlayer = player
		player.SendMessage("waiting", nil)

		// Start the 10-second bot timer
		log.Printf("DEBUG handleJoin: starting matchmaking timer for %s", username)
		time.AfterFunc(matchmakingTimeout, func() {
			log.Printf("DEBUG: Bot timer fired for %s", username)
			gm.startBotGame(player)
		})
	} else {
		// A waiting player exists, start a game
		if gm.waitingPlayer == player {
			return // Should not happen, but safeguard
		}
		log.Printf("DEBUG handleJoin: matching %s with waiting player %s", username, gm.waitingPlayer.Username)
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

// generateRoomCode generates a random 6-character alphanumeric room code.
func (gm *GameManager) generateRoomCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 6

	for {
		code := make([]byte, codeLength)
		for i := range code {
			code[i] = charset[rand.Intn(len(charset))]
		}
		roomCode := string(code)

		// Ensure uniqueness (check if code already exists)
		gm.mutex.RLock()
		_, exists := gm.PrivateRooms[roomCode]
		gm.mutex.RUnlock()

		if !exists {
			return roomCode
		}
	}
}

// handleCreatePrivateRoom creates a new private room with a unique code.
func (gm *GameManager) handleCreatePrivateRoom(player *Player, username string) {
	log.Printf("DEBUG: Entering handleCreatePrivateRoom, username: %s", username)

	if username == "" {
		log.Printf("DEBUG: Username is empty, sending error")
		player.SendError("Username cannot be empty.")
		return
	}

	// Generate unique room code BEFORE acquiring lock to avoid deadlock
	roomCode := gm.generateRoomCode()
	log.Printf("DEBUG: Generated room code: %s", roomCode)

	gm.mutex.Lock()
	log.Printf("DEBUG: Acquired lock, checking username availability")

	// Check if username is already taken
	if _, exists := gm.players[username]; exists {
		log.Printf("DEBUG: Username %s already taken", username)
		gm.mutex.Unlock()
		player.SendError("Username already taken.")
		return
	}

	// Register the player
	player.Username = username
	gm.players[username] = player
	log.Printf("DEBUG: Player %s registered", username)

	// Store player in private rooms
	gm.PrivateRooms[roomCode] = player

	log.Printf("Player %s created private room: %s", username, roomCode)
	gm.mutex.Unlock()
	log.Printf("DEBUG: Lock released")

	// Send room code back to the client
	log.Printf("DEBUG: About to send private_room_created message with code: %s", roomCode)
	player.SendMessage("private_room_created", map[string]interface{}{
		"roomCode": roomCode,
	})
	log.Printf("DEBUG: Sent private_room_created message")

	// Start 40-second expiration timer
	time.AfterFunc(privateRoomTimeout, func() {
		gm.mutex.Lock()
		defer gm.mutex.Unlock()

		// Check if room still exists (not joined)
		if roomPlayer, exists := gm.PrivateRooms[roomCode]; exists && roomPlayer == player {
			// Room expired, clean up
			delete(gm.PrivateRooms, roomCode)
			log.Printf("Private room %s expired for player %s", roomCode, username)

			// Notify the player
			player.SendMessage("private_room_expired", map[string]interface{}{
				"message": "No one joined your room. It has expired.",
			})
		}
	})
}

// handleJoinPrivateRoom allows a player to join an existing private room.
func (gm *GameManager) handleJoinPrivateRoom(player *Player, username, roomCode string) {
	if username == "" {
		player.SendError("Username cannot be empty.")
		return
	}

	if roomCode == "" {
		player.SendError("Room code cannot be empty.")
		return
	}

	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	// Check if username is already taken
	if _, exists := gm.players[username]; exists {
		player.SendError("Username already taken.")
		return
	}

	// Check if room exists
	roomHost, exists := gm.PrivateRooms[roomCode]
	if !exists {
		player.SendError("Invalid or expired room code.")
		return
	}

	// Prevent player from joining their own room
	if roomHost.Username == username {
		player.SendError("You cannot join your own room.")
		return
	}

	// Register the joining player
	player.Username = username
	gm.players[username] = player

	// Remove room from private rooms (it's now matched)
	delete(gm.PrivateRooms, roomCode)

	log.Printf("Player %s joined private room %s (host: %s)", username, roomCode, roomHost.Username)

	// Start the game
	gm.startGame(roomHost, player)
}
