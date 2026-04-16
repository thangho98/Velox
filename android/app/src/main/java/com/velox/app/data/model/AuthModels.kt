package com.velox.app.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class LoginRequest(
    val username: String,
    val password: String,
)

@Serializable
data class DataWrapper<T>(
    val data: T
)

@Serializable
data class LoginResponse(
    @SerialName("access_token") val accessToken: String,
    @SerialName("refresh_token") val refreshToken: String,
    @SerialName("expires_in") val expiresIn: Int,
    val user: UserDto,
)

@Serializable
data class RefreshRequest(
    @SerialName("refresh_token") val refreshToken: String,
)

@Serializable
data class RefreshResponse(
    @SerialName("access_token") val accessToken: String,
    @SerialName("refresh_token") val refreshToken: String,
    @SerialName("expires_in") val expiresIn: Int,
)

@Serializable
data class UserDto(
    val id: Int,
    val username: String,
    @SerialName("display_name") val displayName: String,
    @SerialName("is_admin") val isAdmin: Boolean,
    @SerialName("profile_path") val profilePath: String? = null,
    val profile: com.velox.app.data.model.ImageResourceDto? = null,
)

@Serializable
data class ChangePasswordRequest(
    @SerialName("old_password") val oldPassword: String,
    @SerialName("new_password") val newPassword: String,
)

@Serializable
data class SessionDto(
    val id: Int,
    @SerialName("device_name") val deviceName: String?,
    @SerialName("ip_address") val ipAddress: String?,
    @SerialName("last_active_at") val lastActiveAt: String,
    @SerialName("created_at") val createdAt: String,
)

@Serializable
data class MessageResponse(
    val message: String? = null,
    val error: String? = null,
)
