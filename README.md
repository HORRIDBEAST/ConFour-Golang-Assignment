# 🎮 4 in a Row - Real-Time Multiplayer Game

<div align="center">

A real-time, backend-driven version of Connect Four built with **Go**, featuring WebSocket communication, competitive bot AI, PostgreSQL persistence, and Kafka analytics.

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)
![Docker](https://img.shields.io/badge/Docker-Required-2496ED?style=for-the-badge&logo=docker)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14-336791?style=for-the-badge&logo=postgresql)
![Kafka](https://img.shields.io/badge/Kafka-7.3.0-231F20?style=for-the-badge&logo=apache-kafka)

</div>

Live Vedio Url :- https://www.loom.com/share/d6ccecbc3fc3406ca354a27f852654df

---

## ✨ Features

<table>
<tr>
<td width="50%">

✅ Real-time 1v1 multiplayer gameplay using WebSockets  
✅ Competitive AI bot with strategic decision-making  
✅ Automatic matchmaking (10-second timeout before bot joins)  
✅ Player reconnection support (30-second window)

</td>
<td width="50%">

✅ Persistent game history with PostgreSQL  
✅ Real-time leaderboard  
✅ Game analytics dashboard  
✅ Kafka event streaming for analytics  
✅ Simple, functional frontend UI

</td>
</tr>
</table>

---

## 🏗️ Architecture

### Backend (Go)
- **WebSocket Server:** `gorilla/websocket` for real-time communication.
- **Game Engine:** Core game logic (`game.go`), player management (`player.go`), and matchmaking (`game_manager.go`).
- **Bot AI:** Strategic, non-random decision-making (`bot.go`).
- **Database:** `lib/pq` for PostgreSQL game history and player stats.
- **Event Streaming:** `confluent-kafka-go` for producing analytics events.

### Analytics Service (Go)
- **Kafka Consumer:** A separate service (`consumer/main.go`) that listens to game events for logging and processing.

### Frontend
- Vanilla JavaScript (`static/index.html`) with a WebSocket client to interact with the backend.

---

## 📋 Prerequisites

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
├── main.go                 # Entry point, HTTP routes
├── game.go                 # Core game logic
├── bot.go                  # AI bot logic
├── player.go               # Player WebSocket connection handler
├── game_manager.go         # Matchmaking & game state management
├── database.go             # PostgreSQL operations
├── kafka.go                # Kafka producer
├── consumer/
│   └── main.go             # Kafka consumer service
├── static/
│   └── index.html          # Frontend
├── go.mod                  # Go dependencies
├── go.sum
├── Dockerfile              # Dockerfile for the main app
├── Dockerfile.consumer     # Dockerfile for the consumer
├── docker-compose.yml      # Full stack setup
└── README.md               # This file
```

### 3. Run with Docker Compose
```bash
# Build and start all services in the background
docker-compose up --build -d
```

<div align="center">

**🌐 The application will be available at:** [http://localhost:8080](http://localhost:8080)

</div>

#### This command starts 5 services:

| Service | Description | Port |
|---------|-------------|------|
| `app` | The main Go server | 8080 |
| `consumer` | The Go analytics consumer | - |
| `db` | The PostgreSQL database | 5432 |
| `kafka` | The Kafka broker | 9092 |
| `zookeeper` | Kafka's dependency | 2181 |

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
# Start Postgres, Zookeeper, and Kafka
docker-compose up -d db zookeeper kafka
```

### 3. Set Environment Variables
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/connect4?sslmode=disable"
export KAFKA_BROKERS="localhost:9092"
export PORT="8080"
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

### Step-by-Step Guide

</div>

| Step | Action | Description |
|------|--------|-------------|
| **1** | **Open the Game** | Navigate to [http://localhost:8080](http://localhost:8080) |
| **2** | **Enter Username** | Type your username and click "Join Game" |
| **3** | **Wait for Opponent** | **1v1:** Another player joins within 10 seconds<br>**1vBot:** Bot joins automatically if no player |
| **4** | **Make Moves** | Click on any column to drop your disc |
| **5** | **Win Condition** | Connect 4 discs horizontally, vertically, or diagonally |

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

## 📊 API Endpoints

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

## 🛠️ Tech Stack

| Component | Technology |
|-----------|-----------|
| **Backend** | Go 1.25 |
| **WebSocket** | gorilla/websocket |
| **Database** | PostgreSQL 14 |
| **Message Queue** | Apache Kafka 7.3.0 |
| **Frontend** | Vanilla JavaScript |
| **Containerization** | Docker & Docker Compose |

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
