package com.velox.app.data.repository

import com.velox.app.data.api.VeloxApi
import com.velox.app.data.api.VeloxApiProvider
import com.velox.app.data.model.AutoDownloadSubtitleRequest
import com.velox.app.data.model.SubtitleDownloadRequest
import com.velox.app.data.model.TranslateSubtitleRequest
import com.velox.app.data.model.UpdateProgressRequest
import com.velox.app.data.util.ImageUrlResolver
import com.velox.app.data.model.toDomain
import com.velox.app.domain.model.*
import com.velox.app.domain.repository.MediaRepository
import com.velox.app.domain.repository.SubtitleSearchResult
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class MediaRepositoryImpl @Inject constructor(
    private val apiProvider: VeloxApiProvider,
    private val serverPrefsManager: com.velox.app.data.api.ServerPrefsManager,
) : MediaRepository {

    private val api: VeloxApi
        get() = apiProvider.getApi()

    init {
        ImageUrlResolver.setBaseUrlProvider { serverPrefsManager.getServerUrlSync() }
    }

    private fun getFullUrl(path: String?): String? = ImageUrlResolver.resolve(path)
    private fun getRequiredFullUrl(path: String): String = ImageUrlResolver.resolve(path) ?: path
    private fun getBackdropUrl(path: String?): String? = ImageUrlResolver.resolve(path, "w1280")

    override suspend fun getMediaList(
        type: String?,
        genre: String?,
        year: String?,
        minRating: Float?,
        sort: String?,
        search: String?,
        libraryId: Int?,
        limit: Int,
        offset: Int,
        startChar: String?,
    ): Result<List<MediaItem>> {
        return try {
            val response = api.getMediaList(
                type = type,
                genre = genre,
                year = year,
                minRating = minRating,
                sort = sort,
                search = search,
                libraryId = libraryId,
                limit = limit,
                offset = offset,
                startChar = startChar,
            )
            if (response.isSuccessful) {
                val items = response.body()?.data?.map { dto ->
                    MediaItem(
                        id = dto.id,
                        title = dto.title,
                        posterPath = getFullUrl(dto.posterPath),
                        backdropPath = getBackdropUrl(dto.backdropPath),
                        poster = dto.poster?.toDomain(),
                        backdrop = dto.backdrop?.toDomain(),
                        year = dto.year,
                        rating = dto.rating,
                        mediaType = dto.mediaType ?: dto.type ?: "movie",
                        overview = dto.overview,
                        position = dto.position,
                        duration = dto.duration,
                        completed = dto.completed,
                    )
                } ?: emptyList()
                Result.success(items)
            } else {
                Result.failure(Exception("Failed to fetch media list"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getMediaAlphabet(
        type: String?,
        genre: String?,
        year: String?,
        libraryId: Int?
    ): Result<List<AlphabetCount>> {
        return try {
            val response = api.getMediaAlphabet(
                type = type,
                genre = genre,
                year = year,
                libraryId = libraryId,
            )
            if (response.isSuccessful) {
                val items = response.body()?.data?.map { dto ->
                    AlphabetCount(letter = dto.letter, count = dto.count)
                } ?: emptyList()
                Result.success(items)
            } else {
                Result.failure(Exception("Failed to fetch media alphabet"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getMedia(mediaId: Int): Result<MediaDetail> {
        return try {
            val response = api.getMedia(mediaId)
            if (response.isSuccessful) {
                val dto = response.body()?.data!!
                Result.success(
                    MediaDetail(
                        id = dto.id,
                        title = dto.title,
                        overview = dto.overview,
                        posterPath = getFullUrl(dto.posterPath),
                        backdropPath = getBackdropUrl(dto.backdropPath),
                        poster = dto.poster?.toDomain(),
                        backdrop = dto.backdrop?.toDomain(),
                        logo = dto.logo?.toDomain(),
                        thumb = dto.thumb?.toDomain(),
                        rating = dto.rating,
                        duration = dto.duration,
                        releaseDate = dto.releaseDate,
                        mediaType = dto.mediaType,
                        seriesId = dto.seriesId,
                        seasonId = dto.seasonId,
                        genres = dto.genres?.map { Genre(it.id, it.name) } ?: emptyList(),
                        credits = dto.credits?.let { credits ->
                            Credits(
                                cast = credits.cast?.map { cast ->
                                    CastMember(
                                        id = cast.id,
                                        name = cast.name,
                                        character = cast.character,
                                        profilePath = getFullUrl(cast.profilePath),
                                        order = cast.order,
                                    )
                                } ?: emptyList(),
                                crew = credits.crew?.map { crew ->
                                    CrewMember(
                                        id = crew.id,
                                        name = crew.name,
                                        job = crew.job,
                                        department = crew.department,
                                        profilePath = getFullUrl(crew.profilePath),
                                    )
                                } ?: emptyList(),
                            )
                        },
                        similar = dto.similar?.map { similar ->
                            MediaItem(
                                id = similar.id,
                                title = similar.title,
                                posterPath = getFullUrl(similar.posterPath),
                                backdropPath = getBackdropUrl(similar.backdropPath),
                                poster = similar.poster?.toDomain(),
                                backdrop = similar.backdrop?.toDomain(),
                                year = similar.year,
                                rating = similar.rating,
                                mediaType = similar.mediaType ?: similar.type ?: "movie",
                                overview = similar.overview,
                            )
                        } ?: emptyList(),
                    ),
                )
            } else {
                Result.failure(Exception("Failed to fetch media details"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getSeriesList(
        genre: String?,
        year: String?,
        sort: String?,
        search: String?,
        libraryId: Int?,
        limit: Int,
        offset: Int,
        startChar: String?,
    ): Result<List<SeriesItem>> {
        return try {
            val response = api.getSeriesList(
                genre = genre,
                year = year,
                sort = sort,
                search = search,
                libraryId = libraryId,
                limit = limit,
                offset = offset,
                startChar = startChar,
            )
            if (response.isSuccessful) {
                val items = response.body()?.data?.map { dto ->
                    SeriesItem(
                        id = dto.id,
                        title = dto.title,
                        posterPath = getFullUrl(dto.posterPath),
                        backdropPath = getBackdropUrl(dto.backdropPath),
                        poster = dto.poster?.toDomain(),
                        backdrop = dto.backdrop?.toDomain(),
                        year = dto.firstAirDate?.take(4)?.toIntOrNull() ?: dto.year,
                        rating = dto.rating,
                        overview = dto.overview,
                        seasonCount = dto.seasonCount,
                        episodeCount = dto.episodeCount,
                    )
                } ?: emptyList()
                Result.success(items)
            } else {
                Result.failure(Exception("Failed to fetch series list"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getSeriesAlphabet(
        genre: String?,
        year: String?,
        libraryId: Int?
    ): Result<List<AlphabetCount>> {
        return try {
            val response = api.getSeriesAlphabet(
                genre = genre,
                year = year,
                libraryId = libraryId,
            )
            if (response.isSuccessful) {
                val items = response.body()?.data?.map { dto ->
                    AlphabetCount(letter = dto.letter, count = dto.count)
                } ?: emptyList()
                Result.success(items)
            } else {
                Result.failure(Exception("Failed to fetch series alphabet"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getSeries(seriesId: Int): Result<SeriesDetail> {
        return try {
            val response = api.getSeries(seriesId)
            val seasonsResponse = api.getSeasons(seriesId)
            if (response.isSuccessful) {
                val dto = response.body()?.data!!
                val seasonsList = if (seasonsResponse.isSuccessful) seasonsResponse.body()?.data else dto.seasons
                Result.success(
                    SeriesDetail(
                        id = dto.id,
                        title = dto.title,
                        overview = dto.overview,
                        posterPath = getFullUrl(dto.posterPath),
                        backdropPath = getBackdropUrl(dto.backdropPath),
                        poster = dto.poster?.toDomain(),
                        backdrop = dto.backdrop?.toDomain(),
                        logo = dto.logo?.toDomain(),
                        thumb = dto.thumb?.toDomain(),
                        status = dto.status,
                        network = dto.network,
                        firstAirDate = dto.firstAirDate,
                        genres = dto.genres?.map { Genre(it.id, it.name) } ?: emptyList(),
                        seasons = seasonsList?.map { season ->
                            Season(
                                id = season.id,
                                seriesId = season.seriesId,
                                seasonNumber = season.seasonNumber,
                                title = season.title ?: "Season ${season.seasonNumber}",
                                overview = season.overview,
                                posterPath = getFullUrl(season.posterPath),
                                episodeCount = season.episodeCount,
                            )
                        } ?: emptyList(),
                        credits = dto.credits?.let { credits ->
                            Credits(
                                cast = credits.cast?.map { cast ->
                                    CastMember(
                                        id = cast.id,
                                        name = cast.name,
                                        character = cast.character,
                                        profilePath = getFullUrl(cast.profilePath),
                                        order = cast.order,
                                    )
                                } ?: emptyList(),
                                crew = credits.crew?.map { crew ->
                                    CrewMember(
                                        id = crew.id,
                                        name = crew.name,
                                        job = crew.job,
                                        department = crew.department,
                                        profilePath = getFullUrl(crew.profilePath),
                                    )
                                } ?: emptyList(),
                            )
                        },
                        metadataLocked = dto.metadataLocked,
                    ),
                )
            } else {
                Result.failure(Exception("Failed to fetch series details"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getSeasons(seriesId: Int): Result<List<Season>> {
        return try {
            val response = api.getSeasons(seriesId)
            if (response.isSuccessful) {
                val seasons = response.body()?.data?.map { dto ->
                    Season(
                        id = dto.id,
                        seriesId = dto.seriesId,
                        seasonNumber = dto.seasonNumber,
                        title = dto.title ?: "Season ${dto.seasonNumber}",
                        overview = dto.overview,
                        posterPath = getFullUrl(dto.posterPath),
                        episodeCount = dto.episodeCount,
                    )
                } ?: emptyList()
                Result.success(seasons)
            } else {
                Result.failure(Exception("Failed to fetch seasons"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getEpisodes(seriesId: Int, seasonId: Int): Result<List<Episode>> {
        return try {
            val response = api.getEpisodes(seriesId, seasonId)
            if (response.isSuccessful) {
                val episodes = response.body()?.data?.map { dto ->
                    Episode(
                        id = dto.id,
                        mediaId = dto.mediaId ?: dto.id,  // Use mediaId for playback, fallback to episode id
                        seriesId = dto.seriesId,
                        seasonId = dto.seasonId,
                        seasonNumber = dto.seasonNumber,
                        episodeNumber = dto.episodeNumber,
                        title = dto.title,
                        overview = dto.overview,
                        stillPath = getFullUrl(dto.stillPath),
                        still = dto.still?.toDomain(),
                        airDate = dto.airDate,
                        duration = dto.duration,
                    )
                } ?: emptyList()
                Result.success(episodes)
            } else {
                Result.failure(Exception("Failed to fetch episodes"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun browse(libraryId: Int?, path: String?): Result<List<BrowseItem>> {
        return try {
            val response = api.browse(libraryId, path)
            if (response.isSuccessful) {
                val browseResponse = response.body()?.data!!
                val items = mutableListOf<BrowseItem>()

                browseResponse.folders?.forEach { folder ->
                    items.add(
                        BrowseItem(
                            name = folder.name,
                            path = folder.path,
                            type = folder.type,
                            isFolder = true,
                        ),
                    )
                }

                browseResponse.media?.forEach { media ->
                    items.add(
                        BrowseItem(
                            name = media.title,
                            path = "",
                            type = media.type ?: media.mediaType,
                            isFolder = false,
                            mediaId = media.id,
                            mediaType = media.type ?: media.mediaType,
                            posterPath = getFullUrl(media.posterPath),
                            poster = media.poster?.toDomain(),
                            backdropPath = getBackdropUrl(media.backdropPath),
                            backdrop = media.backdrop?.toDomain(),
                        ),
                    )
                }

                Result.success(items)
            } else {
                Result.failure(Exception("Failed to browse"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getLibraries(): Result<List<Library>> {
        return try {
            val response = api.getLibraries()
            if (response.isSuccessful) {
                val libraries = response.body()?.data?.map { dto ->
                    Library(
                        id = dto.id,
                        name = dto.name,
                        type = dto.type,
                        paths = dto.paths,
                        itemCount = dto.itemCount,
                    )
                } ?: emptyList()
                Result.success(libraries)
            } else {
                Result.failure(Exception("Failed to fetch libraries"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getGenres(type: String?): Result<List<Genre>> {
        return try {
            val response = api.getGenres(type)
            if (response.isSuccessful) {
                val genres = response.body()?.data?.map { dto ->
                    Genre(id = dto.id, name = dto.name)
                } ?: emptyList()
                Result.success(genres)
            } else {
                Result.failure(Exception("Failed to fetch genres"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun search(query: String, limit: Int): Result<Pair<List<MediaItem>, List<SeriesItem>>> {
        return try {
            val response = api.search(query, limit)
            if (response.isSuccessful) {
                val searchResponse = response.body()?.data!!
                val movies = searchResponse.movies?.map { dto ->
                    MediaItem(
                        id = dto.id,
                        title = dto.title,
                        posterPath = getFullUrl(dto.posterPath),
                        backdropPath = getBackdropUrl(dto.backdropPath),
                        poster = dto.poster?.toDomain(),
                        backdrop = dto.backdrop?.toDomain(),
                        year = dto.year,
                        rating = dto.rating,
                        mediaType = "movie",
                        overview = dto.overview,
                    )
                } ?: emptyList()

                val series = searchResponse.series?.map { dto ->
                    SeriesItem(
                        id = dto.id,
                        title = dto.title,
                        posterPath = getFullUrl(dto.posterPath),
                        backdropPath = getBackdropUrl(dto.backdropPath),
                        poster = dto.poster?.toDomain(),
                        backdrop = dto.backdrop?.toDomain(),
                        year = dto.year,
                        rating = dto.rating,
                        overview = dto.overview,
                        seasonCount = dto.seasonCount,
                        episodeCount = dto.episodeCount,
                    )
                } ?: emptyList()

                Result.success(Pair(movies, series))
            } else {
                Result.failure(Exception("Search failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getContinueWatching(limit: Int): Result<List<ContinueWatchingItem>> {
        return try {
            val response = api.getContinueWatching(limit)
            if (response.isSuccessful) {
                val items = response.body()?.data?.map { dto ->
                    ContinueWatchingItem(
                        mediaId = dto.mediaId,
                        seriesId = dto.seriesId,
                        position = dto.position,
                        completed = dto.completed,
                        title = dto.title,
                        posterPath = getFullUrl(dto.posterPath),
                        backdropPath = getBackdropUrl(dto.backdropPath),
                        poster = dto.poster?.toDomain(),
                        backdrop = dto.backdrop?.toDomain(),
                        mediaType = dto.mediaType,
                        duration = dto.duration,
                        seriesTitle = dto.seriesTitle,
                        seasonNumber = dto.seasonNumber,
                        episodeNumber = dto.episodeNumber,
                    )
                } ?: emptyList()
                Result.success(items)
            } else {
                Result.failure(Exception("Failed to fetch continue watching"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getNextUp(limit: Int): Result<List<NextUpItem>> {
        return try {
            val response = api.getNextUp(limit)
            if (response.isSuccessful) {
                val items = response.body()?.data?.map { dto ->
                    NextUpItem(
                        mediaId = dto.mediaId,
                        seriesId = dto.seriesId,
                        title = dto.title,
                        episodeTitle = dto.episodeTitle,
                        stillPath = getFullUrl(dto.stillPath),
                        backdropPath = getBackdropUrl(dto.backdropPath),
                        still = dto.still?.toDomain(),
                        backdrop = dto.backdrop?.toDomain(),
                        poster = dto.poster?.toDomain(),
                        duration = dto.duration,
                        seasonNumber = dto.seasonNumber,
                        episodeNumber = dto.episodeNumber,
                        seriesTitle = dto.seriesTitle,
                        seriesPoster = getFullUrl(dto.seriesPoster),
                    )
                } ?: emptyList()
                Result.success(items)
            } else {
                Result.failure(Exception("Failed to fetch next up"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getRecentlyWatched(limit: Int): Result<List<MediaItem>> {
        return try {
            val response = api.getRecentlyWatched(limit)
            if (response.isSuccessful) {
                val items = response.body()?.data?.map { dto ->
                    MediaItem(
                        id = dto.id,
                        title = dto.title,
                        posterPath = getFullUrl(dto.posterPath),
                        backdropPath = getBackdropUrl(dto.backdropPath),
                        poster = dto.poster?.toDomain(),
                        backdrop = dto.backdrop?.toDomain(),
                        year = dto.year,
                        rating = dto.rating,
                        mediaType = dto.mediaType ?: dto.type ?: "movie",
                        overview = dto.overview,
                    )
                } ?: emptyList()
                Result.success(items)
            } else {
                Result.failure(Exception("Failed to fetch recently watched"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getFavorites(limit: Int, offset: Int): Result<List<MediaItem>> {
        return try {
            val response = api.getFavorites(limit, offset)
            if (response.isSuccessful) {
                val items = response.body()?.data?.map { dto ->
                    MediaItem(
                        id = dto.id,
                        title = dto.title,
                        posterPath = getFullUrl(dto.posterPath),
                        backdropPath = getBackdropUrl(dto.backdropPath),
                        poster = dto.poster?.toDomain(),
                        backdrop = dto.backdrop?.toDomain(),
                        year = dto.year,
                        rating = dto.rating,
                        mediaType = dto.mediaType ?: dto.type ?: "movie",
                        overview = dto.overview,
                    )
                } ?: emptyList()
                Result.success(items)
            } else {
                Result.failure(Exception("Failed to fetch favorites"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun toggleFavorite(mediaId: Int): Result<Boolean> {
        return try {
            val response = api.toggleFavorite(mediaId)
            if (response.isSuccessful) {
                Result.success(response.body()?.data?.isFavorite ?: false)
            } else {
                Result.failure(Exception("Failed to toggle favorite"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun updateProgress(mediaId: Int, position: Float, completed: Boolean): Result<Unit> {
        return try {
            val response = api.updateProgress(mediaId, UpdateProgressRequest(position, completed))
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to update progress"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getPlaybackInfo(mediaId: Int, params: com.velox.app.domain.model.PlaybackInfoParams?): Result<PlaybackInfo> {
        return try {
            val request = params?.let {
                com.velox.app.data.model.PlaybackInfoRequest(
                    videoCodecs = it.videoCodecs,
                    audioCodecs = it.audioCodecs,
                    containers = it.containers,
                    maxHeight = it.maxHeight,
                    mediaFileId = it.mediaFileId,
                    selectedAudioTrack = it.selectedAudioTrack,
                    selectedSubtitle = it.selectedSubtitle,
                    selectedSubtitleId = it.selectedSubtitleId,
                )
            }
            val response = api.getPlaybackInfo(mediaId, request)
            if (response.isSuccessful) {
                val dto = response.body()?.data!!
                Result.success(
                    PlaybackInfo(
                        mediaId = dto.mediaId,
                        primaryFileId = dto.primaryFileId,
                        method = dto.method,
                        streamUrl = getRequiredFullUrl(dto.streamUrl),
                        directUrl = getRequiredFullUrl(dto.directUrl),
                        abrUrl = getFullUrl(dto.abrUrl),
                        // HLS / Pretranscode fields
                        pretranscodeUrl = getFullUrl(dto.pretranscodeUrl),
                        hlsUrl = getRequiredFullUrl(dto.hlsUrl),
                        prefer = dto.prefer,
                        streamSessionId = dto.streamSessionId,
                        // Pretranscode metadata
                        ptVideoCodec = dto.ptVideoCodec,
                        ptAudioCodec = dto.ptAudioCodec,
                        ptHeight = dto.ptHeight,
                        ptVideoBitrate = dto.ptVideoBitrate,
                        ptAudioBitrate = dto.ptAudioBitrate,
                        // Media details
                        videoCodec = dto.videoCodec,
                        videoProfile = dto.videoProfile,
                        videoLevel = dto.videoLevel,
                        videoFps = dto.videoFps,
                        audioCodec = dto.audioCodec,
                        container = dto.container,
                        fileSize = dto.fileSize,
                        bitrate = dto.bitrate,
                        decisionReason = dto.decisionReason,
                        estimatedBitrate = dto.estimatedBitrate,
                        //
                        position = dto.position,
                        duration = dto.duration,
                        width = dto.width,
                        height = dto.height,
                        audioTracks = dto.audioTracks?.map { track ->
                            AudioTrack(
                                id = track.id,
                                language = track.language,
                                label = track.label,
                                codec = track.codec,
                                channels = track.channels,
                                bitrate = track.bitrate,
                                sampleRate = track.sampleRate,
                                isDefault = track.isDefault,
                                selected = track.selected,
                            )
                        } ?: emptyList(),
                        subtitleTracks = dto.subtitleTracks?.map { track ->
                            SubtitleTrack(
                                id = track.id,
                                language = track.language,
                                label = track.label,
                                format = track.format,
                                isDefault = track.isDefault,
                                isImage = track.isImage,
                            )
                        } ?: emptyList(),
                        skipSegments = dto.skipSegments?.map { segment ->
                            SkipSegment(
                                type = segment.type,
                                start = segment.start,
                                end = segment.end,
                            )
                        } ?: emptyList(),
                        availableQualities = dto.availableQualities?.map { quality ->
                            QualityOption(
                                height = quality.height,
                                label = quality.label,
                                instant = quality.instant,
                            )
                        } ?: emptyList(),
                    ),
                )
            } else {
                Result.failure(Exception("Failed to get playback info"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getStreamUrl(mediaId: Int): Result<String> {
        return try {
            val response = api.getStreamUrl(mediaId)
            if (response.isSuccessful) {
                val streamResponse = response.body()?.data!!
                // Prefer direct URL, fall back to HLS URL
                val url = streamResponse.directUrl.ifEmpty { streamResponse.hlsUrl }
                Result.success(getFullUrl(url) ?: "")
            } else {
                Result.failure(Exception("Failed to get stream URL"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getStreamApiKey(mediaId: Int): Result<String> {
        return try {
            val response = api.getStreamUrl(mediaId)
            if (response.isSuccessful) {
                val streamResponse = response.body()?.data
                    ?: return Result.failure(Exception("Empty stream URL response"))
                Result.success(streamResponse.apiKey)
            } else {
                Result.failure(Exception("Failed to get stream api key (${response.code()})"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getProgress(mediaId: Int): Result<WatchProgress?> {
        return try {
            val response = api.getProgress(mediaId)
            if (response.isSuccessful) {
                val dto = response.body()?.data
                if (dto != null) {
                    Result.success(
                        WatchProgress(
                            mediaId = dto.mediaId,
                            position = dto.position,
                            completed = dto.completed,
                            isFavorite = dto.isFavorite,
                            duration = dto.mediaDuration,
                        ),
                    )
                } else {
                    Result.success(null)
                }
            } else {
                Result.failure(Exception("Failed to get progress"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun dismissProgress(mediaId: Int): Result<Unit> {
        return try {
            val response = api.dismissProgress(mediaId)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to dismiss progress"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getMediaFiles(mediaId: Int): Result<List<MediaFile>> {
        return try {
            val response = api.getMediaWithFiles(mediaId)
            if (response.isSuccessful) {
                val files = response.body()?.data?.files?.map { dto ->
                    MediaFile(
                        id = dto.id,
                        mediaId = dto.mediaId,
                        filePath = dto.filePath,
                        fileSize = dto.fileSize,
                        duration = dto.duration,
                        width = dto.width,
                        height = dto.height,
                        videoCodec = dto.videoCodec,
                        audioCodec = dto.audioCodec,
                        container = dto.container,
                        bitrate = dto.bitrate,
                        isPrimary = dto.isPrimary,
                    )
                } ?: emptyList()
                Result.success(files)
            } else {
                Result.failure(Exception("Failed to get media files"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getMediaWithFilesInfo(mediaId: Int): Result<MediaWithFilesInfo> {
        return try {
            val response = api.getMediaWithFiles(mediaId)
            if (response.isSuccessful) {
                val dto = response.body()?.data!!
                Result.success(
                    MediaWithFilesInfo(
                        mediaType = dto.media?.mediaType ?: "movie",
                        title = dto.media?.title ?: "",
                        seriesId = dto.seriesId,
                        seasonId = dto.seasonId,
                        seasonNumber = dto.episodeNumber?.let { dto.seasonNumber },
                        episodeNumber = dto.episodeNumber,
                        episodeOverview = dto.media?.overview,
                    )
                )
            } else {
                Result.failure(Exception("Failed to get media with files"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    // Subtitles
    override suspend fun searchSubtitles(mediaId: Int, lang: String): Result<List<SubtitleSearchResult>> {
        return try {
            val response = api.searchSubtitles(mediaId, lang)
            if (response.isSuccessful) {
                val items = response.body()?.data?.map { dto ->
                    SubtitleSearchResult(
                        provider = dto.provider,
                        externalId = dto.externalId,
                        title = dto.title,
                        language = dto.language,
                        hi = dto.hi,
                        aiTranslated = dto.aiTranslated,
                        format = dto.format,
                    )
                } ?: emptyList()
                Result.success(items)
            } else {
                Result.failure(Exception("Failed to search subtitles"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun downloadSubtitle(
        mediaId: Int,
        provider: String,
        externalId: String,
        language: String,
    ): Result<Unit> {
        return try {
            val request = SubtitleDownloadRequest(
                provider = provider,
                externalId = externalId,
                language = language,
            )
            val response = api.downloadSubtitle(mediaId, request)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to download subtitle"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun autoDownloadSubtitle(mediaId: Int): Result<Unit> {
        return try {
            val response = api.autoDownloadSubtitle(mediaId)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to auto download subtitle"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun translateSubtitle(subtitleId: Int, targetLanguage: String): Result<Unit> {
        return try {
            val request = TranslateSubtitleRequest(
                targetLanguage = targetLanguage,
            )
            val response = api.translateSubtitle(subtitleId, request)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to translate subtitle"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getSubtitleContent(fileId: Int, subId: Int): Result<String> {
        return try {
            val response = api.getSubtitleContent(fileId, subId)
            if (response.isSuccessful) {
                val content = response.body()?.string() ?: ""
                Result.success(content)
            } else {
                Result.failure(Exception("Failed to fetch subtitle content (code: ${response.code()})"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
