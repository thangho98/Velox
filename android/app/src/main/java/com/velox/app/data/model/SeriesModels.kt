package com.velox.app.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class SeriesDto(
    val id: Int,
    @SerialName("library_id") val libraryId: Int? = null,
    val title: String,
    @SerialName("sort_title") val sortTitle: String? = null,
    @SerialName("tmdb_id") val tmdbId: Int? = null,
    val overview: String? = null,
    val status: String? = null,
    val network: String? = null,
    @SerialName("first_air_date") val firstAirDate: String? = null,
    @SerialName("poster_path") val posterPath: String? = null,
    @SerialName("backdrop_path") val backdropPath: String? = null,
    @SerialName("logo_path") val logoPath: String? = null,
    @SerialName("thumb_path") val thumbPath: String? = null,
    val poster: ImageResourceDto? = null,
    val backdrop: ImageResourceDto? = null,
    val logo: ImageResourceDto? = null,
    val thumb: ImageResourceDto? = null,
    @SerialName("metadata_locked") val metadataLocked: Boolean = false,
    val genres: List<GenreDto>? = null,
    val seasons: List<SeasonDto>? = null,
    @SerialName("credits") val credits: CreditsDto? = null,
)

@Serializable
data class SeriesListItemDto(
    val id: Int,
    val title: String,
    @SerialName("sort_title") val sortTitle: String? = null,
    @SerialName("poster_path") val posterPath: String? = null,
    @SerialName("backdrop_path") val backdropPath: String? = null,
    val poster: ImageResourceDto? = null,
    val backdrop: ImageResourceDto? = null,
    @SerialName("first_air_date") val firstAirDate: String? = null,
    val year: Int? = null,
    val rating: Float? = null,
    val genres: List<String>? = null,
    val overview: String? = null,
    @SerialName("season_count") val seasonCount: Int? = null,
    @SerialName("episode_count") val episodeCount: Int? = null,
)

@Serializable
data class SeasonDto(
    val id: Int,
    @SerialName("series_id") val seriesId: Int,
    @SerialName("season_number") val seasonNumber: Int,
    val title: String? = null,
    val overview: String? = null,
    @SerialName("poster_path") val posterPath: String? = null,
    @SerialName("air_date") val airDate: String? = null,
    @SerialName("episode_count") val episodeCount: Int? = null,
)

@Serializable
data class EpisodeDto(
    val id: Int,
    @SerialName("series_id") val seriesId: Int,
    @SerialName("season_id") val seasonId: Int,
    @SerialName("media_id") val mediaId: Int? = null,
    @SerialName("episode_number") val episodeNumber: Int,
    val title: String,
    val overview: String? = null,
    @SerialName("still_path") val stillPath: String? = null,
    val still: ImageResourceDto? = null,
    @SerialName("air_date") val airDate: String? = null,
    val duration: Float? = null,
    @SerialName("season_number") val seasonNumber: Int? = null,
    @SerialName("file_path") val filePath: String? = null,
    @SerialName("media_files") val mediaFiles: List<MediaFileDto>? = null,
    val credits: CreditsDto? = null,
)
