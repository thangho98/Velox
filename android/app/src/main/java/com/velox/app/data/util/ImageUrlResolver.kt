package com.velox.app.data.util

/**
 * Resolve an image path (TMDb, local://, or already-absolute) to a full URL.
 *
 * Mirrors the logic in packages/shared/lib/image.ts on the web side:
 * - `local://foo.jpg` → `{apiBase}/images/local/media/foo.jpg`
 * - `http://…` or `/api/…` → pass-through
 * - bare path like `tmdb/path.jpg` → TMDb CDN URL
 */
object ImageUrlResolver {
    private var baseUrlProvider: (() -> String?)? = null

    fun setBaseUrlProvider(provider: () -> String?) {
        baseUrlProvider = provider
    }

    fun resolve(path: String?, size: String = "w500"): String? {
        if (path.isNullOrEmpty()) return null
        if (path.startsWith("http")) return path
        if (path.startsWith("/api/")) {
            var baseUrl = baseUrlProvider?.invoke()?.removeSuffix("/") ?: return path
            if (baseUrl.endsWith("/api")) {
                baseUrl = baseUrl.removeSuffix("/api")
            }
            return "$baseUrl$path"
        }

        // local:// poster/backdrop uploaded by the user
        if (path.startsWith("local://")) {
            val baseUrl = baseUrlProvider?.invoke() ?: return null
            return "${baseUrl.removeSuffix("/")}/images/local/media/${path.removePrefix("local://")}"
        }

        val baseUrl = baseUrlProvider?.invoke()?.removeSuffix("/") ?: return null

        // Strip leading slash if present
        val cleaned = if (path.startsWith("/")) path.substring(1) else path

        // Proxy through Velox backend like web app
        return "$baseUrl/images/tmdb/$size/$cleaned"
    }
}
