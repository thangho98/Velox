package com.velox.app.data.repository

import com.velox.app.data.model.DataWrapper
import com.velox.app.data.model.MessageResponse
import com.velox.app.domain.model.DataResult
import com.velox.app.domain.model.DataError
import retrofit2.Response
import timber.log.Timber
import java.net.ConnectException
import java.net.SocketTimeoutException
import java.net.UnknownHostException

/**
 * Converts a Retrofit [Response]<[DataWrapper]<T>> into a [DataResult]<R>.
 *
 * Handles:
 * - Network exceptions → DataError.Kind.Network
 * - 401 → DataError.Kind.Unauthorized
 * - 404 → DataError.Kind.NotFound
 * - 400/422 → DataError.Kind.Validation
 * - 5xx → DataError.Kind.Server
 * - Successful response → maps body via [transform]
 *
 * Usage:
 * ```
 * suspend fun getSessions(): DataResult<List<Session>> =
 *     safeApiCall({ api.getSessions() }) { dtoList ->
 *         dtoList.map { it.toDomain() }
 *     }
 * ```
 */
suspend fun <T, R> safeApiCall(
    apiCall: suspend () -> Response<DataWrapper<T>>,
    transform: (T) -> R,
): DataResult<R> {
    return try {
        val response = apiCall()
        if (response.isSuccessful) {
            val body = response.body()
            if (body != null) {
                DataResult.success(transform(body.data))
            } else {
                DataResult.server("Empty response body", response.code())
            }
        } else {
            classifyHttpError(response.code(), response.errorBody()?.string())
        }
    } catch (e: Exception) {
        classifyException(e)
    }
}

/**
 * Variant for [MessageResponse] endpoints that return no meaningful data.
 * Returns [DataResult]<[Unit]>.
 */
suspend fun safeMessageCall(
    apiCall: suspend () -> Response<MessageResponse>,
): DataResult<Unit> {
    return try {
        val response = apiCall()
        if (response.isSuccessful) {
            DataResult.success(Unit)
        } else {
            classifyHttpError(response.code(), response.errorBody()?.string())
        }
    } catch (e: Exception) {
        classifyException(e)
    }
}

/**
 * Variant for raw response bodies (e.g. [AppVersionDto] without [DataWrapper]).
 */
suspend fun <T, R> safeRawApiCall(
    apiCall: suspend () -> Response<T>,
    transform: (T) -> R,
): DataResult<R> {
    return try {
        val response = apiCall()
        if (response.isSuccessful) {
            val body = response.body()
            if (body != null) {
                DataResult.success(transform(body))
            } else {
                DataResult.server("Empty response body", response.code())
            }
        } else {
            classifyHttpError(response.code(), response.errorBody()?.string())
        }
    } catch (e: Exception) {
        classifyException(e)
    }
}

private fun <T> classifyHttpError(code: Int, errorBody: String?): DataResult<T> {
    val message = errorBody?.takeIf { it.isNotBlank() } ?: "HTTP $code"
    return when (code) {
        401 -> DataResult.unauthorized(message)
        404 -> DataResult.notFound(message)
        in 400..499 -> DataResult.error(DataError.Kind.Validation, message, code)
        in 500..599 -> DataResult.server(message, code)
        else -> DataResult.error(DataError.Kind.Unknown, message, code)
    }
}

private fun <T> classifyException(e: Exception): DataResult<T> {
    Timber.e(e, "API call failed")
    return when (e) {
        is UnknownHostException,
        is ConnectException,
        is SocketTimeoutException -> DataResult.network(e.message ?: "Network error")
        else -> DataResult.unknown(e)
    }
}
