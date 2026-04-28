package com.velox.app.data.repository.livetv

import com.velox.app.data.api.VeloxApi
import com.velox.app.data.repository.safeApiCall
import com.velox.app.domain.model.DataResult
import com.velox.app.domain.model.livetv.LiveChannel
import com.velox.app.domain.model.livetv.LiveGroup
import com.velox.app.domain.model.livetv.LiveProgram
import com.velox.app.domain.repository.livetv.LiveTvRepository
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class LiveTvRepositoryImpl @Inject constructor(
    private val api: VeloxApi
) : LiveTvRepository {

    override fun getLiveTvGroups(): Flow<DataResult<List<LiveGroup>>> = flow {
        val result = safeApiCall(
            apiCall = { api.getLiveTvGroups() },
            transform = { list -> list.map { LiveGroup(groupTitle = it, count = 0) } }
        )
        emit(result)
    }

    override fun getLiveTvChannels(group: String?): Flow<DataResult<List<LiveChannel>>> = flow {
        val result = safeApiCall(
            apiCall = { api.getLiveTvChannels(group = group) },
            transform = { dtoList -> dtoList.map { it.toDomain() } }
        )
        emit(result)
    }

    override fun getLiveTvRecentChannels(limit: Int): Flow<DataResult<List<LiveChannel>>> = flow {
        val result = safeApiCall(
            apiCall = { api.getLiveTvRecentChannels(limit = limit) },
            transform = { dtoList -> dtoList.map { it.toDomain() } }
        )
        emit(result)
    }

    override fun getLiveTvChannel(channelId: Int): Flow<DataResult<LiveChannel>> = flow {
        val result = safeApiCall(
            apiCall = { api.getLiveTvChannel(channelId) },
            transform = { it.toDomain() }
        )
        emit(result)
    }

    override fun getLiveTvEpg(channelId: Int): Flow<DataResult<List<LiveProgram>>> = flow {
        val result = safeApiCall(
            apiCall = { api.getLiveTvEpg(channelId) },
            transform = { dtoList -> dtoList.map { it.toDomain() } }
        )
        emit(result)
    }

    override suspend fun toggleChannelHidden(channelId: Int): DataResult<Boolean> {
        return safeApiCall(
            apiCall = { api.toggleLiveTvChannelHidden(channelId) },
            transform = { it.isHidden }
        )
    }
}
