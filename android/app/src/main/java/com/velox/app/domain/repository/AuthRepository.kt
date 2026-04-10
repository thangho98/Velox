package com.velox.app.domain.repository

import com.velox.app.domain.model.User
import kotlinx.coroutines.flow.Flow

interface AuthRepository {
    suspend fun login(username: String, password: String): Result<User>
    suspend fun logout(): Result<Unit>
    suspend fun refreshToken(): Result<Unit>
    fun isLoggedIn(): Flow<Boolean>
    fun getCurrentUser(): Flow<User?>
    suspend fun getMe(): Result<User>
    suspend fun changePassword(oldPassword: String, newPassword: String): Result<Unit>
}
