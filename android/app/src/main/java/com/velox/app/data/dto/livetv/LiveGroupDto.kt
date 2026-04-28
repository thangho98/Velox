package com.velox.app.data.dto.livetv

import androidx.annotation.Keep
import com.velox.app.domain.model.livetv.LiveGroup
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Keep
@Serializable
data class LiveGroupDto(
    @SerialName("group_title") val groupTitle: String,
    @SerialName("count") val count: Int
) {
    fun toDomain(): LiveGroup {
        return LiveGroup(
            groupTitle = groupTitle,
            count = count
        )
    }
}
