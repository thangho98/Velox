package com.velox.app.presentation.viewmodel

data class UserPreferences(
    val subtitleLanguage: String? = null,
    val audioLanguage: String? = null,
    val maxStreamingQuality: String? = null,
    val theme: String? = null,
    val language: String? = null,
)

data class Session(
    val id: Int,
    val deviceName: String,
    val ipAddress: String,
    val lastActive: String,
    val isCurrent: Boolean,
)

data class NotificationSettings(
    val newContent: Boolean = true,
    val libraryUpdates: Boolean = true,
    val systemNotifications: Boolean = true,
)

data class SubtitleSettings(
    val defaultLanguage: String = "English",
    val autoLoad: Boolean = true,
    val fontSize: String = "Medium",
    val background: Boolean = false,
)

data class SkipMarkerSettings(
    val skipIntro: Boolean = true,
    val skipCredits: Boolean = true,
    val autoSkipSponsor: Boolean = true,
)

data class AdminUser(
    val id: Int,
    val username: String,
    val displayName: String,
    val isAdmin: Boolean,
    val lastActive: String?,
    val createdAt: String,
)

data class Webhook(
    val id: Int,
    val url: String,
    val events: List<String>,
    val active: Boolean,
)

enum class SettingsSection(val title: String, val icon: androidx.compose.ui.graphics.vector.ImageVector, val description: String) {
    PROFILE("Profile", com.velox.app.presentation.ui.components.LucideIcons.Person, "Personal details & avatar"),
    PREFERENCES("Preferences", com.velox.app.presentation.ui.components.LucideIcons.Settings, "Language & theme"),
    SECURITY("Security", com.velox.app.presentation.ui.components.LucideIcons.Security, "Password & 2FA"),
    SESSIONS("Sessions", com.velox.app.presentation.ui.components.LucideIcons.Devices, "Active devices"),
    METADATA("Metadata", com.velox.app.presentation.ui.components.LucideIcons.Info, "Scrapers & refresh"),
    SUBTITLES("Subtitles", com.velox.app.presentation.ui.components.LucideIcons.Subtitles, "Languages & appearance"),
    PLAYBACK("Playback", com.velox.app.presentation.ui.components.LucideIcons.PlayCircle, "Quality & transcoding"),
    CINEMA("Cinema Mode", com.velox.app.presentation.ui.components.LucideIcons.Film, "Trailers & intros"),
    PRETRANSCODE("Pre-transcode", com.velox.app.presentation.ui.components.LucideIcons.FlashOn, "Offline encoding & optimization"),
    MARKERS("Skip Intro / Credits", com.velox.app.presentation.ui.components.LucideIcons.SkipNext, "Intros & credits"),
    DASHBOARD("Dashboard", com.velox.app.presentation.ui.components.LucideIcons.GridView, "Server status & stats"),
    LIBRARIES("Libraries", com.velox.app.presentation.ui.components.LucideIcons.Folder, "Folders & scanning"),
    USERS("Users", com.velox.app.presentation.ui.components.LucideIcons.Person, "Accounts & roles"),
    ACTIVITY("Activity", com.velox.app.presentation.ui.components.LucideIcons.Speed, "Playback history & logs"),
    TASKS("Scheduled Tasks", com.velox.app.presentation.ui.components.LucideIcons.Schedule, "Background jobs"),
    WEBHOOKS("Webhooks", com.velox.app.presentation.ui.components.LucideIcons.Link, "External integrations")
}

sealed class SettingsAction {
    data class SelectSection(val section: SettingsSection) : SettingsAction()
}
