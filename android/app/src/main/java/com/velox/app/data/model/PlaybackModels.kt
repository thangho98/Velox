package com.velox.app.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class PlaybackInfoRequest(
    @SerialName("video_codecs") val videoCodecs: List<String>? = null,
    @SerialName("audio_codecs") val audioCodecs: List<String>? = null,
    val containers: List<String>? = null,
    @SerialName("max_height") val maxHeight: Int? = null,
    @SerialName("media_file_id") val mediaFileId: Int? = null,
    @SerialName("selected_audio_track") val selectedAudioTrack: Int? = null,
    @SerialName("selected_subtitle") val selectedSubtitle: String? = null,
    @SerialName("selected_subtitle_id") val selectedSubtitleId: Int? = null,
)

@Serializable
data class PlaybackInfoDto(
    @SerialName("media_id") val mediaId: Int,
    @SerialName("primary_file_id") val primaryFileId: Int,
    val method: String,
    @SerialName("stream_url") val streamUrl: String,
    @SerialName("direct_url") val directUrl: String,
    @SerialName("abr_url") val abrUrl: String? = null,
    @SerialName("pretranscode_url") val pretranscodeUrl: String? = null,  // Pre-encoded MP4
    @SerialName("hls_url") val hlsUrl: String,  // Realtime HLS transcode
    @SerialName("prefer") val prefer: String? = null,  // Playback preference: direct, pretranscode, hls
    @SerialName("stream_session_id") val streamSessionId: String,  // For stream session tracking
    @SerialName("video_codec") val videoCodec: String? = null,
    @SerialName("video_profile") val videoProfile: String? = null,
    @SerialName("video_level") val videoLevel: String? = null,
    @SerialName("video_fps") val videoFps: Float? = null,
    @SerialName("audio_codec") val audioCodec: String? = null,
    val container: String? = null,
    @SerialName("file_size") val fileSize: Long? = null,
    val bitrate: Int? = null,
    val duration: Float? = null,
    val width: Int? = null,
    val height: Int? = null,
    @SerialName("decision_reason") val decisionReason: String? = null,
    @SerialName("estimated_bitrate") val estimatedBitrate: Int? = null,
    val position: Float? = null,
    @SerialName("audio_tracks") val audioTracks: List<AudioTrackDto>? = null,
    @SerialName("subtitle_tracks") val subtitleTracks: List<SubtitleTrackDto>? = null,
    @SerialName("skip_segments") val skipSegments: List<SkipSegmentDto>? = null,
    @SerialName("available_qualities") val availableQualities: List<QualityOptionDto>? = null,
    // Pretranscode fields
    @SerialName("pt_video_codec") val ptVideoCodec: String? = null,
    @SerialName("pt_audio_codec") val ptAudioCodec: String? = null,
    @SerialName("pt_height") val ptHeight: Int? = null,
    @SerialName("pt_video_bitrate") val ptVideoBitrate: Int? = null,
    @SerialName("pt_audio_bitrate") val ptAudioBitrate: Int? = null,
)

@Serializable
data class AudioTrackDto(
    val id: Int,
    val language: String,
    val label: String,
    val codec: String,
    val channels: Int,
    val bitrate: Int? = null,
    @SerialName("sample_rate") val sampleRate: Int? = null,
    @SerialName("is_default") val isDefault: Boolean = false,
    val selected: Boolean = false,
)

@Serializable
data class SubtitleTrackDto(
    val id: Int,
    @SerialName("media_id") val mediaId: Int? = null,
    val language: String,
    val label: String,
    val format: String? = null,
    @SerialName("is_default") val isDefault: Boolean = false,
    @SerialName("is_image") val isImage: Boolean = false,
    @SerialName("file_path") val filePath: String? = null,
)

@Serializable
data class SkipSegmentDto(
    val type: String,
    val start: Float,
    val end: Float,
    val source: String? = null,
    val confidence: Float? = null,
)

