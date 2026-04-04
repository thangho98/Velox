# Phase 09: QA + Release + Android Cutover
Status: ⬜ Pending
Dependencies: Phase 05-08

## Objective
Stabilize the app, verify real-device behavior, release internal builds, and make a deliberate Android cutover decision without breaking the existing user path.

## Release Principle
Do not declare cutover based on emulator success or a happy-path demo. Cutover only happens after real devices, real libraries, and real testers all confirm daily-use viability.

## Implementation Steps

### 1. Build Test Matrix
- [ ] phone devices
- [ ] tablet if supported
- [ ] Android TV / Google TV
- [ ] Chromecast receiver path
- [ ] codec/container matrix from real library

### 2. Build Automated Coverage
- [ ] unit tests for repositories/use cases
- [ ] Compose UI smoke tests
- [ ] API integration tests with MockWebServer where valuable

### 3. Build Manual Smoke Scripts
- [ ] login
- [ ] browse
- [ ] search
- [ ] movie playback
- [ ] episode playback
- [ ] resume
- [ ] subtitles/audio/quality
- [ ] TV remote navigation
- [ ] cast

### 4. Beta Distribution
- [ ] signed internal APK/AAB
- [ ] install instructions
- [ ] feedback capture path

### 5. Cutover Decision
- [ ] native recommended by default for Android
- [ ] RN Android frozen or kept as fallback for a transition period
- [ ] docs updated

## Tasks
1. [ ] Create a real-device QA matrix:
   - Android phone
   - Android tablet if relevant
   - Android TV / Google TV
   - Chromecast target
2. [ ] Define smoke scenarios:
   - login
   - browse
   - detail
   - movie playback
   - episode playback
   - resume progress
   - subtitle/audio/quality change
   - cast session
3. [ ] Add unit tests for critical repositories and session flows.
4. [ ] Add Compose UI tests for top navigation flows.
5. [ ] Add at least a minimal benchmark/perf pass:
   - cold start
   - home render
   - player start latency
6. [ ] Run memory/leak checks around playback screen entry/exit.
7. [ ] Produce signed internal APK/AAB distribution flow.
8. [ ] Run internal beta with a small tester group using real libraries and devices.
9. [ ] Log production issues and fix top blockers before cutover.
10. [ ] Decide Android cutover milestone:
   - recommend native by default
   - keep RN Android as fallback for one transition window
11. [ ] Update README/docs with native Android build/run instructions.
12. [ ] Only after beta stability: decide whether to freeze RN Android development.

## Test Criteria
- [ ] no P0/P1 blockers remain in core browse/playback flows
- [ ] TV path passes real remote checklist
- [ ] cast path works on real hardware
- [ ] performance is acceptable on target devices
- [ ] known deferred features are documented, not regressions

## Done When
- [ ] Team has enough evidence to make a calm go/no-go cutover decision.

## Verification
- [ ] Beta testers can use the native app for daily playback.
- [ ] No P0/P1 crashers remain in core browse/playback flows.
- [ ] TV path passes a real remote-navigation checklist.
- [ ] Team has a documented go/no-go decision for cutover.

## Exit Criteria
- Native Android app is the recommended Android client for Velox.
