package com.velox.app.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

// Server Info
@Serializable
data class AppVersionDto(
    val platform: String,
    @SerialName("version_name") val versionName: String,
    @SerialName("version_code") val versionCode: Int,
    @SerialName("download_url") val downloadUrl: String,
    @SerialName("is_mandatory") val isMandatory: Boolean,
    @SerialName("release_notes") val releaseNotes: String? = null,
    val error: String? = null
)

@Serializable
data class ServerInfoDto(
    val version: String,
    val uptime: String,
    @SerialName("go_version") val goVersion: String,
    val os: String,
    val arch: String,
    @SerialName("ffmpeg_version") val ffmpegVersion: String,
    val database: String,
    @SerialName("hw_accel") val hwAccel: String,
    @SerialName("media_count") val mediaCount: Int,
    @SerialName("series_count") val seriesCount: Int,
    @SerialName("user_count") val userCount: Int,
    @SerialName("total_size_bytes") val totalSize: Long,
)

@Serializable
data class LibraryStatsDto(
    val id: Int,
    val name: String,
    val type: String,
    @SerialName("item_count") val itemCount: Int,
    @SerialName("file_count") val fileCount: Int,
    @SerialName("total_size_bytes") val totalSize: Long,
    @SerialName("last_scanned") val lastScanned: String? = null,
)

// Admin Users
@Serializable
data class AdminUserDto(
    val id: Int,
    val username: String,
    @SerialName("display_name") val displayName: String,
    @SerialName("is_admin") val isAdmin: Boolean,
    @SerialName("last_active") val lastActive: String? = null,
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("updated_at") val updatedAt: String? = null,
    @SerialName("max_streams") val maxStreams: Int = 0,
    @SerialName("allowed_libraries") val allowedLibraries: List<Int>? = null,
)

@Serializable
data class CreateUserRequest(
    val username: String,
    val password: String,
    @SerialName("display_name") val displayName: String,
    @SerialName("is_admin") val isAdmin: Boolean = false,
    @SerialName("max_streams") val maxStreams: Int = 3,
    @SerialName("allowed_libraries") val allowedLibraries: List<Int>? = null,
)

@Serializable
data class UpdateUserRequest(
    @SerialName("display_name") val displayName: String? = null,
    val password: String? = null,
    @SerialName("is_admin") val isAdmin: Boolean? = null,
    @SerialName("max_streams") val maxStreams: Int? = null,
    @SerialName("allowed_libraries") val allowedLibraries: List<Int>? = null,
)

// Admin Libraries
@Serializable
data class CreateLibraryRequest(
    val name: String,
    val type: String, // "movie", "series", "music"
    val paths: List<String>,
    @SerialName("metadata_profile") val metadataProfile: String? = null,
)

@Serializable
data class UpdateLibraryRequest(
    val name: String? = null,
    val paths: List<String>? = null,
    @SerialName("metadata_profile") val metadataProfile: String? = null,
)

// Admin Activity
@Serializable
data class ActivityLogDto(
    val id: Int,
    val username: String? = null,
    val action: String,
    @SerialName("media_title") val mediaTitle: String? = null,
    val details: String? = null,
    @SerialName("created_at") val createdAt: String,
    @SerialName("ip_address") val ipAddress: String? = null,
)

// Admin Tasks
@Serializable
data class TaskDto(
    val name: String,
    val interval: String,
    val running: Boolean,
    @SerialName("last_run") val lastRun: String? = null,
    @SerialName("next_run") val nextRun: String,
)

@Serializable
data class UpdateTaskIntervalRequest(
    val interval: String,
)

// Admin Webhooks
@Serializable
data class WebhookDto(
    val id: Int,
    val url: String,
    val events: List<String>,
    val active: Boolean,
    @SerialName("created_at") val createdAt: String,
    @SerialName("last_triggered") val lastTriggered: String? = null,
)

@Serializable
data class CreateWebhookRequest(
    val url: String,
    val events: List<String>,
    val active: Boolean = true,
)