@Serializable
data class QualityOptionDto(
    val height: Int,
    val label: String,
    val instant: Boolean = false,
    val source: String? = null,
    @SerialName("bitrate_kbps") val bitrateKbps: Int? = null,
)

@Serializable
data class StreamUrlResponse(
    @SerialName("direct_url") val directUrl: String,
    @SerialName("hls_url") val hlsUrl: String,
    @SerialName("stream_session_id") val streamSessionId: String,
    @SerialName("api_key") val apiKey: String,
    @SerialName("expires_in") val expiresIn: Int,
)

// Profile / Progress
@Serializable
data class UserDataDto(
    @SerialName("user_id") val userId: Int,
    @SerialName("media_id") val mediaId: Int,
    val position: Float,
    val completed: Boolean,
    @SerialName("is_favorite") val isFavorite: Boolean,
    val rating: Float? = null,
    @SerialName("play_count") val playCount: Int? = null,
    @SerialName("last_played_at") val lastPlayedAt: String? = null,
    @SerialName("updated_at") val updatedAt: String? = null,
    @SerialName("media_title") val mediaTitle: String? = null,
    @SerialName("media_poster") val mediaPoster: String? = null,
    @SerialName("media_duration") val mediaDuration: Float? = null,
)

@Serializable
data class ContinueWatchingItemDto(
    @SerialName("media_id") val mediaId: Int,
    @SerialName("series_id") val seriesId: Int? = null,
    val position: Float,
    val completed: Boolean,
    val title: String,
    @SerialName("poster_path") val posterPath: String? = null,
    @SerialName("backdrop_path") val backdropPath: String? = null,
    val poster: ImageResourceDto? = null,
    val backdrop: ImageResourceDto? = null,
    @SerialName("media_type") val mediaType: String,
    val duration: Float? = null,
    @SerialName("series_title") val seriesTitle: String? = null,
    @SerialName("season_number") val seasonNumber: Int? = null,
    @SerialName("episode_number") val episodeNumber: Int? = null,
    @SerialName("last_played_at") val lastPlayedAt: String? = null,
)

@Serializable
data class NextUpItemDto(
    @SerialName("media_id") val mediaId: Int,
    @SerialName("series_id") val seriesId: Int,
    val title: String,
    @SerialName("episode_title") val episodeTitle: String? = null,
    @SerialName("media_type") val mediaType: String,
    @SerialName("still_path") val stillPath: String? = null,
    @SerialName("backdrop_path") val backdropPath: String? = null,
    val still: ImageResourceDto? = null,
    val backdrop: ImageResourceDto? = null,
    val poster: ImageResourceDto? = null,
    val duration: Float? = null,
    @SerialName("season_number") val seasonNumber: Int,
    @SerialName("episode_number") val episodeNumber: Int,
    @SerialName("series_title") val seriesTitle: String? = null,
    @SerialName("series_poster") val seriesPoster: String? = null,
    @SerialName("last_watched_at") val lastWatchedAt: String? = null,
)

@Serializable
data class UpdateProgressRequest(
    val position: Float,
    val completed: Boolean = false,
)

@Serializable
data class ToggleFavoriteResponse(
    @SerialName("is_favorite") val isFavorite: Boolean,
)

@Serializable
data class UserPreferencesDto(
    @SerialName("user_id") val userId: Int,
    @SerialName("subtitle_language") val subtitleLanguage: String? = null,
    @SerialName("audio_language") val audioLanguage: String? = null,
    @SerialName("max_streaming_quality") val maxStreamingQuality: String? = null,
    val theme: String? = null,
    val language: String? = null,
)

@Serializable
data class UpdatePreferencesRequest(
    @SerialName("user_id") val userId: Int,
    @SerialName("subtitle_language") val subtitleLanguage: String? = null,
    @SerialName("audio_language") val audioLanguage: String? = null,
    @SerialName("max_streaming_quality") val maxStreamingQuality: String? = null,
    val theme: String? = null,
    val language: String? = null,
)
