# 🎮 4 in a Row - Real-Time Multiplayer Game

<div align="center">

A real-time, backend-driven version of Connect Four built with **Go**, featuring WebSocket communication, competitive bot AI, PostgreSQL persistence, and Kafka analytics.

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)
![Docker](https://img.shields.io/badge/Docker-Required-2496ED?style=for-the-badge&logo=docker)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14-336791?style=for-the-badge&logo=postgresql)
![Redpanda](https://img.shields.io/badge/Redpanda-Kafka--Compatible-E51F24?style=for-the-badge)

</div>

---

## 🚀 Quick Feature Overview

| Feature | Description |
|---------|-------------|
| **🎯 Quick Match** | Instant 1v1 matchmaking with 10-second bot fallback |
| **👥 Private Rooms** | Create 6-character codes to play with specific friends |
| **🔗 Share Links** | One-click invite links with auto-filled room codes |
| **🤖 Smart Bot** | AI opponent with win/block/center strategy |
| **🔄 Reconnection** | 30-second window to rejoin after disconnect |
| **📊 Analytics** | Real-time Kafka event streaming for game metrics |
| **🏆 Leaderboard** | Persistent stats tracked in PostgreSQL |

Live Demo Video: https://www.loom.com/share/d6ccecbc3fc3406ca354a27f852654df

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🎮 Gameplay
✅ Real-time 1v1 multiplayer using WebSockets  
✅ **Quick Match:** Auto-matchmaking with 10s timeout  
✅ **Private Rooms:** Create & share 6-character room codes  
✅ **Play with Friends:** Shareable invite links  
✅ **Bot Opponent:** Competitive AI with strategic moves  
✅ **Reconnection:** 30-second window to rejoin games  
✅ **Dedicated Game Page:** Clean separation of lobby & gameplay

</td>
<td width="50%">

### 📊 Backend & Analytics
✅ Persistent game history with PostgreSQL  
✅ Real-time leaderboard tracking  
✅ Game analytics dashboard  
✅ Kafka event streaming for analytics  
✅ Unique username validation  
✅ Room expiration system (40-second timeout)  
✅ Automatic bot matchmaking

</td>
</tr>
</table>

---

## 🏗️ Architecture

### Backend (Go)
- **WebSocket Server:** `gorilla/websocket` for real-time bidirectional communication
- **Game Engine:** Core game logic (`game.go`) with win detection and state management
- **Player Management:** Connection handling, reconnection support (`player.go`)
- **Matchmaking System:** 
  - Quick Match with 10-second timeout
  - Private rooms with unique 6-character codes
  - Unique username validation across all modes
  - Room expiration after 40 seconds
- **Bot AI:** Strategic, defensive, and offensive decision-making (`bot.go`)
- **Database:** `lib/pq` for PostgreSQL persistence (game history, player stats)
- **Event Streaming:** `confluent-kafka-go` for real-time analytics events

### Analytics Service (Go)
- **Kafka Consumer:** Separate service (`consumer/main.go`) for event processing and logging

### Frontend
- Vanilla JavaScript with WebSocket client
- Multi-page architecture: `/` (lobby) → `/play` (game)
- Auto-reconnection on page refresh
- Native share API integration for invite links

---

## � Game Flow & Reconnection

### Lobby → Game Redirect Flow

```
1. Player joins via Quick Match or Private Room
   ├─ WebSocket connects to /ws
   └─ Sends join/create_private_room message

2. Match found (human or bot opponent)
   ├─ Server sends game_start message
   └─ Client stores game data in sessionStorage

3. Automatic redirect to /play page
   ├─ Old WebSocket closes gracefully
   └─ Player kept in server memory for 30 seconds

4. /play page loads
   ├─ Parses game data from sessionStorage
   ├─ Opens new WebSocket connection
   └─ Sends reconnect message with username

5. Server reconnects player
   ├─ Finds player in memory (not deleted during redirect)
   ├─ Swaps old connection with new connection
   └─ Sends current game state

6. Game continues normally
   └─ Real-time move synchronization
```

### Reconnection Window

- **During Game:** 30-second reconnection window if disconnected
- **During Redirect:** Player preserved in memory during page transition
- **After 30s:** Game forfeited, opponent declared winner
- **Technical:** Player.Game != nil prevents map deletion

---

## �📋 Prerequisites

<div align="center">

| Requirement | Version |
|------------|---------|
| **Go** | 1.25 or higher |
| **Docker** | Latest |
| **Docker Compose** | Latest |

</div>

---

## 🚀 Quick Start with Docker

> **This is the simplest way to run the entire stack.**

### 1. Clone the Repository
```bash
git clone <your-repo-url>
cd <your-repo-directory>
```

### 2. Project Structure
```
./
├── main.go                 # Entry point, HTTP routes, WebSocket handler
├── game.go                 # Core game logic, win detection, reconnection
├── bot.go                  # AI bot strategy (win, block, center, random)
├── player.go               # WebSocket connection management
├── game_manager.go         # Matchmaking, private rooms, quick match
├── database.go             # PostgreSQL persistence layer
├── kafka.go                # Kafka event producer
├── consumer/
│   └── main.go             # Kafka consumer service for analytics
├── static/
│   ├── index.html          # Lobby page (matchmaking, room creation)
│   └── play.html           # Game page (board, moves, reconnection)
├── go.mod                  # Go dependencies
├── go.sum
├── Dockerfile              # Dockerfile for the main app
├── Dockerfile.consumer     # Dockerfile for the consumer
├── docker-compose.yml      # Full stack setup (app, consumer, db, kafka)
└── README.md               # This file
```

### 3. Run with Docker Compose
```bash
# Build and start all services in the background
docker-compose up --build -d
```

<div align="center">

**🌐 Application:** [http://localhost:8081](http://localhost:8081)  
**🌐 Kafka UI:** [http://localhost:8080](http://localhost:8080)

</div>

#### This command starts 4 services:

| Service | Description | Port |
|---------|-------------|------|
| `app` | Main Go WebSocket server | 8081 |
| `consumer` | Analytics event consumer | - |
| `db` | PostgreSQL database | 5432 |
| `redpanda` | Kafka-compatible message broker | 9092 |
| `kafka-ui` | Web UI for Kafka/Redpanda | 8080 |

### 4. Stop the Services
```bash
docker-compose down
```

---

## 🔧 Local Development Setup (Alternative)

### 1. Install Dependencies
```bash
go mod tidy
```

### 2. Start Services (Docker)
```bash
# Start Postgres and Redpanda (Kafka-compatible broker)
docker-compose up -d db redpanda
```

### 3. Set Environment Variables
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/connect4?sslmode=disable"
export KAFKA_BROKERS="localhost:9092"
export PORT="8081"
```

### 4. Run the Main App
```bash
go run .
```

### 5. Run the Consumer (in a new terminal)
```bash
go run ./consumer/
```

---

## 🎮 How to Play

<div align="center">

### 🎯 Three Ways to Play

</div>

#### 1️⃣ **Quick Match** (Auto-Matchmaking)
| Step | Action | Description |
|------|--------|-------------|
| **1** | **Open the Game** | Navigate to [http://localhost:8081](http://localhost:8081) |
| **2** | **Enter Username** | Type your username and click "Quick Match" |
| **3** | **Wait for Opponent** | Another player joins within 10 seconds |
| **4** | **Auto-Bot Fallback** | If no player joins, bot automatically enters |
| **5** | **Game Starts** | Redirects to `/play` page with live game |

#### 2️⃣ **Private Room** (Play with Friends)
| Step | Action | Description |
|------|--------|-------------|
| **1** | **Create Room** | Enter username and click "👥 Play with Friend" |
| **2** | **Get Room Code** | Receive unique 6-character code (e.g., `ABC123`) |
| **3** | **Share Link** | Click "📤 Share Invite Link" to copy shareable URL |
| **4** | **Friend Joins** | Friend pastes link or enters code manually |
| **5** | **Game Starts** | Both players redirect to `/play` page |

> **⏱️ Room Timeout:** Private rooms expire after 40 seconds if no one joins

#### 3️⃣ **Join Private Room** (Using Code/Link)
| Step | Action | Description |
|------|--------|-------------|
| **1** | **Receive Invite** | Get room code or link from friend |
| **2** | **Enter Code** | Paste code in "Enter 6-character room code" field |
| **3** | **Click Join** | Click "🔗 Join Private Room" |
| **4** | **Auto-Start** | Instantly joins and starts game |

---

### 🎯 Gameplay Instructions

| Step | Action | Description |
|------|--------|-------------|
| **1** | **Make Moves** | Click on any column to drop your disc |
| **2** | **Win Condition** | Connect 4 discs horizontally, vertically, or diagonally |
| **3** | **Reconnection** | If disconnected, refresh within 30 seconds to resume |

---

## 🔗 Share Link Feature

### How Invite Links Work

When you create a private room, you can share an invite link that automatically fills in the room code:

```
Example: http://localhost:8081/?room=ABC123
```

**Benefits:**
- ✅ No manual code entry required
- ✅ Direct join experience for friends
- ✅ Works across browsers and devices
- ✅ Supports native share dialog on mobile

**Implementation:**
```javascript
// Auto-fill room code from URL parameter
const urlParams = new URLSearchParams(window.location.search);
const roomCode = urlParams.get('room');
if (roomCode) {
    document.getElementById('roomCodeInput').value = roomCode.toUpperCase();
}
```

---

## 🤖 Bot AI Strategy

The bot prioritizes moves in this order:
```
┌─────────────────────────────────────────┐
│  1. WIN                                 │
│     ↓  Check for immediate winning moves│
├─────────────────────────────────────────┤
│  2. BLOCK                               │
│     ↓  Block opponent's winning moves   │
├─────────────────────────────────────────┤
│  3. CENTER                              │
│     ↓  Prefer center columns (3,2,4)    │
├─────────────────────────────────────────┤
│  4. RANDOM                              │
│     ↓  Play any valid move as fallback  │
└─────────────────────────────────────────┘
```

---

## � WebSocket Messages

### Client → Server Messages

| Message Type | Description | Payload |
|--------------|-------------|---------|
| `join` | Join quick match queue | `{"type":"join","username":"alice"}` |
| `create_private_room` | Create a private room | `{"type":"create_private_room","username":"alice"}` |
| `join_private_room` | Join existing private room | `{"type":"join_private_room","username":"bob","roomCode":"ABC123"}` |
| `move` | Make a game move | `{"type":"move","column":3}` |
| `reconnect` | Reconnect to active game | `{"type":"reconnect","username":"alice"}` |

### Server → Client Messages

| Message Type | Description | Data |
|--------------|-------------|------|
| `waiting` | Waiting for opponent | `null` |
| `game_start` | Game starting, redirect to /play | `{...gameState}` |
| `game_update` | Board state update | `{...gameState}` |
| `private_room_created` | Private room created successfully | `{"roomCode":"ABC123"}` |
| `private_room_expired` | Room expired (40s timeout) | `{"message":"..."}` |
| `reconnected` | Successfully reconnected | `{...gameState}` |
| `error` | Error message | `{"message":"Username taken"}` |

---

## �📊 API Endpoints

<div align="center">

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/leaderboard` | Get top 10 players |
| `GET` | `/api/analytics` | Get real-time game statistics |

</div>

---

## 📈 Kafka Events

The application emits events to the **`game-events`** topic:

### 📤 Event Types

#### **game_started**
```json
{
  "type": "game_started",
  "data": {
    "gameId": "uuid",
    "player1": "alice",
    "player2": "bob",
    "isBot": false
  },
  "timestamp": 1234567890
}
```

#### **move_made**
```json
{
  "type": "move_made",
  "data": {
    "gameId": "uuid",
    "player": "alice",
    "column": 3,
    "row": 5
  },
  "timestamp": 1234567890
}
```

#### **game_ended**
```json
{
  "type": "game_ended",
  "data": {
    "gameId": "uuid",
    "winner": "alice",
    "duration": 123.45,
    "isBot": false
  },
  "timestamp": 1234567890
}
```

---

<div align="center">

## 🎯 Game Rules

| Rule | Description |
|------|-------------|
| **Grid** | 7 columns × 6 rows |
| **Players** | 2 (Player vs Player or Player vs Bot) |
| **Discs** | 🔴 Red (Player 1) / 🟡 Yellow (Player 2) |
| **Objective** | Connect 4 discs in a row |
| **Win Conditions** | Horizontal, Vertical, or Diagonal |
| **Draw** | Board fills with no winner |

---

## � Technical Highlights

### 🔐 Unique Username Validation
- Global username registry across all game modes
- Real-time availability checking
- Prevents duplicate usernames in concurrent matches

### 🏠 Private Room System
- **6-Character Codes:** Alphanumeric, collision-resistant generation
- **40-Second Expiration:** Auto-cleanup if no one joins
- **Shareable Links:** Query parameter integration (`?room=ABC123`)
- **Host Protection:** Rooms auto-delete if host disconnects

### 🔄 Smart Reconnection
- **Lobby → Game Redirect:** Player stays in memory during page navigation
- **30-Second Window:** Grace period for accidental disconnects
- **Connection Swapping:** Old WebSocket replaced with new one seamlessly
- **State Preservation:** Game continues from exact position

### 🎮 Matchmaking Intelligence
- **Queue System:** FIFO for quick match
- **Bot Fallback:** Automatic after 10-second timeout
- **Concurrent Games:** Multiple matches running simultaneously
- **No Duplicate Lobbies:** Each player can only be in one queue

---

## �🛠️ Tech Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Backend** | Go 1.25 | High-performance WebSocket server |
| **WebSocket** | gorilla/websocket | Real-time bidirectional communication |
| **Database** | PostgreSQL 14 | Persistent game history & player stats |
| **Message Broker** | Redpanda (Kafka-compatible) | Event streaming for analytics |
| **Kafka Client** | confluent-kafka-go | Producer/Consumer implementation |
| **Frontend** | Vanilla JavaScript | Lightweight, no-framework approach |
| **Admin UI** | Kafka UI | Visual Kafka/Redpanda monitoring |
| **Containerization** | Docker & Docker Compose | Full-stack orchestration |

---

## 📝 License

MIT License - feel free to use this project for learning or your portfolio!

---

## 🤝 Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

---

## 📧 Support

If you encounter any issues or have questions, please open an issue on GitHub.

---

<div align="center">

**Built with ❤️ using Go**

⭐ Star this repo if you found it helpful!

</div>

</div>
