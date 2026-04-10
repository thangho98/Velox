package com.velox.app.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class NotificationDto(
    val id: Int,
    val type: String, // "new_content", "library_update", "system"
    val title: String,
    val message: String,
    @SerialName("media_id") val mediaId: Int? = null,
    @SerialName("created_at") val createdAt: String,
    val read: Boolean = false,
)

@Serializable
data class MarkReadRequest(
    val ids: List<Int>,
)

@Serializable
data class DeleteNotificationsRequest(
    val ids: List<Int>,
)

@Serializable
data class UnreadCountResponse(
    val count: Int,
)

@Serializable
data class NotificationListResponse(
    val notifications: List<NotificationDto>,
    @SerialName("unread_count") val unreadCount: Int
)
