package com.velox.app.data.util

/**
 * Resolve an image path returned by the backend to a full URL.
 *
 * The backend normalizes everything to fully-routed relative paths like
 * `/api/images/local/media/42/poster.jpg` or `/api/images/tmdb/abc.jpg`.
 * Size is expressed as `?width=N` — the backend maps that to the nearest
 * TMDb bucket (w92/w185/.../original).
 *
 * External URLs (http/https) pass through untouched.
 *
 * `size` accepts either a numeric width ("500") or a TMDb size token
 * ("w500", "h632", "original") for source compatibility with older callers.
 */
object ImageUrlResolver {
    private var baseUrlProvider: (() -> String?)? = null

    fun setBaseUrlProvider(provider: () -> String?) {
        baseUrlProvider = provider
    }

    fun resolve(path: String?, size: String? = "w500"): String? {
        if (path.isNullOrEmpty()) return null
        if (path.startsWith("http://") || path.startsWith("https://")) return path

        val base = baseUrlProvider?.invoke()?.removeSuffix("/") ?: return null
        val baseWithoutApi = if (base.endsWith("/api")) base.removeSuffix("/api") else base

        val joined = when {
            path.startsWith("/api/") -> "$baseWithoutApi$path"
            path.startsWith("/") -> "$baseWithoutApi$path"
            else -> "$baseWithoutApi/$path"
        }

        val width = coerceWidth(size) ?: return joined
        val separator = if ('?' in joined) '&' else '?'
        return "$joined${separator}width=$width"
    }

    // Strip TMDb size prefixes (`w500`, `h632`) down to the numeric width the
    // backend expects. Returns null to skip the width query param entirely.
    private fun coerceWidth(raw: String?): String? {
        if (raw.isNullOrBlank()) return null
        if (raw == "original") return "original"
        val match = Regex("^[a-zA-Z](\\d+)$").matchEntire(raw)
        if (match != null) return match.groupValues[1]
        val n = raw.toIntOrNull()
        return if (n != null && n > 0) n.toString() else null
    }
}
