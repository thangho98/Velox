package com.velox.app.data.dto.livetv

import androidx.annotation.Keep
import com.velox.app.domain.model.livetv.LiveChannel
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Keep
@Serializable
data class LiveChannelDto(
    @SerialName("id") val id: Int,
    @SerialName("name") val name: String,
    @SerialName("logo") val logo: String?,
    @SerialName("groupTitle") val groupTitle: String,
    @SerialName("streamUrl") val url: String,
    @SerialName("country") val country: String? = null,
    @SerialName("isHidden") val isHidden: Boolean = false,
) {
    fun toDomain(): LiveChannel {
        return LiveChannel(
            id = id,
            name = name,
            logo = logo,
            groupTitle = groupTitle,
            url = url,
            country = country.orEmpty(),
            isHidden = isHidden,
        )
    }
}

@Keep
@Serializable
data class ToggleHiddenResponse(
    @SerialName("success") val success: Boolean = true,
    @SerialName("is_hidden") val isHidden: Boolean,
)
