package com.velox.app.domain.repository

import com.velox.app.domain.model.*
import kotlinx.coroutines.flow.Flow

data class SubtitleSearchResult(
    val provider: String,
    val externalId: String,
    val title: String,
    val language: String,
    val hi: Boolean = false,
    val aiTranslated: Boolean = false,
    val format: String? = null,
)

interface MediaRepository {
    suspend fun getMediaList(
        type: String? = null,
        genre: String? = null,
        year: String? = null,
        minRating: Float? = null,
        sort: String? = null,
        search: String? = null,
        libraryId: Int? = null,
        limit: Int = 50,
        offset: Int = 0,
        startChar: String? = null,
    ): Result<List<MediaItem>>

    suspend fun getMediaAlphabet(
        type: String? = null,
        genre: String? = null,
        year: String? = null,
        libraryId: Int? = null,
    ): Result<List<AlphabetCount>>

    suspend fun getMedia(mediaId: Int): Result<MediaDetail>
    suspend fun getSeriesList(
        genre: String? = null,
        year: String? = null,
        sort: String? = null,
        search: String? = null,
        libraryId: Int? = null,
        limit: Int = 50,
        offset: Int = 0,
        startChar: String? = null,
    ): Result<List<SeriesItem>>

    suspend fun getSeriesAlphabet(
        genre: String? = null,
        year: String? = null,
        libraryId: Int? = null,
    ): Result<List<AlphabetCount>>

    suspend fun getSeries(seriesId: Int): Result<SeriesDetail>
    suspend fun getSeasons(seriesId: Int): Result<List<Season>>
    suspend fun getEpisodes(seriesId: Int, seasonId: Int): Result<List<Episode>>
    suspend fun browse(libraryId: Int?, path: String?): Result<List<BrowseItem>>
    suspend fun getLibraries(): Result<List<Library>>
    suspend fun getGenres(type: String? = null): Result<List<Genre>>
    suspend fun search(query: String, limit: Int = 20): Result<Pair<List<MediaItem>, List<SeriesItem>>>
    suspend fun getContinueWatching(limit: Int = 20): Result<List<ContinueWatchingItem>>
    suspend fun getNextUp(limit: Int = 20): Result<List<NextUpItem>>
    suspend fun getRecentlyWatched(limit: Int = 20): Result<List<MediaItem>>
    suspend fun getFavorites(limit: Int = 20, offset: Int = 0): Result<List<MediaItem>>
    suspend fun toggleFavorite(mediaId: Int): Result<Boolean>
    suspend fun updateProgress(mediaId: Int, position: Float, completed: Boolean): Result<Unit>
    suspend fun getPlaybackInfo(mediaId: Int, params: PlaybackInfoParams? = null): Result<PlaybackInfo>
    suspend fun getStreamUrl(mediaId: Int): Result<String>
    suspend fun getStreamApiKey(mediaId: Int): Result<String>
    suspend fun getProgress(mediaId: Int): Result<WatchProgress?>
    suspend fun dismissProgress(mediaId: Int): Result<Unit>
    suspend fun getMediaFiles(mediaId: Int): Result<List<MediaFile>>
    suspend fun getMediaWithFilesInfo(mediaId: Int): Result<MediaWithFilesInfo>

    // Subtitles
    suspend fun searchSubtitles(mediaId: Int, lang: String): Result<List<SubtitleSearchResult>>
    suspend fun downloadSubtitle(mediaId: Int, provider: String, externalId: String, language: String): Result<Unit>
    suspend fun autoDownloadSubtitle(mediaId: Int): Result<Unit>
    suspend fun translateSubtitle(subtitleId: Int, targetLanguage: String): Result<Unit>
    suspend fun getSubtitleContent(fileId: Int, subId: Int): Result<String>

    // Downloads
    suspend fun startDownload(mediaId: Int): Result<Unit>
    suspend fun startSeriesDownload(seriesId: Int): Result<Unit>
    suspend fun deleteDownload(mediaId: Int): Result<Unit>
    suspend fun deleteSeriesDownload(seriesId: Int): Result<Unit>
}
