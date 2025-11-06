# Testing Guide for 4 in a Row

## Quick Testing Checklist

### 1. Basic Functionality ✅

#### Single Player vs Bot
1. Open `http://localhost:8080`
2. Enter username "TestPlayer1"
3. Click "Join Game"
4. Wait 10 seconds - Bot should join automatically
5. Make moves by clicking columns
6. Bot should respond intelligently within 1 second

**Expected Behavior:**
- Game starts after 10 seconds
- Bot makes strategic moves (blocks, tries to win)
- Game ends when someone connects 4 discs
- Winner is displayed
- Leaderboard updates

#### Two Player Multiplayer
1. Open two browser windows/tabs
2. Window 1: Enter username "Alice", join
3. Window 2: Enter username "Bob", join
4. Game should start immediately in both windows
5. Make moves alternately
6. Both windows should update in real-time

**Expected Behavior:**
- Game starts immediately when second player joins
- Both players see moves in real-time
- Turn indicator shows current player
- Can't make move when it's opponent's turn

### 2. Reconnection Testing ✅

#### Disconnect and Reconnect
1. Start a game (vs player or bot)
2. Close browser tab (simulate disconnect)
3. Within 30 seconds, reopen and enter same username
4. Game should resume

**Expected Behavior:**
- Reconnection successful within 30 seconds
- Game state is restored
- Can continue playing from where you left off

#### Forfeit Test
1. Start a game
2. Close browser
3. Wait more than 30 seconds
4. Check opponent's screen

**Expected Behavior:**
- After 30 seconds, opponent wins by forfeit
- Game is marked as finished

### 3. Game Logic Testing ✅

#### Horizontal Win
```
Make moves in sequence to test horizontal win:
Player 1: Column 0, 1, 2, 3
Result: Player 1 wins (bottom row)
```

#### Vertical Win
```
Make moves in sequence:
Player 1: Column 0, 0, 0, 0 (same column 4 times)
Result: Player 1 wins (vertical)
```

#### Diagonal Win
```
Setup for diagonal:
P1: Col 0, Col 1, Col 2, Col 3
P2: Col 1, Col 2, Col 3, Col 4
P1: Col 2, Col 3, Col 4
P2: Col 3, Col 4, Col 5
P1: Col 3
Result: Diagonal win
```

#### Draw Game
```
Fill entire board without 4 in a row
Result: Game ends in draw
```

### 4. Bot AI Testing ✅

#### Bot Blocks Player Win
1. Create a situation where you have 3 in a row
2. Bot should block the 4th position

Example:
```
Your discs: [0,0], [0,1], [0,2]
Bot should: Play [0,3] to block
```

#### Bot Takes Winning Move
1. Let bot create 3 in a row
2. Bot should win on next move

**Test Scenarios:**
```
Scenario 1: Bot has [5,0], [5,1], [5,2]
Expected: Bot plays [5,3] and wins

Scenario 2: Bot has [5,3], [4,3], [3,3]
Expected: Bot plays [2,3] and wins (vertical)
```

### 5. Database & Analytics Testing ✅

#### Leaderboard
1. Play multiple games with different outcomes
2. Check leaderboard updates after each game
3. Verify win counts are accurate

**Expected:**
- Leaderboard shows top 10 players
- Win counts are correct
- Sorted by wins (highest first)

#### Analytics
1. Play several games
2. Check analytics dashboard
3. Verify metrics:

**Expected Metrics:**
- Total games count increases
- Average duration is calculated
- Bot vs Player game ratio is correct
- Games today count is accurate
- Most frequent winner is shown

### 6. Kafka Events Testing ✅

#### View Kafka Events
```bash
# In terminal, view consumer logs
docker logs -f connect4-kafka-consumer-1

# Or if running locally
cd consumer
go run main.go
```

**Expected Events:**
1. `game_started` - When game begins
2. `move_made` - For each move
3. `game_ended` - When game finishes

