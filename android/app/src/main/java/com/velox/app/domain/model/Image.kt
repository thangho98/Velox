package com.velox.app.domain.model

data class ImageResource(
    val url: String,
    val srcset: Map<String, String>,
    val type: String,
    val aspect: String,
    val width: Int?,
    val height: Int?,
    val blurhash: String?,
)
