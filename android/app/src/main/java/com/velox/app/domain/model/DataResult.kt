package com.velox.app.domain.model

/**
 * Sealed result type for repository operations.
 *
 * Unlike [kotlin.Result], this preserves error semantics so callers can
 * distinguish between network failures, auth errors, validation issues,
 * and unexpected server errors without catching generic exceptions.
 *
 * Named "DataResult" (not "ApiResult") because repositories may also
 * read from cache, database, or other local sources in the future.
 */
sealed class DataResult<out T> {
    data class Success<T>(val data: T) : DataResult<T>()
    data class Error(val error: DataError) : DataResult<Nothing>()

    val isSuccess: Boolean get() = this is Success
    val isError: Boolean get() = this is Error

    fun getOrNull(): T? = (this as? Success)?.data

    fun errorOrNull(): DataError? = (this as? Error)?.error

    inline fun <R> map(transform: (T) -> R): DataResult<R> = when (this) {
        is Success -> Success(transform(data))
        is Error -> this
    }

    inline fun onSuccess(action: (T) -> Unit): DataResult<T> {
        if (this is Success) action(data)
        return this
    }

    inline fun onError(action: (DataError) -> Unit): DataResult<T> {
        if (this is Error) action(error)
        return this
    }

    companion object {
        fun <T> success(data: T): DataResult<T> = Success(data)

        fun <T> error(
            kind: DataError.Kind,
            message: String,
            httpCode: Int? = null,
        ): DataResult<T> = Error(DataError(kind, message, httpCode))

        fun <T> unauthorized(message: String = "Unauthorized"): DataResult<T> =
            error(DataError.Kind.Unauthorized, message, 401)

        fun <T> network(message: String = "Network error"): DataResult<T> =
            error(DataError.Kind.Network, message)

        fun <T> server(message: String, httpCode: Int? = null): DataResult<T> =
            error(DataError.Kind.Server, message, httpCode)

        fun <T> notFound(message: String = "Not found"): DataResult<T> =
            error(DataError.Kind.NotFound, message, 404)

        fun <T> unknown(cause: Throwable): DataResult<T> =
            error(DataError.Kind.Unknown, cause.message ?: "Unknown error")
    }
}

/**
 * Structured error with a classifiable [kind] for branching in ViewModels/UI.
 */
data class DataError(
    val kind: Kind,
    val message: String,
    val httpCode: Int? = null,
) {
    enum class Kind {
        /** 401 — token expired or invalid */
        Unauthorized,
        /** No network connectivity or DNS failure */
        Network,
        /** 400 / 422 — request validation failed */
        Validation,
        /** 404 — resource not found */
        NotFound,
        /** 5xx — server-side failure */
        Server,
        /** Catch-all for unclassified errors */
        Unknown,
    }
}
