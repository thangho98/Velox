package com.velox.app.domain.model.livetv

data class LiveChannel(
    val id: Int,
    val name: String,
    val logo: String?,
    val groupTitle: String,
    val url: String,
    val country: String = "",
    val isHidden: Boolean = false,
)
