package main

import (
	"log"
	"math/rand"
	"time"
)

// Bot represents the AI opponent.
type Bot struct{}

// NewBot creates a new bot.
func NewBot() *Bot {
	return &Bot{}
}

// MakeMove triggers the bot to find and make a move.
func (b *Bot) MakeMove(g *Game) {
	// Add a small delay to make it feel more human
	time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)

	g.mutex.RLock() // Use RLock for reading the board state
	// Create a copy of the board for the bot to analyze
	boardCopy := g.Board
	g.mutex.RUnlock()

	// Find the best move
	col := b.findBestMove(boardCopy, Player2, Player1)

	// Make the move by calling the game's handler
	// We pass 'nil' as the player to signify it's a bot move
	g.HandleMove(nil, col)
}

// findBestMove is the core bot logic.
func (b *Bot) findBestMove(board [Rows][Cols]int, botPlayer, humanPlayer int) int {

	// 1. Check for immediate winning moves for the bot
	for c := 0; c < Cols; c++ {
		if !isValidMove(board, c) {
			continue
		}
		r := getNextOpenRow(board, c)
		board[r][c] = botPlayer // Try move
		if checkWin(board, r, c, botPlayer) {
			log.Println("Bot: Found winning move at col", c)
			return c
		}
		board[r][c] = Empty // Undo move
	}

	// 2. Check for immediate winning moves for the human (and block them)
	for c := 0; c < Cols; c++ {
		if !isValidMove(board, c) {
			continue
		}
		r := getNextOpenRow(board, c)
		board[r][c] = humanPlayer // Try human move
		if checkWin(board, r, c, humanPlayer) {
			log.Println("Bot: Found blocking move at col", c)
			return c
		}
		board[r][c] = Empty // Undo move
	}

	// 3. Simple heuristic: try to play in the center
	centerCols := []int{3, 2, 4, 1, 5, 0, 6}
	for _, c := range centerCols {
		if isValidMove(board, c) {
			// Basic check: don't set up the opponent for a win
			r := getNextOpenRow(board, c)
			if r > 0 { // Don't check if we're at the very top
				board[r-1][c] = humanPlayer
				if checkWin(board, r-1, c, humanPlayer) {
					board[r-1][c] = Empty // Undo check
					continue              // This move would let the human win, skip it
				}
				board[r-1][c] = Empty // Undo check
			}
			log.Println("Bot: Playing preferred center col", c)
			return c
		}
	}

	// 4. Fallback: play any valid random move
	for {
		c := rand.Intn(Cols)
		if isValidMove(board, c) {
			log.Println("Bot: Playing random fallback col", c)
			return c
		}
	}
}

// --- Bot Utility Functions ---
// These are static helpers for the bot to analyze hypothetical boards.

func isValidMove(board [Rows][Cols]int, col int) bool {
	return board[0][col] == Empty
}

func getNextOpenRow(board [Rows][Cols]int, col int) int {
	for r := Rows - 1; r >= 0; r-- {
		if board[r][col] == Empty {
			return r
		}
	}
	return -1
}

// checkWin is a static version of the game's win check for the bot.
func checkWin(board [Rows][Cols]int, lastRow, lastCol, playerNum int) bool {
	// Check horizontal
	count := 0
	for c := 0; c < Cols; c++ {
		if board[lastRow][c] == playerNum {
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
		if board[r][lastCol] == playerNum {
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
		if board[r][c] == playerNum {
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
		if board[r][c] == playerNum {
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
