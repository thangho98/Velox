# Phase 05: Testing & Polish
Status: ⬜ Pending
Dependencies: Phase 01-04

## Objective
Test end-to-end flow, edge cases, cleanup logic, và polish UX.

## Implementation Steps

### End-to-End Tests
1. [ ] Test full flow: enable → estimate → encode → play pre-transcode → disable → cleanup
2. [ ] Test fallback: delete pre-transcode file → player falls back to realtime transcode
3. [ ] Test concurrent: 2 users play same pre-transcoded file simultaneously

### Edge Cases
4. [ ] Source file deleted while encoding → cancel job, cleanup partial output
5. [ ] Disk full during encoding → fail gracefully, notify admin, pause queue
6. [ ] NAS restart during encoding → resume queue on next startup
7. [ ] Source file replaced (different fingerprint) → re-encode, delete old pre-transcode

### Performance
8. [ ] Verify VAAPI encoding works for pre-transcode (same as realtime)
9. [ ] Benchmark: encode speed vs realtime (should be similar or faster)
10. [ ] Memory usage: pre-transcode job should not exceed ~200MB

### Cleanup & Maintenance
11. [ ] Scheduler task: verify pre-transcode files still valid (source exists, file not corrupted)
12. [ ] Admin API: `DELETE /api/admin/pretranscode/files/{media_id}` — delete specific file
13. [ ] Settings: "Delete all pre-transcode files" button with confirmation

### Notifications
14. [ ] WebSocket notification when encoding batch completes
15. [ ] Notification when disk space drops below threshold during encoding

## Files to Create/Modify
- Various test files
- `backend/internal/service/pretranscode.go` — edge case handling
- `backend/internal/handler/pretranscode.go` — cleanup endpoints

## Test Criteria
- [ ] All edge cases handled gracefully
- [ ] No orphan files on disk
- [ ] No orphan DB rows
- [ ] Notifications work
- [ ] Restart recovery works
