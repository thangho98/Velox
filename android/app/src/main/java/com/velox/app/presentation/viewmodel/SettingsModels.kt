package com.velox.app.presentation.viewmodel

import androidx.annotation.StringRes
import com.velox.app.R

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

enum class SettingsSection(@StringRes val titleRes: Int, val icon: androidx.compose.ui.graphics.vector.ImageVector, val description: String) {
    PROFILE(R.string.settings_title_profile, com.velox.app.presentation.ui.components.LucideIcons.Person, "Personal details & avatar"),
    PREFERENCES(R.string.settings_title_preferences, com.velox.app.presentation.ui.components.LucideIcons.Settings, "Language & theme"),
    SECURITY(R.string.settings_title_security, com.velox.app.presentation.ui.components.LucideIcons.Security, "Password & 2FA"),
    SESSIONS(R.string.settings_title_sessions, com.velox.app.presentation.ui.components.LucideIcons.Devices, "Active devices"),
    METADATA(R.string.settings_title_metadata, com.velox.app.presentation.ui.components.LucideIcons.Info, "Scrapers & refresh"),
    SUBTITLES(R.string.settings_title_subtitles, com.velox.app.presentation.ui.components.LucideIcons.Subtitles, "Languages & appearance"),
    PLAYBACK(R.string.settings_title_playback, com.velox.app.presentation.ui.components.LucideIcons.PlayCircle, "Quality & transcoding"),
    CINEMA(R.string.settings_title_cinema, com.velox.app.presentation.ui.components.LucideIcons.Film, "Trailers & intros"),
    PRETRANSCODE(R.string.settings_title_pretranscode, com.velox.app.presentation.ui.components.LucideIcons.FlashOn, "Offline encoding & optimization"),
    MARKERS(R.string.settings_title_markers, com.velox.app.presentation.ui.components.LucideIcons.SkipNext, "Intros & credits"),
    DASHBOARD(R.string.settings_title_dashboard, com.velox.app.presentation.ui.components.LucideIcons.GridView, "Server status & stats"),
    LIBRARIES(R.string.settings_title_libraries, com.velox.app.presentation.ui.components.LucideIcons.Folder, "Folders & scanning"),
    USERS(R.string.settings_title_users, com.velox.app.presentation.ui.components.LucideIcons.Person, "Accounts & roles"),
    ACTIVITY(R.string.settings_title_activity, com.velox.app.presentation.ui.components.LucideIcons.Speed, "Playback history & logs"),
    TASKS(R.string.settings_title_tasks, com.velox.app.presentation.ui.components.LucideIcons.Schedule, "Background jobs"),
    WEBHOOKS(R.string.settings_title_webhooks, com.velox.app.presentation.ui.components.LucideIcons.Link, "External integrations")
}

sealed class SettingsAction {
    data class SelectSection(val section: SettingsSection) : SettingsAction()
}
