package com.velox.app.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class SubtitleSearchResultDto(
    @SerialName("provider") val provider: String,
    @SerialName("external_id") val externalId: String,
    @SerialName("title") val title: String,
    @SerialName("language") val language: String,
    @SerialName("hi") val hi: Boolean = false,
    @SerialName("ai_translated") val aiTranslated: Boolean = false,
    @SerialName("format") val format: String? = null,
)

@Serializable
data class SubtitleDownloadRequest(
    @SerialName("provider") val provider: String,
    @SerialName("external_id") val externalId: String,
    @SerialName("language") val language: String,
)

@Serializable
data class TranslateSubtitleRequest(
    @SerialName("target_language") val targetLanguage: String,
)

@Serializable
data class SubtitleDto(
    @SerialName("id") val id: Int,
    @SerialName("media_file_id") val mediaFileId: Int,
    @SerialName("title") val title: String?,
    @SerialName("language") val language: String,
    @SerialName("codec") val codec: String?,
    @SerialName("is_embedded") val isEmbedded: Boolean,
    @SerialName("is_forced") val isForced: Boolean,
    @SerialName("is_default") val isDefault: Boolean,
    @SerialName("is_sdh") val isSdh: Boolean,
    @SerialName("stream_index") val streamIndex: Int,
    @SerialName("file_path") val filePath: String?,
)