@Serializable
data class UpdateWebhookRequest(
    val url: String? = null,
    val events: List<String>? = null,
    val active: Boolean? = null,
)

@Serializable
data class WebhookTestResponse(
    val success: Boolean,
    val message: String,
    @SerialName("response_code") val responseCode: Int? = null,
)

// Admin Pretranscode
@Serializable
data class PretranscodeJobDto(
    val id: Int,
    @SerialName("media_id") val mediaId: Int,
    @SerialName("media_title") val mediaTitle: String,
    val status: String, // "queued", "running", "completed", "failed", "cancelled"
    val progress: Float,
    @SerialName("output_format") val outputFormat: String,
    @SerialName("created_at") val createdAt: String,
    @SerialName("completed_at") val completedAt: String? = null,
)

@Serializable
data class CreatePretranscodeRequest(
    @SerialName("media_ids") val mediaIds: List<Int>? = null,
    @SerialName("library_id") val libraryId: Int? = null,
    @SerialName("output_format") val outputFormat: String = "hls",
    @SerialName("quality") val quality: String = "1080p",
)

// Admin Markers
@Serializable
data class SkipMarkerDto(
    val id: Int,
    @SerialName("media_id") val mediaId: Int,
    val type: String, // "intro", "credits", "sponsor", "chapter"
    val start: Float,
    val end: Float,
    val title: String? = null,
)

@Serializable
data class CreateMarkerRequest(
    @SerialName("media_id") val mediaId: Int,
    val type: String,
    val start: Float,
    val end: Float,
    val title: String? = null,
)

@Serializable
data class UpdateMarkerRequest(
    val type: String? = null,
    val start: Float? = null,
    val end: Float? = null,
    val title: String? = null,
)

// Admin Settings
@Serializable
data class AdminSettingsDto(
    @SerialName("server_name") val serverName: String,
    @SerialName("max_concurrent_streams") val maxConcurrentStreams: Int,
    @SerialName("transcoding_bitrate") val transcodingBitrate: Int,
    @SerialName("direct_play_codecs") val directPlayCodecs: List<String>,
    @SerialName("allowed_audio_codecs") val allowedAudioCodecs: List<String>,
    @SerialName("allowed_subtitle_formats") val allowedSubtitleFormats: List<String>,
    @SerialName("metadata_server") val metadataServer: String? = null,
    @SerialName("log_level") val logLevel: String,
)

@Serializable
data class UpdateAdminSettingsRequest(
    @SerialName("server_name") val serverName: String? = null,
    @SerialName("max_concurrent_streams") val maxConcurrentStreams: Int? = null,
    @SerialName("transcoding_bitrate") val transcodingBitrate: Int? = null,
    @SerialName("direct_play_codecs") val directPlayCodecs: List<String>? = null,
    @SerialName("allowed_audio_codecs") val allowedAudioCodecs: List<String>? = null,
    @SerialName("allowed_subtitle_formats") val allowedSubtitleFormats: List<String>? = null,
    @SerialName("log_level") val logLevel: String? = null,
)

// Provider Settings
@Serializable
data class ProviderSettingsDto(
    @SerialName("api_key") val apiKey: String,
    @SerialName("has_builtin") val hasBuiltin: Boolean = false,
)

@Serializable
data class UpdateProviderRequest(
    @SerialName("api_key") val apiKey: String,
)

@Serializable
data class AITranslationSettingsDto(
    val provider: String,
    @SerialName("api_key") val apiKey: String,
    @SerialName("base_url") val baseUrl: String,
    val model: String,
)

@Serializable
data class UpdateAITranslationRequest(
    val provider: String,
    @SerialName("api_key") val apiKey: String,
    @SerialName("base_url") val baseUrl: String,
    val model: String,
)

// OpenSubtitles Settings
@Serializable
data class OpenSubsSettingsDto(
    @SerialName("api_key") val apiKey: String,
    val username: String,
    @SerialName("password_set") val passwordSet: Boolean = false,
)

