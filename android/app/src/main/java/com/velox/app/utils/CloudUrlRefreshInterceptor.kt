package com.velox.app.utils

import android.util.Log
import okhttp3.HttpUrl
import okhttp3.Interceptor
import okhttp3.Request
import okhttp3.Response
import java.util.concurrent.ConcurrentHashMap

/**
 * CloudUrlRefreshInterceptor reactively refreshes cloud-CDN URLs that return
 * 403 or 404 mid-stream.
 *
 * fshare direct URLs are short-lived (minutes, sometimes single-use). When
 * ExoPlayer seeks or reconnects and the CDN rejects the old URL, this
 * interceptor transparently fetches a fresh URL from the Velox backend and
 * retries the request.
 *
 * The [refresh] callback is supplied by the PlaybackRepository because the
 * Velox API client itself uses OkHttp — we need lazy access to avoid a DI
 * cycle. Tag OkHttp requests with [MediaIdTag] so the interceptor knows which
 * media to refresh. Requests without a tag or to non-cloud hosts fall through
 * untouched.
 */
class CloudUrlRefreshInterceptor(
    private val refresh: suspend (mediaId: Long) -> String?,
) : Interceptor {

    // Track retry counts per request tag to avoid infinite loops when the
    // refreshed URL also 403s immediately.
    private val retries = ConcurrentHashMap<Long, Int>()

    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        val response = chain.proceed(request)

        if (!shouldRefresh(response, request)) {
            return response
        }

        val mediaTag = request.tag(MediaIdTag::class.java) ?: return response
        val attempts = retries.merge(mediaTag.id, 1) { old, _ -> old + 1 } ?: 1
        if (attempts > MAX_RETRIES) {
            retries.remove(mediaTag.id)
            Log.w(TAG, "Refresh limit reached for media=${mediaTag.id}; giving up")
            return response
        }

        response.close()

        val freshUrl = try {
            kotlinx.coroutines.runBlocking { refresh(mediaTag.id) }
        } catch (e: Exception) {
            Log.w(TAG, "URL refresh failed for media=${mediaTag.id}: ${e.message}")
            return chain.proceed(request)
        }

        if (freshUrl.isNullOrBlank()) {
            return chain.proceed(request)
        }

        val fresh = request.newBuilder()
            .url(freshUrl)
            .tag(MediaIdTag::class.java, mediaTag) // preserve tag across retry
            .build()
        Log.d(TAG, "Retrying media=${mediaTag.id} with refreshed URL (attempt=$attempts)")
        return chain.proceed(fresh)
    }

    private fun shouldRefresh(response: Response, request: Request): Boolean {
        if (response.code != 403 && response.code != 404) return false
        return isCloudCdnHost(request.url)
    }

    /**
     * Cloud CDN host detection. fshare serves from `download.fsXX.fshare.vn`;
     * future drivers (GDrive, OneDrive) match their own patterns below.
     */
    private fun isCloudCdnHost(url: HttpUrl): Boolean {
        val host = url.host.lowercase()
        return host.contains("fshare.vn") ||
            host.startsWith("download.fs") ||
            host.endsWith(".googleusercontent.com") || // future: gdrive
            host.endsWith(".sharepoint.com")           // future: onedrive
    }

    companion object {
        private const val TAG = "CloudUrlRefresh"
        private const val MAX_RETRIES = 2
    }
}

/**
 * Request tag used by CloudUrlRefreshInterceptor. Attach to OkHttp requests
 * built for ExoPlayer DataSource so the interceptor can refresh the right URL.
 *
 * Usage:
 *
 *     okHttpClient.newCall(
 *         Request.Builder()
 *             .url(streamUrl)
 *             .tag(MediaIdTag::class.java, MediaIdTag(mediaId))
 *             .build()
 *     )
 */
data class MediaIdTag(val id: Long)
