package com.velox.app.data.repository

import com.velox.app.data.api.AuthManager
import com.velox.app.data.api.VeloxApiProvider
import com.velox.app.data.model.ChangePasswordRequest
import com.velox.app.data.model.LoginRequest
import com.velox.app.data.model.RefreshRequest
import com.velox.app.data.model.toDomain
import com.velox.app.data.util.ImageUrlResolver
import com.velox.app.domain.model.User
import com.velox.app.domain.repository.AuthRepository
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepositoryImpl @Inject constructor(
    private val apiProvider: VeloxApiProvider,
    private val authManager: AuthManager,
) : AuthRepository {

    private val api: com.velox.app.data.api.VeloxApi
        get() = apiProvider.getApi()

    init {
        ImageUrlResolver.setBaseUrlProvider { apiProvider.getServerUrlSync() }
    }

    override suspend fun login(username: String, password: String): Result<User> {
        return try {
            val response = api.login(LoginRequest(username, password))
            if (response.isSuccessful) {
                val body = response.body()!!.data
                authManager.saveTokens(
                    accessToken = body.accessToken,
                    refreshToken = body.refreshToken,
                    expiresIn = body.expiresIn,
                )
                val profileImage = body.user.profile?.toDomain()
                authManager.saveUser(
                    userId = body.user.id,
                    username = body.user.username,
                    displayName = body.user.displayName,
                    isAdmin = body.user.isAdmin,
                    profilePath = getFullUrl(body.user.profilePath),
                    profileImage = profileImage,
                )
                Result.success(
                    User(
                        id = body.user.id,
                        username = body.user.username,
                        displayName = body.user.displayName,
                        isAdmin = body.user.isAdmin,
                        profilePath = getFullUrl(body.user.profilePath),
                        profile = profileImage,
                    ),
                )
            } else {
                val errorBody = response.errorBody()?.string()
                Result.failure(Exception(errorBody ?: "Login failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun logout(): Result<Unit> {
        return try {
            api.logout()
            authManager.clearAuth()
            Result.success(Unit)
        } catch (e: Exception) {
            authManager.clearAuth()
            Result.success(Unit) // Still clear local auth even if API fails
        }
    }

    override suspend fun getMe(): Result<User> {
        return try {
            val response = api.getMe()
            if (response.isSuccessful) {
                val dto = response.body()!!.data
                val profileImage = dto.profile?.toDomain()
                val user = User(
                    id = dto.id,
                    username = dto.username,
                    displayName = dto.displayName,
                    isAdmin = dto.isAdmin,
                    profilePath = getFullUrl(dto.profilePath),
                    profile = profileImage,
                )
                // Sync with local storage
                authManager.saveUser(
                    userId = user.id,
                    username = user.username,
                    displayName = user.displayName,
                    isAdmin = user.isAdmin,
                    profilePath = user.profilePath,
                    profileImage = profileImage,
                )
                Result.success(user)
            } else {
                Result.failure(Exception("Failed to fetch user info"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    private fun getFullUrl(path: String?): String? = ImageUrlResolver.resolve(path)

    override suspend fun refreshToken(): Result<Unit> {
        return try {
            val refreshToken = authManager.getRefreshTokenSync()
                ?: return Result.failure(Exception("No refresh token"))

            val response = api.refreshToken(RefreshRequest(refreshToken))
            if (response.isSuccessful) {
                val body = response.body()!!.data
                authManager.saveTokens(
                    accessToken = body.accessToken,
                    refreshToken = body.refreshToken,
                    expiresIn = body.expiresIn,
                )
                Result.success(Unit)
            } else {
                Result.failure(Exception("Token refresh failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override fun isLoggedIn(): Flow<Boolean> = authManager.isLoggedIn

    override fun getCurrentUser(): Flow<User?> {
        return combine(
            authManager.userId,
            authManager.username,
            authManager.displayName,
            authManager.isAdmin,
            authManager.profilePath,
            authManager.profileImageResource,
        ) { values ->
            val userId = values[0] as? Int
            val username = values[1] as? String
            val displayName = values[2] as? String
            val isAdmin = values[3] as? Boolean ?: false
            val profilePath = values[4] as? String
            @Suppress("UNCHECKED_CAST")
            val profileImage = values[5] as? com.velox.app.domain.model.ImageResource
            if (userId != null && username != null) {
                User(
                    id = userId,
                    username = username,
                    displayName = displayName ?: username,
                    isAdmin = isAdmin,
                    profilePath = profilePath,
                    profile = profileImage,
                )
            } else {
                null
            }
        }
    }

    override suspend fun changePassword(oldPassword: String, newPassword: String): Result<Unit> {
        return try {
            val response = api.changePassword(
                ChangePasswordRequest(oldPassword, newPassword),
            )
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                val errorBody = response.errorBody()?.string()
                Result.failure(Exception(errorBody ?: "Password change failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
