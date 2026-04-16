# Phase 07: Android Integration — URL Refresh + Badge
Status: ⬜ Pending
Dependencies: Phase 05

## Objective

Enable Android app playback fshare media: handle stream URL response (direct CDN), implement URL refresh khi 403 giữa chừng, show ☁️ badge trên MediaCard.

## Context

- Android app: Kotlin + Jetpack Compose + Media3 ExoPlayer + Hilt
- Existing stream flow: [android/.../data/repository/PlaybackRepository.kt](android/app/src/main/java/com/velox/app/data/repository/PlaybackRepository.kt) calls `POST /api/stream/{id}/url` → feeds URL vào ExoPlayer
- ExoPlayer native support direct URLs (HTTP/HTTPS, MKV/MP4, MP3/AAC/AC3/DTS) — no config change
- **URL TTL uncertainty (verified từ OSS research):** fshare direct URLs minutes TTL, có thể single-use. Existing HTTP connection có thể play được full movie (single long-lived TCP stream). Nhưng seek/resume/network reconnect = new connection = may need fresh URL.
- **Strategy: reactive-only refresh.** KHÔNG proactive timer-based refresh (URL không survive giữa refreshes). Refresh on-demand khi ExoPlayer gặp 403/404 from fshare CDN.

## Implementation Steps

### 1. Update Stream URL Response Model
- [ ] Modify `android/.../data/api/dto/StreamURLResponse.kt`:
  ```kotlin
  data class StreamURLResponse(
      val url: String,
      val sourceType: String, // "local" | "fshare"
      val apiKey: String? = null,
      val expiresAt: String, // ISO 8601
      val expiresIn: Int,
      val directCdn: Boolean,
  )
  ```
- [ ] Parse `expiresAt` to `Instant` trong repository layer

### 2. URL Refresh Interceptor (OkHttp)
- [ ] Tạo `android/.../utils/FshareUrlRefreshInterceptor.kt`:
  ```kotlin
  @Singleton
  class FshareUrlRefreshInterceptor @Inject constructor(
      private val playbackRepo: Lazy<PlaybackRepository>,
  ) : Interceptor {

      override fun intercept(chain: Interceptor.Chain): Response {
          val request = chain.request()
          val response = chain.proceed(request)

          // Refresh on 403 OR 404 from fshare CDN (URL expired/single-use triggered)
          if ((response.code == 403 || response.code == 404) && isFshareCdnDomain(request.url)) {
              response.close()
              val mediaId = request.tag(MediaIdTag::class.java)?.id
                  ?: return fallbackErrorResponse(chain, request)

              val freshUrl = try {
                  runBlocking(Dispatchers.IO) {
                      withTimeout(5_000) { playbackRepo.get().refreshStreamUrl(mediaId).url }
                  }
              } catch (e: Exception) {
                  return fallbackErrorResponse(chain, request)
              }

              val newRequest = request.newBuilder()
                  .url(freshUrl)
                  .build()
              return chain.proceed(newRequest)
          }

          return response
      }

      private fun isFshareCdnDomain(url: HttpUrl): Boolean =
          url.host.contains("fshare.vn") || url.host.startsWith("download.fs")
  }

  data class MediaIdTag(val id: Long)
  ```
- [ ] Register OkHttpClient used by ExoPlayer DataSource (Hilt module):
  ```kotlin
  @Provides @Singleton
  fun provideExoPlayerDataSourceFactory(okHttpClient: OkHttpClient): DataSource.Factory =
      OkHttpDataSource.Factory(okHttpClient).setUserAgent("Velox/${BuildConfig.VERSION_NAME}")
  ```
- [ ] Tag requests với `MediaIdTag`: done trong PlaybackRepository khi build MediaItem
- [ ] Timeout 5s cho refresh (user không wait forever); fail fast → ExoPlayer shows error → user retry

### 3. ~~Proactive Refresh~~ — DROPPED

**Decision:** KHÔNG implement proactive timer-based refresh. Research xác nhận fshare URL có thể single-use → swap URL giữa chừng streaming = ExoPlayer phải re-seek vào URL mới → giật. Reactive-only an toàn hơn.

**Rationale:**
- Initial URL thường survive toàn bộ single HTTP GET (ExoPlayer stream file qua 1 long-lived connection).
- Khi ExoPlayer gặp 403/404 (seek, resume, network reconnect) → interceptor refresh ngay.
- Proactive refresh không giúp vì URL mới cũng single-use, thay thế URL đang stream = reset connection.

Nếu future testing show initial URL break giữa movie dài → reconsider với hybrid approach (but phase này KHÔNG implement).

### 4. PlaybackRepository.refreshStreamUrl
- [ ] Add method trong `PlaybackRepository.kt`:
  ```kotlin
  suspend fun refreshStreamUrl(mediaId: Long): StreamURLResponse {
      return api.getStreamUrl(mediaId) // reuse existing endpoint
  }
  ```
