package com.velox.app.domain.model.livetv

data class LiveProgram(
    val id: Int,
    val channelId: Int,
    val title: String,
    val description: String,
    val startTimeIso: String,
    val endTimeIso: String,
)
