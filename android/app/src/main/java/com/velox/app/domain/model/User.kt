package com.velox.app.domain.model

data class User(
    val id: Int,
    val username: String,
    val displayName: String,
    val isAdmin: Boolean,
    val profilePath: String? = null,
    val profile: ImageResource? = null,
)

data class MediaItem(
    val id: Int,
    val title: String,
    val posterPath: String?,
    val backdropPath: String?,
    val poster: ImageResource? = null,
    val backdrop: ImageResource? = null,
    val year: Int?,
    val rating: Float?,
    val mediaType: String,
    val overview: String?,
    val position: Float? = null,
    val duration: Float? = null,
    val completed: Boolean? = null,
)

data class MediaDetail(
    val id: Int,
    val title: String,
    val overview: String?,
    val posterPath: String?,
    val backdropPath: String?,
    val poster: ImageResource? = null,
    val backdrop: ImageResource? = null,
    val logo: ImageResource? = null,
    val thumb: ImageResource? = null,
    val rating: Float?,
    val duration: Float?,
    val releaseDate: String?,
    val mediaType: String = "movie",
    val seriesId: Int? = null,
    val seasonId: Int? = null,
    val genres: List<Genre>,
    val credits: Credits?,
    val similar: List<MediaItem>,
)

data class Genre(
    val id: Int,
    val name: String,
)

data class Credits(
    val cast: List<CastMember>,
    val crew: List<CrewMember>,
)

data class CastMember(
    val id: Int,
    val name: String,
    val character: String?,
    val profilePath: String?,
    val order: Int?,
)

data class CrewMember(
    val id: Int,
    val name: String,
    val job: String?,
    val department: String?,
    val profilePath: String?,
)

data class SeriesItem(
    val id: Int,
    val title: String,
    val posterPath: String?,
    val backdropPath: String?,
    val poster: ImageResource? = null,
    val backdrop: ImageResource? = null,
    val year: Int?,
    val rating: Float?,
    val overview: String?,
    val seasonCount: Int?,
    val episodeCount: Int?,
)

data class SeriesDetail(
    val id: Int,
    val title: String,
    val overview: String?,
    val posterPath: String?,
    val backdropPath: String?,
    val poster: ImageResource? = null,
    val backdrop: ImageResource? = null,
    val logo: ImageResource? = null,
    val thumb: ImageResource? = null,
    val status: String?,
    val network: String?,
    val firstAirDate: String?,
    val genres: List<Genre>,
    val seasons: List<Season>,
    val credits: Credits?,
    val metadataLocked: Boolean = false,
)

data class Season(
    val id: Int,
    val seriesId: Int,
    val seasonNumber: Int,
    val title: String?,
    val overview: String?,
    val posterPath: String?,
    val episodeCount: Int?,
)

data class Episode(
    val id: Int,
    val mediaId: Int,  // Required for playback - media file ID for this episode
    val seriesId: Int,
    val seasonId: Int,
    val seasonNumber: Int?,
    val episodeNumber: Int,
    val title: String,
    val overview: String?,
    val stillPath: String?,
    val still: ImageResource? = null,
    val airDate: String?,
    val duration: Float?,
)

data class ContinueWatchingItem(
    val mediaId: Int,
    val seriesId: Int?,
    val position: Float,
    val completed: Boolean,
    val title: String,
    val posterPath: String?,
    val backdropPath: String?,
    val poster: ImageResource? = null,
    val backdrop: ImageResource? = null,
    val mediaType: String,
    val duration: Float?,
    val seriesTitle: String?,
    val seasonNumber: Int?,
    val episodeNumber: Int?,
)

data class NextUpItem(
    val mediaId: Int,
    val seriesId: Int,
    val title: String,
    val episodeTitle: String?,
    val stillPath: String?,
    val backdropPath: String?,
    val still: ImageResource? = null,
    val backdrop: ImageResource? = null,
    val poster: ImageResource? = null,
    val duration: Float?,
    val seasonNumber: Int,
    val episodeNumber: Int,
    val seriesTitle: String?,
    val seriesPoster: String?,
)

