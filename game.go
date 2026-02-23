package main

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	Rows = 6
	Cols = 7
)

const (
	Empty   = 0
	Player1 = 1
	Player2 = 2
)

// Game holds the state of a single 4-in-a-Row game.
type Game struct {
	ID            string `json:"id"`
	Board         [Rows][Cols]int
	Player1       *Player
	Player2       *Player // Nil if bot game
	Bot           *Bot    // Nil if player game
	IsBot         bool    `json:"isBot"`
	CurrentPlayer int     `json:"currentPlayer"`
	Status        string  `json:"status"` // "playing", "finished"
	Winner        int     `json:"winner"` // 0 for draw
	StartTime     time.Time
	EndTime       time.Time
	manager       *GameManager
	mutex         sync.RWMutex
}

// GameState is a serializable representation of the game.
type GameState struct {
	ID            string          `json:"id"`
	Board         [Rows][Cols]int `json:"board"`
	Player1       string          `json:"player1"`
	Player2       string          `json:"player2"`
	IsBot         bool            `json:"isBot"`
	CurrentPlayer int             `json:"currentPlayer"`
	Status        string          `json:"status"`
	Winner        int             `json:"winner"`
}

// NewGame creates a 1v1 game.
func NewGame(id string, manager *GameManager, p1, p2 *Player) *Game {
	return &Game{
		ID:            id,
		Player1:       p1,
		Player2:       p2,
		IsBot:         false,
		CurrentPlayer: Player1,
		Status:        "playing",
		StartTime:     time.Now(),
		manager:       manager,
	}
}

// NewBotGame creates a player vs bot game.
func NewBotGame(id string, manager *GameManager, p1 *Player) *Game {
	return &Game{
		ID:            id,
		Player1:       p1,
		Bot:           NewBot(),
		IsBot:         true,
		CurrentPlayer: Player1,
		Status:        "playing",
		StartTime:     time.Now(),
		manager:       manager,
	}
}

// getPlayerName is a helper to get the opponent's name (human or bot).
func (g *Game) getPlayerName(p *Player) string {
	if p != nil {
		return p.Username
	}
	if g.IsBot {
		return "Bot"
	}
	return "Unknown"
}

// CreateState builds a serializable game state.
func (g *Game) CreateState() *GameState {
	return &GameState{
		ID:            g.ID,
		Board:         g.Board,
		Player1:       g.Player1.Username,
		Player2:       g.getPlayerName(g.Player2),
		IsBot:         g.IsBot,
		CurrentPlayer: g.CurrentPlayer,
		Status:        g.Status,
		Winner:        g.Winner,
	}
}

// BroadcastState sends the current game state to all players in the game.
func (g *Game) BroadcastState() {
	state := g.CreateState()
	g.Player1.SendMessage("game_update", state)

	if !g.IsBot && g.Player2 != nil {
		g.Player2.SendMessage("game_update", state)
	}
}

// HandleMove processes a move from a player or bot.
func (g *Game) HandleMove(player *Player, col int) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.Status != "playing" {
		return // Game is already over
	}

	// Identify which player is making the move
	var playerNum int
	if player == g.Player1 {
		playerNum = Player1
	} else if !g.IsBot && player == g.Player2 {
		playerNum = Player2
	} else if player == nil && g.IsBot { // `player == nil` signifies a bot move
		playerNum = Player2
	} else {
		log.Printf("Error: Move from unassociated player %s in game %s", player.Username, g.ID)
		return
	}

	if playerNum != g.CurrentPlayer {
		if player != nil {
			player.SendError("It's not your turn.")
		}
		return
	}

	// Attempt to make the move
	row, err := g.makeMove(col, playerNum)
	if err != nil {
		if player != nil {
			player.SendError(err.Error())
		}
		return
	}

	// Produce analytics event for the move
	go ProduceEvent("move_made", map[string]interface{}{
		"gameId":   g.ID,
		"player":   g.getPlayerName(player),
		"column":   col,
		"row":      row,
		"gameTime": time.Now().Unix(),
	})

	// Check for win
	if g.checkWin(row, col, playerNum) {
		g.endGame(playerNum)
		g.BroadcastState()
		return
	}

	// Check for draw
	if g.checkDraw() {
		g.endGame(Empty) // 0 for draw
		g.BroadcastState()
		return
	}

	// Switch players
	g.CurrentPlayer = 3 - g.CurrentPlayer // Switches between 1 and 2

	// Broadcast the updated state
	g.BroadcastState()

	// If it's now the bot's turn, trigger its move
	if g.IsBot && g.CurrentPlayer == Player2 {
		go g.Bot.MakeMove(g)
	}
}

// makeMove places a disc on the board.
func (g *Game) makeMove(col int, playerNum int) (int, error) {
	if col < 0 || col >= Cols {
		return -1, errors.New("Invalid column")
	}

	// Find the lowest empty row in this column
	for r := Rows - 1; r >= 0; r-- {
		if g.Board[r][col] == Empty {
			g.Board[r][col] = playerNum
			return r, nil
		}
	}

	return -1, errors.New("Column is full")
}

