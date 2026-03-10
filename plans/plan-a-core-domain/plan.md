# Plan A: Core Domain & Ingestion
Status: ⬜ Pending
Priority: 🔴 Critical (làm đầu tiên)

## Mục tiêu
Xây nền móng data model + ingestion pipeline đủ vững cho toàn bộ hệ thống.
Sau plan này: scan folder → identify file → fetch metadata → sẵn sàng play.

## Phases

| Phase | Name | Tasks | Status |
|-------|------|-------|--------|
| 01 | Migration System | 5 tasks | ⬜ |
| 02 | Core Data Model | 9 tasks | ⬜ |
| 03 | Scan Pipeline & File Identity | 8 tasks | ⬜ |
| 04 | TMDb Integration | 7 tasks | ⬜ |
| 05 | Subtitle Discovery | 5 tasks | ⬜ |

## Key Design Decisions
- Migration versioning thay vì CREATE IF NOT EXISTS
- File fingerprint (size + path) để detect rename/move
- Scan job state machine: queued → scanning → done/failed
- Media item tách biệt khỏi physical file (1 movie có thể có nhiều versions)
