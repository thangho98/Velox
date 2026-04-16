package com.velox.app.data.model

import com.velox.app.domain.model.ImageResource
import com.velox.app.data.util.ImageUrlResolver
import kotlinx.serialization.Serializable

@Serializable
data class ImageResourceDto(
    val url: String,
    val srcset: Map<String, String>,
    val type: String,
    val aspect: String,
    val width: Int? = null,
    val height: Int? = null,
    val blurhash: String? = null,
)

fun ImageResourceDto.toDomain() = ImageResource(
    url = ImageUrlResolver.resolve(url) ?: url,
    srcset = srcset.mapValues { ImageUrlResolver.resolve(it.value) ?: it.value },
    type = type,
    aspect = aspect,
    width = width,
    height = height,
    blurhash = blurhash
)
