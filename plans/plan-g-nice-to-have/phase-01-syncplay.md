# Phase 01: SyncPlay (Watch Party)
Status: ⬜ Pending

## Tasks
### 1. WebSocket Server
- [ ] `/ws` endpoint, auth via token query param

### 2. Room Management
- [ ] Create/join/leave rooms by room code, max 10 users

### 3. Playback Sync Protocol
- [ ] Messages: play, pause, seek, buffering, ready
- [ ] Host controls, play only when all clients ready

### 4. Chat
- [ ] Simple text chat broadcast to room members

### 5. Latency Compensation
- [ ] Ping/pong per client, adjust sync offset, ±2s tolerance
