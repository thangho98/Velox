# Phase 02: Chapter Support
Status: ⬜ Pending

## Tasks
### 1. Extract Chapters from FFprobe
- [ ] Parse MKV chapters: title, start_time, end_time
- [ ] Store in `chapters` table

### 2. Chapter API & Thumbnails
- [ ] `GET /api/media/{id}/chapters`
- [ ] Generate thumbnail per chapter start point

### 3. Intro/Credits Detection
- [ ] Basic: first chapter < 3min = intro, last = credits
- [ ] Frontend: "Skip Intro" button
