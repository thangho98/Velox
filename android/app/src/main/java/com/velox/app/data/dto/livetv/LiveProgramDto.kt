package com.velox.app.data.dto.livetv

import androidx.annotation.Keep
import com.velox.app.domain.model.livetv.LiveProgram
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Keep
@Serializable
data class LiveProgramDto(
    @SerialName("id") val id: Int,
    @SerialName("channel_id") val channelId: Int,
    @SerialName("title") val title: String,
    @SerialName("description") val description: String? = null,
    @SerialName("start_time") val startTime: String,
    @SerialName("end_time") val endTime: String,
) {
    fun toDomain(): LiveProgram = LiveProgram(
        id = id,
        channelId = channelId,
        title = title,
        description = description.orEmpty(),
        startTimeIso = startTime,
        endTimeIso = endTime,
    )
}