**Event Format:**
```json
{
  "type": "game_started",
  "data": {
    "gameId": "uuid",
    "player1": "Alice",
    "player2": "Bob",
    "isBot": false
  },
  "timestamp": 1234567890
}
```

### 7. Performance Testing ✅

#### Multiple Concurrent Games
1. Open 4-6 browser windows
2. Create 2-3 simultaneous games
3. Make moves in different games
4. Check responsiveness

**Expected:**
- All games run smoothly
- No lag or delays
- No interference between games

#### Stress Test
```bash
# Use a tool like Artillery or k6
# Example with curl
for i in {1..10}; do
  curl -X GET http://localhost:8080/api/leaderboard &
done
```

**Expected:**
- Server handles multiple requests
- No crashes or errors
- Response times < 1 second

### 8. Edge Cases Testing ✅

#### Invalid Moves
1. Try clicking same column when full
2. Try moving when it's not your turn
3. Try moving after game ends

**Expected:**
- Error message displayed
- Move is rejected
- Game state unchanged

#### Duplicate Usernames
1. Player 1 joins with "TestUser"
2. Player 2 tries to join with "TestUser"

**Expected:**
- Second player can join (handled by unique ID)
- Both games work independently

#### Network Issues
1. Disconnect internet during game
2. Reconnect within 30 seconds

**Expected:**
- Reconnection works
- Game resumes
- No data loss

### 9. UI/UX Testing ✅

#### Responsive Design
1. Test on desktop (1920x1080)
2. Test on tablet (768x1024)
3. Test on mobile (375x667)

**Expected:**
- Board is visible and playable
- Buttons are clickable
- Text is readable

#### Visual Feedback
1. Hover over cells
2. Check active player indicator
3. Watch disc drop animation

**Expected:**
- Cells highlight on hover
- Active player has green border
- Smooth animations

### 10. Database Persistence Testing ✅

#### Game History
```bash
# Connect to PostgreSQL
docker exec -it postgres psql -U postgres -d connect4

# Check games table
SELECT * FROM games ORDER BY created_at DESC LIMIT 5;

# Check players table
SELECT * FROM players ORDER BY games_won DESC;
```

**Expected:**
- All completed games are stored
- Player stats are accurate
- Board state is saved as JSON

## Automated Testing Commands

```bash
# Run all tests
go test ./... -v

# Test with coverage
go test ./... -cover

# Benchmark tests
go test ./... -bench=.

# Race condition detection
go test ./... -race
```

## Common Issues & Solutions

### Issue: WebSocket connection fails
**Solution:** 
- Check if server is running
- Verify port 8080 is not in use
- Check browser console for errors

### Issue: Bot doesn't make moves
**Solution:**
- Check game manager logs
- Verify bot logic in bot.go
- Check if game status is "playing"

### Issue: Database connection fails
**Solution:**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check connection string
echo $DATABASE_URL

# Test connection
psql $DATABASE_URL
```

### Issue: Kafka events not received
**Solution:**
```bash
# Check Kafka broker
docker logs kafka

# Check if topic exists
docker exec kafka kafka-topics --list --bootstrap-server localhost:9092

# Check consumer
docker logs connect4-kafka-consumer-1
```

## Test Results Checklist

Mark each test as you complete it:

- [ ] Single player vs bot works
- [ ] Two player multiplayer works
- [ ] Reconnection within 30 seconds works
- [ ] Forfeit after 30 seconds works
- [ ] Horizontal win detection works
- [ ] Vertical win detection works
- [ ] Diagonal win detection works
- [ ] Draw detection works
- [ ] Bot blocks player wins
- [ ] Bot takes winning moves
- [ ] Leaderboard updates correctly
- [ ] Analytics are accurate
- [ ] Kafka events are sent
- [ ] Multiple concurrent games work
- [ ] Invalid moves are rejected
- [ ] UI is responsive
- [ ] Database persistence works

## Report Issues

If you find any bugs, please include:
1. Steps to reproduce
2. Expected behavior
3. Actual behavior
4. Browser/environment details
5. Screenshots or logs

---

Happy Testing! 🎮