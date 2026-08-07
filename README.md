# Neon Space Survivor

**Neon Space Survivor** is a high-octane, arcade-style bullet hell survival game built with the Ebiten Go game engine. Navigate your ship through endless waves of increasingly difficult cosmic horrors, upgrade your firepower, and face the ultimate void boss.

## 🚀 Features

- **Dynamic Difficulty**: Enemies scale in speed, health, and size as the mission progresses.
- **Procedural Soundscape**: Synthetic sound effects generated in real-time, paired with a dynamic MP3 music system that crossfades and evolves based on game intensity.
- **Minute-by-Minute Events**:
  - **Minute 8**: Combat Overdrive—your fire rate reaches its peak and spawning intensifies.
  - **Minute 10**: The Void Wave—a massive screen-shaking crisis where all existing enemies explode to make room for purple elites.
  - **Minute 11**: The Final Boss—the ultimate test of skill and survival.
- **Procedural Vector Graphics**: High-performance neon visuals that maintain sharpness at any resolution.
- **Persistence**: Track your progress with a highscore system that saves kills, victory status, and session time.

## 🎮 Controls

| Key | Action |
|-----|--------|
| **W, A, S, D** | Move Ship |
| **Mouse / Auto-Aim** | Target Enemies |
| **L** | Toggle Auto-Aim |
| **I** | Toggle UI Overlay |
| **Enter** | Start Game / Submit Score |
| **R** | Restart Game |
| **Backspace** | Delete character in name input |
| **Escape** | Exit Game |

## 🛠️ Technical Overview

### Procedural Audio
The game features a custom `audio.go` engine that:
- Synthesizes laser, explosion, and hit sounds mathematically at startup.
- Manages complex music states (Splash, Mission Start, Crisis, Ending).
- Implements seamless 8-second crossfades between track variations.

### Difficulty Scaling
The game uses a linear-to-exponential scaling curve:
- **Fire Rate**: Decreases cooldown from 12 frames to 3 frames over time, with a Minute 8 "Overdrive" setting it to 2.
- **Damage**: Elite and Boss enemies deal percentage-based damage (50% and 75% of max HP respectively), ensuring late-game stakes remain high regardless of player health.

## 📦 Requirements

- **Go** (1.21+)
- **Ebiten v2** dependencies (ALSA/Mesa on Linux)

## 🏗️ Building and Running

1. Clone the repository.
2. Run `go mod tidy` to install dependencies.
3. Start the game:
   ```bash
   go run .
   ```

## 🌐 Centralized Online Highscore Leaderboard

The game supports a centralized multi-user online leaderboard allowing players on different installations to share scores.

### Running the Leaderboard Server
Start the central HTTP REST API server:
```bash
go run ./server
```
By default, the server runs on port `8080` and stores scores in `server_scores.db`. You can customize port and database path via environment variables:
```bash
PORT=8080 DB_PATH=my_scores.db go run ./server
```

### Running the Game Client
The game automatically attempts to connect to `http://localhost:8080`. To connect to a remote hosted server:
```bash
SCORE_SERVER_URL=https://your-leaderboard-server.com go run .
```
If the server is unreachable, the game automatically falls back to local offline caching without interrupting gameplay.

---
*Created with love by Gwen-Pigat and the Antigravity AI.*