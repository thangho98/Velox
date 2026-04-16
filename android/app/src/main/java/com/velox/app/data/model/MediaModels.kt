package com.velox.app.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

import kotlinx.serialization.json.JsonNames
import kotlinx.serialization.ExperimentalSerializationApi

@OptIn(ExperimentalSerializationApi::class)
@Serializable
data class MediaListItemDto(
    @JsonNames("media_id") val id: Int,
    @JsonNames("media_title") val title: String = "Unknown",
    @SerialName("sort_title") val sortTitle: String? = null,
    @SerialName("poster_path") @JsonNames("media_poster") val posterPath: String? = null,
    @SerialName("backdrop_path") val backdropPath: String? = null,
    val poster: ImageResourceDto? = null,
    val backdrop: ImageResourceDto? = null,
    @SerialName("media_type") val mediaType: String? = null,
    val type: String? = null,
    val year: Int? = null,
    val rating: Float? = null,
    val genres: List<String>? = null,
    val overview: String? = null,
    @SerialName("series_id") val seriesId: Int? = null,
    @SerialName("release_date") val releaseDate: String? = null,
    val position: Float? = null,
    val duration: Float? = null,
    val completed: Boolean? = null,
)

@Serializable
data class MediaDto(
    val id: Int,
    @SerialName("library_id") val libraryId: Int? = null,
    @SerialName("media_type") val mediaType: String,
    val title: String,
    @SerialName("sort_title") val sortTitle: String? = null,
    @SerialName("tmdb_id") val tmdbId: Int? = null,
    @SerialName("imdb_id") val imdbId: String? = null,
    val overview: String? = null,
    val tagline: String? = null,
    @SerialName("release_date") val releaseDate: String? = null,
    val rating: Float? = null,
    @SerialName("imdb_rating") val imdbRating: Float? = null,
    @SerialName("poster_path") val posterPath: String? = null,
    @SerialName("backdrop_path") val backdropPath: String? = null,
    @SerialName("logo_path") val logoPath: String? = null,
    @SerialName("thumb_path") val thumbPath: String? = null,
    val poster: ImageResourceDto? = null,
    val backdrop: ImageResourceDto? = null,
    val logo: ImageResourceDto? = null,
    val thumb: ImageResourceDto? = null,
    @SerialName("metadata_locked") val metadataLocked: Boolean = false,
    val duration: Float? = null,
    @SerialName("series_id") val seriesId: Int? = null,
    @SerialName("season_id") val seasonId: Int? = null,
    @SerialName("episode_number") val episodeNumber: Int? = null,
    @SerialName("season_number") val seasonNumber: Int? = null,
    val genres: List<GenreDto>? = null,
    val credits: CreditsDto? = null,
    @SerialName("similar") val similar: List<MediaListItemDto>? = null,
)

@Serializable
data class MediaWithFilesDto(
    val media: MediaDto? = null,
    val files: List<MediaFileDto> = emptyList(),
    @SerialName("series_id") val seriesId: Int? = null,
    @SerialName("season_id") val seasonId: Int? = null,
    @SerialName("episode_number") val episodeNumber: Int? = null,
    @SerialName("season_number") val seasonNumber: Int? = null,
)

@Serializable
data class MediaFileDto(
    val id: Int,
    @SerialName("media_id") val mediaId: Int,
    @SerialName("file_path") val filePath: String,
    @SerialName("file_size") val fileSize: Long,
    val duration: Float? = null,
    val width: Int? = null,
    val height: Int? = null,
    @SerialName("video_codec") val videoCodec: String? = null,
    @SerialName("audio_codec") val audioCodec: String? = null,
    val container: String? = null,
    val bitrate: Int? = null,
    @SerialName("is_primary") val isPrimary: Boolean = false,
)

@Serializable
data class GenreDto(
    val id: Int,
    val name: String,
    @SerialName("tmdb_id") val tmdbId: Int? = null,
)

@Serializable
data class CreditsDto(
    val cast: List<CastMemberDto>? = null,
    val crew: List<CrewMemberDto>? = null,
)

@Serializable
data class CastMemberDto(
    val id: Int,
    val name: String,
    val character: String? = null,
    @SerialName("profile_path") val profilePath: String? = null,
    val order: Int? = null,
)

@Serializable
data class CrewMemberDto(
    val id: Int,
    val name: String,
    val job: String? = null,
    val department: String? = null,
    @SerialName("profile_path") val profilePath: String? = null,
)

@Serializable
data class BrowseResponse(
    val folders: List<FolderDto>? = null,
    val media: List<MediaListItemDto>? = null,
)

@Serializable
data class FolderDto(
    val name: String,
    val path: String,
    val type: String? = null,
)

@Serializable
data class LibraryDto(
    val id: Int,
    val name: String,
    val type: String,
    val paths: List<String>,
    @SerialName("item_count") val itemCount: Int? = null,
    @SerialName("last_scanned") val lastScanned: String? = null,
)

@Serializable
data class SearchResponse(
    val movies: List<MediaListItemDto>? = null,
    val series: List<SeriesListItemDto>? = null,
)
