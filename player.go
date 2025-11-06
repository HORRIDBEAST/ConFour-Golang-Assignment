package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// Message is a struct for WebSocket messages
type Message struct {
	Type     string          `json:"type"`
	Username string          `json:"username,omitempty"`
	Column   int             `json:"column,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

// Player represents a single connected user.
type Player struct {
	ID       string
	Username string
	Conn     *websocket.Conn
	Game     *Game
	Manager  *GameManager
	Send     chan []byte
	mutex    sync.Mutex
}

// NewPlayer creates a new player instance.
func NewPlayer(conn *websocket.Conn, manager *GameManager) *Player {
	return &Player{
		Conn:    conn,
		Manager: manager,
		Send:    make(chan []byte, 256),
	}
}

// ReadMessages handles reading incoming messages from the player's WebSocket.
func (p *Player) ReadMessages() {
	defer func() {
		p.Manager.UnregisterPlayer(p)
		p.Conn.Close()
	}()

	p.Conn.SetReadDeadline(time.Now().Add(pongWait))
	p.Conn.SetPongHandler(func(string) error { p.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := p.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Player read error: %v", err)
			}
			break
		}
		p.Manager.HandleMessage(p, message)
	}
}

// WriteMessages handles writing outgoing messages to the player's WebSocket.
func (p *Player) WriteMessages() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-p.Send:
			p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The manager closed the channel.
				p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := p.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Player write error: %v", err)
				return
			}
		case <-ticker.C:
			p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage sends a marshaled JSON message to the player.
func (p *Player) SendMessage(msgType string, data interface{}) {
	payload, err := json.Marshal(map[string]interface{}{"type": msgType, "data": data})
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}
	p.Send <- payload
}

// SendError sends an error message to the player.
func (p *Player) SendError(message string) {
	p.SendMessage("error", map[string]string{"message": message})
}