data class PlaybackInfo(
    val mediaId: Int,
    val primaryFileId: Int,
    val method: String,
    val streamUrl: String,
    val directUrl: String,
    val abrUrl: String?,
    // HLS / Pretranscode support
    val pretranscodeUrl: String? = null,  // Pre-encoded MP4 URL
    val hlsUrl: String,  // Realtime HLS transcode URL
    val prefer: String? = null,  // Playback preference: direct, pretranscode, hls
    val streamSessionId: String,  // Stream session tracking
    // Pretranscode metadata
    val ptVideoCodec: String? = null,
    val ptAudioCodec: String? = null,
    val ptHeight: Int? = null,
    val ptVideoBitrate: Int? = null,
    val ptAudioBitrate: Int? = null,
    // Media details
    val videoCodec: String? = null,
    val videoProfile: String? = null,
    val videoLevel: String? = null,
    val videoFps: Float? = null,
    val audioCodec: String? = null,
    val container: String? = null,
    val fileSize: Long? = null,
    val bitrate: Int? = null,
    val decisionReason: String? = null,
    val estimatedBitrate: Int? = null,
    //
    val position: Float?,
    val duration: Float?,
    val width: Int?,
    val height: Int?,
    val audioTracks: List<AudioTrack>,
    val subtitleTracks: List<SubtitleTrack>,
    val skipSegments: List<SkipSegment>,
    val availableQualities: List<QualityOption>,
)

data class AudioTrack(
    val id: Int,
    val language: String,
    val label: String,
    val codec: String,
    val channels: Int,
    val bitrate: Int? = null,
    val sampleRate: Int? = null,
    val isDefault: Boolean,
    val selected: Boolean,
)

data class SubtitleTrack(
    val id: Int,
    val language: String,
    val label: String,
    val format: String?,
    val isDefault: Boolean,
    val isImage: Boolean,
)

data class SkipSegment(
    val type: String,
    val start: Float,
    val end: Float,
)

data class QualityOption(
    val height: Int,
    val label: String,
    val instant: Boolean,
)

data class Library(
    val id: Int,
    val name: String,
    val type: String,
    val paths: List<String>,
    val itemCount: Int?,
)

data class BrowseItem(
    val name: String,
    val path: String,
    val type: String?,
    val isFolder: Boolean,
    val mediaId: Int? = null,
    val mediaType: String? = null,
    val posterPath: String? = null,
    val poster: ImageResource? = null,
    val backdropPath: String? = null,
    val backdrop: ImageResource? = null,
)

data class MediaWithFilesInfo(
    val mediaType: String,
    val title: String,
    val seriesId: Int?,
    val seasonId: Int?,
    val seasonNumber: Int? = null,
    val episodeNumber: Int? = null,
    val episodeOverview: String? = null,
)

data class WatchProgress(
    val mediaId: Int,
    val position: Float,
    val completed: Boolean,
    val isFavorite: Boolean,
    val duration: Float?,
)

data class MediaFile(
    val id: Int,
    val mediaId: Int,
    val filePath: String,
    val fileSize: Long,
    val duration: Float?,
    val width: Int?,
    val height: Int?,
    val videoCodec: String?,
    val audioCodec: String?,
    val container: String?,
    val bitrate: Int?,
    val isPrimary: Boolean,
)

/**
 * Domain-level request params for playback info.
 * Replaces direct usage of data.model.PlaybackInfoRequest in the domain layer.
 */
data class PlaybackInfoParams(
    val videoCodecs: List<String>? = null,
    val audioCodecs: List<String>? = null,
    val containers: List<String>? = null,
    val maxHeight: Int? = null,
    val mediaFileId: Int? = null,
    val selectedAudioTrack: Int? = null,
    val selectedSubtitle: String? = null,
    val selectedSubtitleId: Int? = null,
)
