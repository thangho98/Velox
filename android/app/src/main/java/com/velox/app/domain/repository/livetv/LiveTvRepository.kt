package com.velox.app.domain.repository.livetv

import com.velox.app.domain.model.DataResult
import com.velox.app.domain.model.livetv.LiveChannel
import com.velox.app.domain.model.livetv.LiveGroup
import com.velox.app.domain.model.livetv.LiveProgram
import kotlinx.coroutines.flow.Flow

interface LiveTvRepository {
    fun getLiveTvGroups(): Flow<DataResult<List<LiveGroup>>>

    fun getLiveTvChannels(group: String? = null): Flow<DataResult<List<LiveChannel>>>

    fun getLiveTvRecentChannels(limit: Int = 20): Flow<DataResult<List<LiveChannel>>>

    fun getLiveTvChannel(channelId: Int): Flow<DataResult<LiveChannel>>

    fun getLiveTvEpg(channelId: Int): Flow<DataResult<List<LiveProgram>>>

    suspend fun toggleChannelHidden(channelId: Int): DataResult<Boolean>
}
