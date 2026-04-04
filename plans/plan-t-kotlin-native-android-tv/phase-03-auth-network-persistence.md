# Phase 03: Auth + Network + Persistence
Status: ⬜ Pending
Dependencies: Phase 02

## Objective
Port auth/session/base networking from the current TS layer into Kotlin-native repositories, including token refresh, secure persistence, and app bootstrap auth state.

## Key Mapping from Current Client
- TS API client → Retrofit/OkHttp service layer
- token refresh queue → OkHttp authenticator/interceptor strategy
- SecureStore/Zustand persistence → DataStore
- platform device header → native device-info provider

## Verified Contract Sources
- `packages/shared/hooks/auth.ts`
- `packages/shared/api/client.ts`
- `packages/shared/types/auth.ts`
- `packages/shared/types/common.ts`

## Implementation Steps

### 1. Port DTOs
- [ ] login request/response
- [ ] refresh request/response
- [ ] `UserInfo`, `User`, `Session`, `UserPreferences`
- [ ] common API response envelope/error model

### 2. Build Networking Core
- [ ] `ApiEnvelope` / error parsing
- [ ] Retrofit services for auth/profile
- [ ] OkHttp logging + auth interceptors
- [ ] device-name header provider

### 3. Build Persistence Core
- [ ] `AuthPreferences` DataStore
- [ ] token expiry persistence
- [ ] player/settings preference placeholders

### 4. Build Session Flow
- [ ] `AuthRepository`
- [ ] `SessionRepository`
- [ ] startup bootstrap use case
- [ ] session-expired handler

### 5. Build UI Flow
- [ ] login screen
- [ ] splash/bootstrap gate
- [ ] logout path
- [ ] profile bootstrap after login

### 6. Build Tests
- [ ] refresh success
- [ ] refresh failure
- [ ] session restore
- [ ] logout clears state

## Tasks
1. [ ] Port auth/request/response models from `packages/shared/types`.
2. [ ] Create Retrofit interfaces for:
   - login
   - refresh
   - logout
   - current user/profile bootstrap
3. [ ] Implement `AuthTokenStore` in DataStore:
   - access token
   - refresh token
   - expiry timestamp
4. [ ] Implement OkHttp interceptor for:
   - bearer header
   - device name header
5. [ ] Implement token refresh authenticator with single-refresh protection.
6. [ ] Build `AuthRepository` and `SessionRepository`.
7. [ ] Build `AuthViewModel` and login screen UI state.
8. [ ] Implement app startup flow:
   - splash/bootstrap
   - existing session restore
   - redirect to auth or main graph
9. [ ] Add logout flow and session-expired handling.
10. [ ] Write unit tests for token refresh, session restore, and unauthorized behavior.

## Files / Modules to Create
- [ ] `core:model/.../auth/*.kt`
- [ ] `core:network/.../AuthApi.kt`
- [ ] `core:network/.../ProfileApi.kt`
- [ ] `core:network/.../AuthRepositoryImpl.kt`
- [ ] `core:datastore/.../AuthDataStore.kt`
- [ ] `feature:auth/.../LoginViewModel.kt`
- [ ] `feature:auth/.../LoginScreen.kt`
- [ ] `app/.../BootstrapViewModel.kt`

## Test Criteria
- [ ] Wrong credentials show typed UI error.
- [ ] Valid login persists session.
- [ ] App restart restores session without manual login.
- [ ] Expired access token refreshes once and retries the original request.
- [ ] Invalid refresh token forces clean logout.

## Done When
- [ ] User can launch app, log in, kill app, reopen, and remain authenticated.

## Verification
- [ ] Fresh login works against backend.
- [ ] App relaunch restores session from persisted tokens.
- [ ] Expired token refreshes transparently.
- [ ] Refresh failure returns user to auth screen cleanly.

## Exit Criteria
- Auth foundation is stable enough that feature screens no longer care about token mechanics.
