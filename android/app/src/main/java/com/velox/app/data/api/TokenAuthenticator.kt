package com.velox.app.data.api

import com.velox.app.data.model.RefreshRequest
import kotlinx.coroutines.runBlocking
import okhttp3.Authenticator
import okhttp3.Request
import okhttp3.Response
import okhttp3.Route
import javax.inject.Inject
import javax.inject.Provider

class TokenAuthenticator @Inject constructor(
    private val authManager: AuthManager,
    private val apiProvider: Provider<VeloxApiProvider>
) : Authenticator {

    override fun authenticate(route: Route?, response: Response): Request? {
        // Skip auth for login and refresh endpoints to prevent infinite loops
        val path = response.request.url.encodedPath
        if (path.contains("/auth/login") || path.contains("/auth/refresh")) {
            return null
        }

        return runBlocking {
            synchronized(this@TokenAuthenticator) {
                // Get the current token from storage
                val currentToken = authManager.getAccessTokenSync()
                // Get the token that was sent with the failed request
                val requestToken = response.request.header("Authorization")?.removePrefix("Bearer ")

                // If the tokens are different, it means another thread already refreshed the token.
                // We can just retry the request with the new token.
                if (currentToken != null && currentToken != requestToken) {
                    return@synchronized response.request.newBuilder()
                        .header("Authorization", "Bearer $currentToken")
                        .build()
                }

                // Need to refresh
                val refreshToken = authManager.getRefreshTokenSync()
                if (refreshToken == null) {
                    runBlocking { authManager.clearAuth() }
                    return@synchronized null
                }

                try {
                    val api = apiProvider.get().getApi()
                    // Calling refreshToken will suspend but we are in runBlocking
                    val refreshResponse = runBlocking { api.refreshToken(RefreshRequest(refreshToken)) }

                    if (refreshResponse.isSuccessful) {
                        val data = refreshResponse.body()?.data
                        if (data != null) {
                            runBlocking {
                                authManager.saveTokens(
                                    accessToken = data.accessToken,
                                    refreshToken = data.refreshToken,
                                    expiresIn = data.expiresIn
                                )
                            }
                            return@synchronized response.request.newBuilder()
                                .removeHeader("Authorization")
                                .addHeader("Authorization", "Bearer ${data.accessToken}")
                                .build()
                        }
                    }
                    
                    // If refresh failed (e.g. 401 or invalid token), clear auth 
                    runBlocking { authManager.clearAuth() }
                    null
                } catch (e: Exception) {
                    null
                }
            }
        }
    }
}