- [ ] NO client-side cache (fshare URLs fragile, server không cache → client cũng không cache)
- [ ] Retry policy: 3 attempts với exponential backoff (network flakiness của Velox backend, không fshare)

### 5. MediaCard Badge
- [ ] Modify `android/.../presentation/ui/components/MediaCard.kt`:
  ```kotlin
  @Composable
  fun MediaCard(media: MediaItem, ...) {
      Box {
          // existing card content
          if (media.sourceType == "fshare") {
              Icon(
                  imageVector = LucideIcons.Cloud,
                  contentDescription = "Cloud storage",
                  modifier = Modifier
                      .align(Alignment.TopEnd)
                      .padding(8.dp)
                      .size(16.dp),
                  tint = Color.White.copy(alpha = 0.7f),
              )
          }
      }
  }
  ```
- [ ] Add `sourceType` field vào MediaItem domain model (nếu chưa có)
- [ ] API response include `source_type` trong media list endpoints (backend already returns — verify)

## Manual Test Scenarios

- [ ] **Happy path**: Fshare library → pick movie → play → verify stream từ fshare CDN (adb logcat `DataSource` shows `download.fs*.fshare.vn`)
- [ ] **Verify backend bandwidth = 0**: Monitor Velox server network during playback — không traffic cho video data (chỉ initial API call)
- [ ] **Seek mid-playback**: Seek forward/backward → verify refresh interceptor trigger nếu 403 (có thể seek work với same URL, có thể không — test real-world)
- [ ] **Long movie (3h)**: Start 3h movie → verify stream không break giữa chừng (initial URL có thể survive toàn single GET; nếu break → interceptor catch)
- [ ] **Network drop mid-playback**: Turn WiFi off 10s → on → ExoPlayer retry → interceptor fresh URL nếu CDN trả 403
- [ ] **Session expired on backend**: Force-expire fshare session (wait 35min) → playback request → resolver auto-relogin → success
- [ ] **Mixed library**: Scroll feed with local + fshare → both types play correctly
- [ ] **Background playback**: Backgrounded app tiếp tục stream → verify (kết hợp Background ExoPlayer implementation pending)
- [ ] **Free account (NotVIP)**: Add fshare library với free account → GetStreamURL returns `ErrNotVIP` → user sees clear error message

## Acceptance Criteria

- [ ] ExoPlayer plays fshare URL successfully (MKV 4K tested on dev device)
- [ ] URL refresh interceptor triggers on 403/404 — verified via adb logcat filtering OkHttp events
- [ ] ☁️ badge hiển thị đúng cho fshare media, không hiện cho local media
- [ ] Unit tests: `FshareUrlRefreshInterceptor` với mocked PlaybackRepository + OkHttp MockWebServer
- [ ] Integration test: start playback → simulate 403 on 2nd seek → verify URL refreshes + playback resumes
- [ ] Zero regression trên local media playback
- [ ] Free account path: clear error UI, không crash

## Gotchas

- **OkHttp interceptor scope**: Interceptor chỉ wrap OkHttp client. ExoPlayer default dùng `DefaultHttpDataSource` (không OkHttp). PHẢI configure `OkHttpDataSource.Factory` rõ ràng (Hilt module).
- **Tag lost across retries:** OkHttp interceptor retry request lose tag. Giải pháp: build new request từ scratch với tag re-attached khi refresh URL.
- **runBlocking trong interceptor:** OkHttp interceptor chạy trên network thread (không UI) — OK. Strict timeout 5s cho refresh call (user không wait forever).
- **Refresh loop risk**: Nếu refresh returns URL cũ → interceptor loop. Guard: track refresh attempts per request (max 2), fail sau đó.
- **ExoPlayer cache:** Disable ExoPlayer DataSource cache cho fshare source type — cache hit URL cũ = 403 loop. Check `ExoPlayer.Builder` config.
- **Free account UX:** `ErrNotVIP` từ backend → show clear dialog "Fshare account must be VIP to stream". Deep-link to fshare.vn upgrade page.
- **Same-URL seek:** ExoPlayer seek dùng HTTP Range request trên same URL. Fshare CDN có thể accept multiple ranges on same URL (common pattern) → không cần refresh trên seek. Interceptor catch 403 nếu khác.

## Out of Scope

- Cast fshare media to Chromecast (Chromecast nhận direct URL → Cast SDK needs app server URL reachable; phức tạp — defer)
- Offline download fshare content (Expo-file-system + background download — phase sau)
- Adaptive bitrate cho fshare (không có HLS, ExoPlayer play single file)
- Analytics tracking cho fshare stream sessions (fshare CDN không send callback — blind spot)