@Serializable
data class UpdateOpenSubsRequest(
    @SerialName("api_key") val apiKey: String,
    val username: String,
    val password: String? = null,
)

// Auto-Subtitle Settings
@Serializable
data class AutoSubSettingsDto(
    val languages: String,
)

@Serializable
data class UpdateAutoSubRequest(
    val languages: String,
)

// Playback Settings
@Serializable
data class PlaybackSettingsDto(
    @SerialName("playback_mode") val playbackMode: String,
)

@Serializable
data class UpdatePlaybackRequest(
    @SerialName("playback_mode") val playbackMode: String,
)

// Cinema Settings
@Serializable
data class CinemaSettingsDto(
    val enabled: Boolean,
    @SerialName("max_trailers") val maxTrailers: String,
    @SerialName("has_intro") val hasIntro: Boolean = false,
)

@Serializable
data class UpdateCinemaRequest(
    val enabled: Boolean? = null,
    @SerialName("max_trailers") val maxTrailers: String? = null,
)

// Pre-transcode Settings
@Serializable
data class PretranscodeSettingsDto(
    val enabled: Boolean,
    val schedule: String,
    val concurrency: String,
)

@Serializable
data class UpdatePretranscodeRequest(
    val enabled: Boolean? = null,
    val schedule: String? = null,
    val concurrency: String? = null,
)

@Serializable
data class PretranscodeStatusDto(
    val enabled: Boolean,
    val schedule: String,
    val concurrency: Int,
    val paused: Boolean,
    val total: Int,
    val done: Int,
    val encoding: Int,
    val failed: Int,
    val queued: Int,
    @SerialName("disk_used") val diskUsed: Long,
    @SerialName("current_file") val currentFile: String?,
    val speed: String?,
)

@Serializable
data class PretranscodeProfileDto(
    val id: Int,
    val name: String,
    val height: Int,
    @SerialName("video_bitrate") val videoBitrate: Int,
    @SerialName("audio_bitrate") val audioBitrate: Int,
    @SerialName("video_codec") val videoCodec: String,
    @SerialName("audio_codec") val audioCodec: String,
    val enabled: Boolean,
)

@Serializable
data class EstimateProfileDto(
    @SerialName("profile_id") val profileId: Int,
    @SerialName("profile_name") val profileName: String,
    val height: Int,
    @SerialName("estimated_gb") val estimatedGb: Double,
    @SerialName("file_count") val fileCount: Int,
)

@Serializable
data class StorageEstimateDto(
    val profiles: List<EstimateProfileDto>?,
    @SerialName("total_bytes") val totalBytes: Long,
    @SerialName("disk_free_bytes") val diskFreeBytes: Long,
    @SerialName("file_count") val fileCount: Int,
)

@Serializable
data class MarkerStatsDto(
    @SerialName("total_markers") val totalMarkers: Int,
    @SerialName("intro_markers") val introMarkers: Int,
    @SerialName("credits_markers") val creditsMarkers: Int,
    @SerialName("chapter_source") val chapterSource: Int,
    @SerialName("fingerprint_source") val fingerprintSource: Int,
    @SerialName("manual_source") val manualSource: Int,
    @SerialName("files_with_intro") val filesWithIntro: Int,
    @SerialName("files_with_credits") val filesWithCredits: Int,
    @SerialName("total_files") val totalFiles: Int,
)

@Serializable
data class BackfillMarkersRequest(
    @SerialName("library_id") val libraryId: Int? = null,
)

@Serializable
data class BackfillResultDto(
    val processed: Int,
    val skipped: Int,
    val errors: List<String>? = null,
)

@Serializable
data class AutoTranslateSettingsDto(
    val enabled: Boolean,
    val languages: String,
)

@Serializable
data class UpdateAutoTranslateRequest(
    val enabled: Boolean,
    val languages: String,
)