// checkWin checks if the last move resulted in a win.
func (g *Game) checkWin(lastRow, lastCol, playerNum int) bool {
	// Check horizontal
	count := 0
	for c := 0; c < Cols; c++ {
		if g.Board[lastRow][c] == playerNum {
			count++
			if count >= 4 {
				return true
			}
		} else {
			count = 0
		}
	}

	// Check vertical
	count = 0
	for r := 0; r < Rows; r++ {
		if g.Board[r][lastCol] == playerNum {
			count++
			if count >= 4 {
				return true
			}
		} else {
			count = 0
		}
	}

	// Check diagonals (top-left to bottom-right)
	count = 0
	for r, c := lastRow-min(lastRow, lastCol), lastCol-min(lastRow, lastCol); r < Rows && c < Cols; r, c = r+1, c+1 {
		if g.Board[r][c] == playerNum {
			count++
			if count >= 4 {
				return true
			}
		} else {
			count = 0
		}
	}

	// Check diagonals (bottom-left to top-right)
	count = 0
	for r, c := lastRow+min(Rows-1-lastRow, lastCol), lastCol-min(Rows-1-lastRow, lastCol); r >= 0 && c < Cols; r, c = r-1, c+1 {
		if g.Board[r][c] == playerNum {
			count++
			if count >= 4 {
				return true
			}
		} else {
			count = 0
		}
	}

	return false
}

// checkDraw checks if the board is full.
func (g *Game) checkDraw() bool {
	for c := 0; c < Cols; c++ {
		if g.Board[0][c] == Empty { // Only need to check the top row
			return false
		}
	}
	return true
}

// endGame concludes the game, saves stats, and updates players.
func (g *Game) endGame(winner int) {
	g.Status = "finished"
	g.Winner = winner
	g.EndTime = time.Now()

	// Save to database
	go SaveGame(g)

	// Update player stats
	var winnerUsername string
	if winner == Player1 {
		go UpdatePlayerStats(g.Player1.Username, true)
		winnerUsername = g.Player1.Username
		if !g.IsBot && g.Player2 != nil {
			go UpdatePlayerStats(g.Player2.Username, false)
		}
	} else if winner == Player2 {
		winnerUsername = g.getPlayerName(g.Player2)
		go UpdatePlayerStats(g.Player1.Username, false)
		if !g.IsBot && g.Player2 != nil {
			go UpdatePlayerStats(g.Player2.Username, true)
		}
	} else {
		winnerUsername = "Draw"
		go UpdatePlayerStats(g.Player1.Username, false)
		if !g.IsBot && g.Player2 != nil {
			go UpdatePlayerStats(g.Player2.Username, false)
		}
	}

	// Produce analytics event
	go ProduceEvent("game_ended", map[string]interface{}{
		"gameId":   g.ID,
		"winner":   winnerUsername,
		"duration": g.EndTime.Sub(g.StartTime).Seconds(),
		"isBot":    g.IsBot,
		"gameTime": g.EndTime.Unix(),
	})

	// Remove game from active list
	delete(g.manager.games, g.ID)
	g.Player1.Game = nil
	if !g.IsBot && g.Player2 != nil {
		g.Player2.Game = nil
	}
}

// HandleDisconnect handles a player disconnecting mid-game.
func (g *Game) HandleDisconnect(player *Player) {
	g.mutex.Lock()
	if g.Status != "playing" {
		g.mutex.Unlock()
		return // Game already ended
	}

	// Start reconnect timer
	log.Printf("Starting 30s reconnect timer for %s in game %s", player.Username, g.ID)
	g.mutex.Unlock() // Unlock to allow reconnects

	time.AfterFunc(reconnectTimeout, func() {
		g.mutex.Lock()
		defer g.mutex.Unlock()

		// Check if game is still playing (i.e., player did NOT reconnect)
		if g.Status != "playing" {
			return
		}

		log.Printf("Reconnect timer expired for %s. Forfeiting game %s.", player.Username, g.ID)

		var winner int
		if player == g.Player1 {
			winner = Player2
		} else {
			winner = Player1
		}

		g.endGame(winner)
		g.BroadcastState()
	})
}

// HandleReconnect handles a player re-joining a game.
func (g *Game) HandleReconnect(oldPlayer *Player, newConn *websocket.Conn) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.Status != "playing" {
		oldPlayer.SendError("Game has already finished.")
		return
	}

	log.Printf("Player %s reconnected to game %s.", oldPlayer.Username, g.ID)

	// Close the old connection and channel
	oldPlayer.mutex.Lock()
	oldPlayer.Conn.Close()
	close(oldPlayer.Send)

	// Assign new connection and create new Send channel
	oldPlayer.Conn = newConn
	oldPlayer.Send = make(chan []byte, 256)
	oldPlayer.mutex.Unlock()

	// Restart the WriteMessages goroutine
	go oldPlayer.WriteMessages()

	// Re-send the game state using SendMessage (which is now safe)
	state := g.CreateState()
	oldPlayer.SendMessage("reconnected", state)
}
