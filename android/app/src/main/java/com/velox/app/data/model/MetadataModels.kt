package com.velox.app.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class CinemaDto(
    val trailers: List<TrailerDto>? = null,
    val intros: List<IntroDto>? = null,
)

@Serializable
data class TrailerDto(
    val id: Int,
    val name: String,
    val url: String,
    val site: String, // "youtube", "vimeo"
    val key: String? = null, // YouTube video key
)

@Serializable
data class IntroDto(
    val id: Int,
    val name: String,
    @SerialName("file_path") val filePath: String,
    val duration: Float? = null,
    val start: Float? = null,
    val end: Float? = null,
)

@Serializable
data class UpdateMetadataRequest(
    val title: String? = null,
    val overview: String? = null,
    @SerialName("release_date") val releaseDate: String? = null,
    val genres: List<String>? = null,
    val rating: Float? = null,
)

@Serializable
data class UpdateImagesRequest(
    @SerialName("poster_path") val posterPath: String? = null,
    @SerialName("backdrop_path") val backdropPath: String? = null,
    @SerialName("logo_path") val logoPath: String? = null,
)
