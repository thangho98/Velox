package com.velox.app.data.api

import com.velox.app.domain.model.ImageResource
import kotlinx.serialization.Serializable

/**
 * Lightweight serializable wrapper for persisting [ImageResource] in DataStore.
 * Only stores the fields needed for avatar display — keeps the stored JSON small.
 */
@Serializable
data class ProfileImageJson(
    val url: String,
    val srcset: Map<String, String> = emptyMap(),
    val type: String = "profile",
    val aspect: String = "1:1",
    val width: Int? = null,
    val height: Int? = null,
    val blurhash: String? = null,
) {
    fun toImageResource() = ImageResource(
        url = url,
        srcset = srcset,
        type = type,
        aspect = aspect,
        width = width,
        height = height,
        blurhash = blurhash,
    )

    companion object {
        fun fromImageResource(img: ImageResource) = ProfileImageJson(
            url = img.url,
            srcset = img.srcset,
            type = img.type,
            aspect = img.aspect,
            width = img.width,
            height = img.height,
            blurhash = img.blurhash,
        )
    }
}
