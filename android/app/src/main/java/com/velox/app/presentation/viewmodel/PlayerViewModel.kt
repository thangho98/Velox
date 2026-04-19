package com.velox.app.presentation.viewmodel

import timber.log.Timber
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.media3.common.C
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import com.velox.app.domain.model.Episode as DomainEpisode
import com.velox.app.domain.model.MediaDetail as DomainMediaDetail
import com.velox.app.domain.model.MediaWithFilesInfo as DomainMediaWithFilesInfo
import com.velox.app.domain.model.Season as DomainSeason
import com.velox.app.domain.model.PlaybackInfo as DomainPlaybackInfo
import com.velox.app.domain.model.AudioTrack as DomainAudioTrack
import com.velox.app.domain.model.SubtitleTrack as DomainSubtitleTrack
import com.velox.app.domain.model.SkipSegment as DomainSkipSegment
import com.velox.app.domain.model.QualityOption as DomainQualityOption
import com.velox.app.domain.repository.MediaRepository
import com.velox.app.playback.PlaybackManager
import com.velox.app.playback.PlaybackMetadata
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.async
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class PlayerUiState(
    val isLoading: Boolean = true,
    val playbackInfo: DomainPlaybackInfo? = null,
    val mediaDetail: DomainMediaDetail? = null,
    val mediaContext: DomainMediaWithFilesInfo? = null,
    val isPlaying: Boolean = false,
    val currentPosition: Long = 0L,
    val duration: Long = 0L,
    val isBuffering: Boolean = false,
    val error: String? = null,
    val mediaTitle: String? = null,
    val videoUrl: String? = null,
    val audioTracks: List<AudioTrackUi> = emptyList(),
    val subtitleTracks: List<SubtitleTrackUi> = emptyList(),
    val selectedAudioTrackIndex: Int = 0,
    val selectedSubtitleIndex: Int = -1, // -1 = disabled
    val selectedSecondarySubtitleIndex: Int = -1, // -1 = disabled
    val skipSegments: List<SkipSegmentUi> = emptyList(),
    val showControls: Boolean = true,
    val playbackSpeed: Float = 1.0f,
    val isLocked: Boolean = false,
    val volume: Float = 1.0f,

    // Quality
    val maxQuality: Int = 0, // 0 = auto (no limit), >0 = max height in pixels
    val qualityOptions: List<QualityOptionUi> = emptyList(),

    // Aspect ratio
    val aspectRatio: AspectRatioUi = AspectRatioUi.Contain,

    // Repeat mode
    val repeatMode: RepeatModeUi = RepeatModeUi.None,

    // Subtitle delay
    val subtitleDelay: Float = 0f,

    // Subtitle appearance
    val subtitleSize: SubtitleSizeUi = SubtitleSizeUi.Large,
    val subtitleColor: String = "#ffffff",
    val subtitleBackground: SubtitleBackgroundUi = SubtitleBackgroundUi.None,

    // Playback info stats
    val showPlaybackStats: Boolean = false,
    val videoWidth: Int = 0,
    val videoHeight: Int = 0,

    // Dual subtitle support
    val primarySubtitleCues: List<SubtitleCueUi> = emptyList(),
    val secondarySubtitleCues: List<SubtitleCueUi> = emptyList(),
    val secondarySubtitleEnabled: Boolean = false,

    // Cinema overlay
    val cinemaItems: List<CinemaItemUi> = emptyList(),
    val showCinemaOverlay: Boolean = false,
    val cinemaIndex: Int = 0,

    // Seek feedback
    val showSeekFeedback: Boolean = false,
    val seekFeedbackSide: String = "right", // "left" or "right"
    val seekFeedbackAmount: Int = 10,

    // Subtitle search
    val showSubtitleSearch: Boolean = false,
    val subtitleSearchResults: List<SubtitleSearchResultUi> = emptyList(),
    val subtitleSearchLang: String = "en",
    val isSearchingSubtitles: Boolean = false,
    val isDownloadingSubtitle: Boolean = false,

    // Up next / next episode
    val showUpNext: Boolean = false,
    val upNextTitle: String? = null,
    val upNextId: Int? = null,
    val upNextCountdown: Int = 15,
    val nextEpisodeId: Int? = null,
    val nextEpisodeTitle: String? = null,
    val nextEpisodeWatched: Boolean = false,
    val nextNextEpisodeId: Int? = null,
    val showWatchedWarning: Boolean = false,

    // Watch detail panel
    val activeDetailPanel: DetailPanelUi = DetailPanelUi.None,
    val seasons: List<DomainSeason> = emptyList(),
    val seasonPanelSeasonId: Int = 0,
    val seasonPanelEpisodes: List<DomainEpisode> = emptyList(),
    val isSeasonPanelLoading: Boolean = false,
) {
    val effectiveDuration: Long
        get() {
            val apiDur = playbackInfo?.duration?.let { (it * 1000).toLong() } ?: 0L
            return maxOf(duration, apiDur)
        }
}

enum class DetailPanelUi {
    None, Info, Season
}

enum class AspectRatioUi {
    Contain, Cover, Fill
}

enum class RepeatModeUi {
    None, One, All
}

enum class SubtitleSizeUi {
    Small, Medium, Large
}

enum class SubtitleBackgroundUi {
    None, Semi, Solid
}

data class QualityOptionUi(
    val height: Int,
    val label: String,
    val instant: Boolean = false,
)

data class SubtitleCueUi(
    val start: Long, // milliseconds
    val end: Long,    // milliseconds
    val text: String,
)

data class CinemaItemUi(
    val type: String, // "intro" or "trailer"
    val title: String,
    val url: String,
    val skippable: Boolean = true,
)

data class AudioTrackUi(
    val index: Int,
    val label: String,
    val language: String,
    val codec: String,
    val channels: Int,
    val isDefault: Boolean,
)

data class SubtitleTrackUi(
    val id: Int,
    val index: Int,
    val label: String,
    val language: String,
    val format: String?,
)

data class SkipSegmentUi(
    val type: String,
    val start: Long,
    val end: Long,
)

data class SubtitleSearchResultUi(
    val id: String,
    val title: String,
    val lang: String,
    val hi: Boolean = false,
    val aiTranslated: Boolean = false,
)

interface PlayerActions {
    val player: androidx.media3.common.Player?
    val mediaId: Int
    val isTranslatingSubtitle: StateFlow<Boolean>

    fun loadPlaybackInfo()
    fun togglePlayPause()
    fun seekTo(positionMs: Long)
    fun seekForward(seconds: Int = 5)
    fun seekBackward(seconds: Int = 5)
    fun skipSegment(segment: SkipSegmentUi)
    fun selectAudioTrack(trackIndex: Int)
    fun selectSubtitleTrack(trackIndex: Int)
    fun selectSecondarySubtitleTrack(trackIndex: Int)
    fun toggleControls()
    fun toggleDetailPanel(panel: DetailPanelUi)
    fun closeDetailPanel()
    fun selectSeasonPanelSeason(seasonId: Int)
    fun setPlaybackSpeed(speed: Float)
    fun toggleLock()
    fun setVolume(volume: Float)
    fun setMaxQuality(height: Int)
    fun setAspectRatio(ratio: AspectRatioUi)
    fun setRepeatMode(mode: RepeatModeUi)
    fun cycleRepeatMode()
    fun setSubtitleDelay(delay: Float)
    fun adjustSubtitleDelay(delta: Float)
    fun setSubtitleSize(size: SubtitleSizeUi)
    fun setSubtitleColor(color: String)
    fun setSubtitleBackground(bg: SubtitleBackgroundUi)
    fun togglePlaybackStats()
    fun showSeekFeedback(side: String, amount: Int)
    fun hideSeekFeedback()
    fun loadCinemaItems(cinemaEnabled: Boolean = false)
    fun skipCinemaItem()
    fun skipAllCinema()
    fun enableSecondarySubtitle(enabled: Boolean)
    fun showUpNext(nextEpisodeId: Int, nextEpisodeTitle: String)
    fun dismissUpNext()
    fun getUpNextId(): Int?
    fun updatePosition()
    fun saveProgress(completed: Boolean = false)
    fun showSubtitleSearch()
    fun hideSubtitleSearch()
    fun searchSubtitles(lang: String)
    fun downloadSubtitle(provider: String, externalId: String, language: String)
    fun translateSubtitle(subtitleId: Int, targetLanguage: String)
    fun showWatchedWarning()
    fun hideWatchedWarning()
    fun resetProgressAndNavigate(targetMediaId: Int, navigate: (Int) -> Unit)
}

@dagger.hilt.android.lifecycle.HiltViewModel
class PlayerViewModel @Inject constructor(
    private val mediaRepository: MediaRepository,
    private val playerPrefsManager: com.velox.app.data.local.PlayerPrefsManager,
    private val playbackManager: PlaybackManager,
    savedStateHandle: SavedStateHandle,
) : ViewModel(), PlayerActions {

    override val mediaId: Int = checkNotNull(savedStateHandle["mediaId"])

    private val _uiState = MutableStateFlow(PlayerUiState())
    val uiState: StateFlow<PlayerUiState> = _uiState.asStateFlow()

    private var _player: ExoPlayer? = null
    override val player: ExoPlayer?
        get() = _player

    private var currentPlaybackSource: String? = null
    private var fallbackOrder: List<String> = emptyList()
    private var streamApiKey: String? = null

    private val _nextMediaRequested = MutableSharedFlow<Int>(extraBufferCapacity = 1)
    val nextMediaRequested: SharedFlow<Int> = _nextMediaRequested.asSharedFlow()

    // Track the exact handler instance registered with PlaybackManager so onCleared
    // can clear by identity (see PlaybackManager.clearNextMediaHandlerIfMatches).
    private var registeredNextHandler: (() -> Unit)? = null

    private val playerListener = object : Player.Listener {
        override fun onIsPlayingChanged(isPlaying: Boolean) {
            _uiState.update { it.copy(isPlaying = isPlaying) }
        }

        override fun onPlaybackStateChanged(playbackState: Int) {
            when (playbackState) {
                Player.STATE_BUFFERING -> {
                    _uiState.update { it.copy(isBuffering = true) }
                }
                Player.STATE_READY -> {
                    _uiState.update { it.copy(isBuffering = false, duration = _player?.duration ?: 0L) }
                }
                Player.STATE_ENDED -> {
                    _uiState.update { it.copy(isPlaying = false) }
                    saveProgress(completed = true)
                }
                Player.STATE_IDLE -> {
                    _uiState.update { it.copy(isBuffering = false) }
                }
            }
        }

        override fun onPositionDiscontinuity(
            oldPosition: Player.PositionInfo,
            newPosition: Player.PositionInfo,
            reason: Int,
        ) {
            _uiState.update { it.copy(currentPosition = newPosition.positionMs) }
        }

        override fun onVideoSizeChanged(videoSize: androidx.media3.common.VideoSize) {
            _uiState.update { it.copy(videoWidth = videoSize.width, videoHeight = videoSize.height) }
        }

        override fun onPlayerError(error: androidx.media3.common.PlaybackException) {
            super.onPlayerError(error)
            handlePlaybackError(error)
        }
    }

    init {
        playbackManager.ensureServiceStarted()
        initializePlayer()
        observeNextEpisodeForSession()
        loadPlaybackInfo()
    }

    private fun observeNextEpisodeForSession() {
        viewModelScope.launch {
            _uiState
                .map { it.nextEpisodeId }
                .distinctUntilChanged()
                .collect { nextId ->
                    // Clear our previously registered handler by identity so we don't
                    // accidentally wipe a handler that a newer screen has set.
                    registeredNextHandler?.let { playbackManager.clearNextMediaHandlerIfMatches(it) }
                    registeredNextHandler = if (nextId != null) {
                        val handler: () -> Unit = { _nextMediaRequested.tryEmit(nextId); Unit }
                        playbackManager.setNextMediaHandler(handler)
                        handler
                    } else {
                        null
                    }
                }
        }
    }

    private fun initializePlayer() {
        _player = playbackManager.attachUiPlayer(playerListener)
        syncUiStateFromPlayer()
    }

    private fun syncUiStateFromPlayer() {
        val player = _player ?: return
        val activeMediaId = playbackManager.currentMediaIdOrNull
        if (activeMediaId != null && activeMediaId != mediaId) {
            return
        }

        val duration = player.duration.takeIf { it != C.TIME_UNSET && it > 0 } ?: _uiState.value.duration
        val videoSize = player.videoSize
        _uiState.update {
            it.copy(
                isPlaying = player.isPlaying,
                isBuffering = player.playbackState == Player.STATE_BUFFERING,
                currentPosition = player.currentPosition.coerceAtLeast(0L),
                duration = duration,
                videoWidth = videoSize.width.takeIf { width -> width > 0 } ?: it.videoWidth,
                videoHeight = videoSize.height.takeIf { height -> height > 0 } ?: it.videoHeight,
            )
        }
    }

    override fun loadPlaybackInfo() {
        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    isLoading = true,
                    error = null,
                    nextEpisodeId = null,
                    nextEpisodeTitle = null,
                    showUpNext = false,
                )
            }
            upNextTriggered = false

            // Dynamic capability detection: Use MediaCodecUtil to check exactly what the hardware supports
            val supportedVideoCodecs = mutableListOf("h264") // Baseline minimum
            try {
                if (androidx.media3.exoplayer.mediacodec.MediaCodecUtil.getDecoderInfos(androidx.media3.common.MimeTypes.VIDEO_H265, false, false).isNotEmpty()) supportedVideoCodecs.add("hevc")
                if (androidx.media3.exoplayer.mediacodec.MediaCodecUtil.getDecoderInfos(androidx.media3.common.MimeTypes.VIDEO_VP9, false, false).isNotEmpty()) supportedVideoCodecs.add("vp9")
                if (androidx.media3.exoplayer.mediacodec.MediaCodecUtil.getDecoderInfos(androidx.media3.common.MimeTypes.VIDEO_AV1, false, false).isNotEmpty()) supportedVideoCodecs.add("av1")
            } catch (e: Exception) {
                // Ignore exception safely
            }

            val currentMaxQuality = _uiState.value.maxQuality
            val request = com.velox.app.domain.model.PlaybackInfoParams(
                videoCodecs = supportedVideoCodecs,
                audioCodecs = listOf("aac", "opus", "mp3", "flac", "ac3", "eac3"),
                containers = listOf("mp4", "mkv", "webm", "ts"),
                maxHeight = if (currentMaxQuality > 0) currentMaxQuality else null
            )

            val apiKeyDeferred = if (streamApiKey == null) {
                async { mediaRepository.getStreamApiKey(mediaId) }
            } else {
                null
            }

            // Run metadata fetches in parallel with getPlaybackInfo below. We await them
            // inside the getPlaybackInfo onSuccess branch so that MediaMetadata (title,
            // artwork) is populated BEFORE preparePlayer — otherwise the notification
            // renders with blank title for the first few hundred ms.
            val mediaDetailDeferred = async { mediaRepository.getMedia(mediaId) }
            val mediaFilesDeferred = async { mediaRepository.getMediaWithFilesInfo(mediaId) }

            mediaRepository.getPlaybackInfo(mediaId, request)
                .onSuccess { info ->
                    _uiState.update { state ->
                        state.copy(
                            isLoading = false,
                            playbackInfo = info,
                            videoWidth = info.width ?: state.videoWidth,
                            videoHeight = info.height ?: state.videoHeight,
                            audioTracks = info.audioTracks.mapIndexed { index, track ->
                                AudioTrackUi(index, track.label, track.language, track.codec, track.channels, track.isDefault)
                            },
                            subtitleTracks = listOf(
                                SubtitleTrackUi(-1, -1, "Off", "", null),
                            ) + info.subtitleTracks
                                .mapIndexed { index, track -> index to track }
                                .filter { (_, track) ->
                                    val fmt = track.format?.lowercase() ?: ""
                                    fmt == "srt" || fmt == "vtt" || fmt == "ass" || fmt.contains("subdl")
                                }
                                .map { (originalIndex, track) ->
                                    SubtitleTrackUi(track.id, originalIndex, track.label, track.language, track.format)
                                },
                            skipSegments = info.skipSegments.map { segment ->
                                SkipSegmentUi(
                                    type = segment.type,
                                    start = (segment.start * 1000).toLong(),
                                    end = (segment.end * 1000).toLong(),
                                )
                            },
                            qualityOptions = info.availableQualities.map { q ->
                                QualityOptionUi(height = q.height, label = q.label, instant = q.instant)
                            },
                        )
                    }

                    apiKeyDeferred?.await()
                        ?.onSuccess { key -> streamApiKey = key }
                        ?.onFailure { error -> Timber.w(error, "Failed to fetch stream api key") }

                    // Await metadata so the MediaItem's MediaMetadata (notification +
                    // lock screen) is populated before preparePlayer.
                    mediaDetailDeferred.await().onSuccess { media ->
                        _uiState.update { it.copy(mediaDetail = media) }
                    }
                    mediaFilesDeferred.await()
                        .onSuccess { filesInfo ->
                            Timber.d("mediaWithFiles OK: type=${filesInfo.mediaType} seriesId=${filesInfo.seriesId} seasonId=${filesInfo.seasonId}")
                            _uiState.update { it.copy(mediaTitle = filesInfo.title, mediaContext = filesInfo) }
                            if (filesInfo.mediaType == "episode" && filesInfo.seriesId != null && filesInfo.seasonId != null) {
                                loadNextEpisode(filesInfo.seriesId, filesInfo.seasonId)
                                loadSeasonPanel(filesInfo.seriesId, filesInfo.seasonId)
                            }
                        }
                        .onFailure { e ->
                            Timber.e(e, "mediaWithFiles FAILED: ${e.message}")
                            _uiState.value.mediaDetail?.let { media ->
                                _uiState.update { it.copy(mediaTitle = media.title) }
                            }
                        }

                    // Pick initial source from backend hint, then fallback always: direct → pretranscode → hls
                    val prefer = info.prefer ?: "direct"
                    val initOrder: List<String> = when (prefer) {
                        "pretranscode" -> listOf("pretranscode", "direct", "hls")
                        "hls" -> listOf("hls", "pretranscode", "direct")
                        else -> listOf("direct", "pretranscode", "hls")
                    }
                    // Pick first available source
                    val picked = initOrder.firstOrNull { getUrlForTier(it, info) != null }
                    if (picked != null) {
                        currentPlaybackSource = picked
                        // Fallback always follows: direct → pretranscode → hls (starting after current)
                        val fixedChain = listOf("direct", "pretranscode", "hls")
                        val startIdx = fixedChain.indexOf(picked)
                        fallbackOrder = if (startIdx >= 0) fixedChain.drop(startIdx) else fixedChain
                        val url = getUrlForTier(picked, info)!!
                        preparePlayer(url, info, maintainPosition = true)
                    } else {
                        // No source available at all
                        currentPlaybackSource = "hls"
                        fallbackOrder = emptyList()
                        preparePlayer(info.streamUrl, info, maintainPosition = true)
                    }

                    // Restore Subtitle Preferences
                    viewModelScope.launch {
                        val primaryLang = playerPrefsManager.primarySubLang.first()
                        val secondaryLang = playerPrefsManager.secondarySubLang.first()

                        val tracks = _uiState.value.subtitleTracks
                        if (primaryLang != null) {
                            val match = tracks.find { it.index != -1 && it.language == primaryLang }
                            if (match != null) {
                                selectSubtitleTrack(match.index)
                            }
                        }

                        if (secondaryLang != null) {
                            val match = tracks.find { it.index != -1 && it.language == secondaryLang }
                            if (match != null) {
                                selectSecondarySubtitleTrack(match.index)
                            }
                        }
                    }
                }
                .onFailure { error ->
                    _uiState.update { it.copy(isLoading = false, error = error.message) }
                }
        }
    }

    private fun loadSeasonPanel(seriesId: Int, initialSeasonId: Int) {
        viewModelScope.launch {
            mediaRepository.getSeasons(seriesId).onSuccess { seasons ->
                val selectedSeasonId = seasons.firstOrNull { it.id == initialSeasonId }?.id
                    ?: seasons.firstOrNull()?.id
                    ?: 0
                _uiState.update {
                    it.copy(
                        seasons = seasons,
                        seasonPanelSeasonId = selectedSeasonId,
                    )
                }
                if (selectedSeasonId > 0) {
                    loadSeasonEpisodes(seriesId, selectedSeasonId)
                }
            }
        }
    }

    private fun loadSeasonEpisodes(seriesId: Int, seasonId: Int) {
        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    isSeasonPanelLoading = true,
                    seasonPanelSeasonId = seasonId,
                )
            }
            mediaRepository.getEpisodes(seriesId, seasonId)
                .onSuccess { episodes ->
                    _uiState.update {
                        it.copy(
                            seasonPanelEpisodes = episodes,
                            isSeasonPanelLoading = false,
                        )
                    }
                }
                .onFailure {
                    _uiState.update {
                        it.copy(
                            seasonPanelEpisodes = emptyList(),
                            isSeasonPanelLoading = false,
                        )
                    }
                }
        }
    }

    private fun loadNextEpisode(seriesId: Int, seasonId: Int) {
        viewModelScope.launch {
            mediaRepository.getEpisodes(seriesId, seasonId).onSuccess { episodes ->
                val currentIdx = episodes.indexOfFirst { it.mediaId == mediaId }
                Timber.d("episodes: count=${episodes.size} currentIdx=$currentIdx mediaId=$mediaId")
                if (currentIdx != -1 && currentIdx < episodes.size - 1) {
                    val next = episodes[currentIdx + 1]
                    val nextNext = if (currentIdx + 2 < episodes.size) episodes[currentIdx + 2] else null
                    Timber.d("nextEpisode: id=${next.mediaId} title=${next.title}")

                    _uiState.update {
                        it.copy(
                            nextEpisodeId = next.mediaId,
                            nextEpisodeTitle = next.title,
                            nextNextEpisodeId = nextNext?.mediaId
                        )
                    }

                    // Check if the next episode is already watched
                    mediaRepository.getProgress(next.mediaId).onSuccess { progress ->
                        if (progress != null && progress.completed) {
                            _uiState.update { it.copy(nextEpisodeWatched = true) }
                        }
                    }
                }
            }.onFailure { e ->
                Timber.e("getEpisodes failed: ${e.message}")
            }
        }
    }

    private fun trySwitchToNextSource(info: DomainPlaybackInfo) {
        val nextTier = fallbackOrder.firstOrNull { it != currentPlaybackSource && getUrlForTier(it, info) != null }
        if (nextTier != null) {
            val url = getUrlForTier(nextTier, info)
            if (url != null) {
                currentPlaybackSource = nextTier
                preparePlayer(url, info, maintainPosition = true)
            }
        } else {
            // Already exhausted all tiers or only one was available
            currentPlaybackSource = "hls"
            preparePlayer(info.hlsUrl, info, maintainPosition = true)
        }
    }

    private fun getUrlForTier(tier: String, info: DomainPlaybackInfo): String? {
        return when (tier) {
            "direct" -> info.directUrl.takeIf { it.isNotEmpty() }
            "pretranscode" -> info.pretranscodeUrl?.takeIf { it.isNotEmpty() }
            "hls" -> info.hlsUrl.takeIf { it.isNotEmpty() }
            else -> null
        }
    }

    private fun handlePlaybackError(error: androidx.media3.common.PlaybackException) {
        Timber.e(
            error,
            "Playback failed on source=%s code=%s message=%s",
            currentPlaybackSource,
            error.errorCodeName,
            error.message,
        )
        val info = _uiState.value.playbackInfo
        if (info == null) {
            _uiState.update { it.copy(error = error.message) }
            return
        }

        // Remove the current source from the fallback order
        fallbackOrder = fallbackOrder.filter { it != currentPlaybackSource }

        if (fallbackOrder.isNotEmpty()) {
            // Try fallback
            trySwitchToNextSource(info)
        } else {
            _uiState.update { it.copy(error = "Playback error and no fallback available: ${error.message}") }
        }
    }

    private fun preparePlayer(url: String, info: DomainPlaybackInfo, maintainPosition: Boolean = false) {
        _player?.let { player ->
            // Re-capture current position if we are falling back
            val resumePosition = if (maintainPosition && player.currentPosition > 0) {
                player.currentPosition
            } else {
                info.position?.let { (it * 1000).toLong() } ?: 0L
            }

            Timber.d("Preparing source=%s url=%s resumeMs=%s", currentPlaybackSource, url, resumePosition)
            playbackManager.prepare(
                mediaId = mediaId,
                url = url,
                resumePosition = resumePosition,
                apiKey = streamApiKey,
                metadata = buildPlaybackMetadata(),
            )

            _uiState.update { it.copy(videoUrl = url) }
        }
    }

    private fun buildPlaybackMetadata(): PlaybackMetadata {
        val state = _uiState.value
        val title = state.mediaTitle ?: state.mediaDetail?.title
        val ctx = state.mediaContext
        val subtitle = if (ctx?.mediaType == "episode" && ctx.seasonNumber != null && ctx.episodeNumber != null) {
            "S${ctx.seasonNumber} · E${ctx.episodeNumber}"
        } else {
            state.mediaDetail?.releaseDate?.take(4)
        }
        // The repository already resolved posterPath/backdropPath to full URLs
        // via ImageUrlResolver, and /api/images/* is in the auth skip list, so
        // no extra api_key appending is needed for Media3's BitmapLoader.
        val artwork = state.mediaDetail?.backdrop ?: state.mediaDetail?.poster
        return PlaybackMetadata(title = title, subtitle = subtitle, artwork = artwork)
    }

    override fun togglePlayPause() {
        _player?.let { player ->
            if (player.isPlaying) {
                player.pause()
                // Ensure controls are visible when paused
                _uiState.update { it.copy(showControls = true) }
            } else {
                player.play()
                // Hide controls when playing and close detail panel like webapp
                _uiState.update {
                    it.copy(
                        showControls = false,
                        activeDetailPanel = DetailPanelUi.None,
                    )
                }
            }
        }
    }

    override fun seekTo(positionMs: Long) {
        _player?.seekTo(positionMs)
        _uiState.update { it.copy(currentPosition = positionMs) }
    }

    override fun seekForward(seconds: Int) {
        _player?.let { player ->
            val newPosition = (player.currentPosition + seconds * 1000).coerceAtMost(player.duration)
            player.seekTo(newPosition)
        }
    }

    override fun seekBackward(seconds: Int) {
        _player?.let { player ->
            val newPosition = (player.currentPosition - seconds * 1000).coerceAtLeast(0)
            player.seekTo(newPosition)
        }
    }

    override fun skipSegment(segment: SkipSegmentUi) {
        _player?.seekTo(segment.end)
    }

    override fun selectAudioTrack(trackIndex: Int) {
        _uiState.update { it.copy(selectedAudioTrackIndex = trackIndex) }

        _player?.let { player ->
            val track = _uiState.value.playbackInfo?.audioTracks?.getOrNull(trackIndex)
            if (track != null) {
                if (currentPlaybackSource == "hls") {
                    val info = _uiState.value.playbackInfo
                    val url = info?.hlsUrl
                    if (url != null) {
                        val baseWithoutAt = url.replace(Regex("&at=\\d+"), "").replace(Regex("\\?at=\\d+"), "")
                        val sep = if (baseWithoutAt.contains("?")) "&" else "?"
                        val newUrl = "$baseWithoutAt${sep}at=${track.id}"
                        preparePlayer(newUrl, info, maintainPosition = true)
                    }
                } else {
                    player.trackSelectionParameters = player.trackSelectionParameters
                        .buildUpon()
                        .setPreferredAudioLanguage(track.language)
                        .build()
                }
            }
        }
    }

    override fun selectSubtitleTrack(trackIndex: Int) {
        // trackIndex -1 means disabled
        _uiState.update { it.copy(selectedSubtitleIndex = trackIndex) }

        viewModelScope.launch {
            val trackLang = if (trackIndex == -1) null else _uiState.value.playbackInfo?.subtitleTracks?.getOrNull(trackIndex)?.language
            playerPrefsManager.setPrimarySubLang(trackLang)
        }

        _player?.let { player ->
            if (trackIndex == -1) {
                player.trackSelectionParameters = player.trackSelectionParameters
                    .buildUpon()
                    .setTrackTypeDisabled(androidx.media3.common.C.TRACK_TYPE_TEXT, true)
                    .clearOverridesOfType(androidx.media3.common.C.TRACK_TYPE_TEXT)
                    .build()
                _uiState.update { it.copy(primarySubtitleCues = emptyList()) }
            } else {
                val track = _uiState.value.playbackInfo?.subtitleTracks?.getOrNull(trackIndex)
                if (track != null) {
                    val primaryFileId = _uiState.value.playbackInfo?.primaryFileId
                    if (primaryFileId != null && !track.isImage) {
                        // Fetch subtitle content for DualSubtitleOverlay
                        viewModelScope.launch {
                            val result = mediaRepository.getSubtitleContent(primaryFileId, track.id)
                            result.onSuccess { content ->
                                val cues = parseSubtitleContent(content)
                                _uiState.update { it.copy(primarySubtitleCues = cues) }
                                // Disable native text rendering to avoid double overlay
                                _player?.trackSelectionParameters = _player!!.trackSelectionParameters
                                    .buildUpon()
                                    .setTrackTypeDisabled(androidx.media3.common.C.TRACK_TYPE_TEXT, true)
                                    .build()
                            }
                        }
                    } else if (track.isImage) {
                        // Image subtitle (PGS), DualSubtitleOverlay can't handle it, let ExoPlayer Native handle it
                        player.trackSelectionParameters = player.trackSelectionParameters
                            .buildUpon()
                            .setTrackTypeDisabled(androidx.media3.common.C.TRACK_TYPE_TEXT, false)
                            .setPreferredTextLanguage(track.language)
                            .build()
                    }
                } else {
                    player.trackSelectionParameters = player.trackSelectionParameters
                        .buildUpon()
                        .setTrackTypeDisabled(androidx.media3.common.C.TRACK_TYPE_TEXT, false)
                        .build()
                }
            }
        }
    }

    override fun selectSecondarySubtitleTrack(trackIndex: Int) {
        _uiState.update { it.copy(selectedSecondarySubtitleIndex = trackIndex) }

        viewModelScope.launch {
            val trackLang = if (trackIndex == -1) null else _uiState.value.playbackInfo?.subtitleTracks?.getOrNull(trackIndex)?.language
            playerPrefsManager.setSecondarySubLang(trackLang)
        }

        if (trackIndex == -1) {
            _uiState.update { it.copy(secondarySubtitleCues = emptyList(), secondarySubtitleEnabled = false) }
            return
        }

        val track = _uiState.value.playbackInfo?.subtitleTracks?.getOrNull(trackIndex)
        val primaryFileId = _uiState.value.playbackInfo?.primaryFileId
        if (track != null && primaryFileId != null && !track.isImage) {
            viewModelScope.launch {
                val result = mediaRepository.getSubtitleContent(primaryFileId, track.id)
                result.onSuccess { content ->
                    val cues = parseSubtitleContent(content)
                    _uiState.update { it.copy(secondarySubtitleCues = cues, secondarySubtitleEnabled = true) }
                }
            }
        }
    }

    private fun parseSubtitleContent(content: String): List<SubtitleCueUi> {
        val cues = mutableListOf<SubtitleCueUi>()
        val lines = content.replace("\r\n", "\n").split("\n")
        var i = 0
        while (i < lines.size) {
            val line = lines[i].trim()
            if (line.isEmpty() || line.startsWith("WEBVTT") || line.startsWith("NOTE")) {
                i++
                continue
            }
            if (line.contains("-->")) {
                val times = line.split("-->")
                if (times.size == 2) {
                    val start = parseTimeMs(times[0].trim())
                    val end = parseTimeMs(times[1].trim())
                    i++
                    val textBuilder = java.lang.StringBuilder()
                    while (i < lines.size && lines[i].trim().isNotEmpty()) {
                        textBuilder.append(lines[i].trim()).append("\n")
                        i++
                    }
                    cues.add(SubtitleCueUi(start, end, textBuilder.toString().trim()))
                }
            } else {
                if (i + 1 < lines.size && lines[i + 1].contains("-->")) {
                    i++
                    continue
                } else {
                    i++
                }
            }
        }
        return cues
    }

    private fun parseTimeMs(timeStr: String): Long {
        var totalMs = 0L
        try {
            val timeRegex = Regex("(?:(\\d+):)?(\\d+):(\\d+)[.,](\\d+)")
            val match = timeRegex.find(timeStr)
            if (match != null) {
                val h = match.groups[1]?.value?.toLong() ?: 0L
                val m = match.groups[2]?.value?.toLong() ?: 0L
                val s = match.groups[3]?.value?.toLong() ?: 0L
                val msStr = match.groups[4]?.value ?: "0"
                val ms = msStr.padEnd(3, '0').substring(0, 3).toLong()
                totalMs = h * 3600000L + m * 60000L + s * 1000L + ms
            }
        } catch(e: Exception) {}
        return totalMs
    }

    override fun toggleControls() {
        if (_uiState.value.isLocked) return
        _uiState.update { it.copy(showControls = !it.showControls) }
    }

    override fun toggleDetailPanel(panel: DetailPanelUi) {
        if (panel == DetailPanelUi.Season && _uiState.value.seasons.isEmpty()) {
            return
        }

        val nextPanel = if (_uiState.value.activeDetailPanel == panel) {
            DetailPanelUi.None
        } else {
            panel
        }

        if (nextPanel != DetailPanelUi.None && _player?.isPlaying == true) {
            _player?.pause()
        }

        _uiState.update {
            it.copy(
                activeDetailPanel = nextPanel,
                showControls = true,
            )
        }
    }

    override fun closeDetailPanel() {
        _uiState.update { it.copy(activeDetailPanel = DetailPanelUi.None) }
    }

    override fun selectSeasonPanelSeason(seasonId: Int) {
        val seriesId = _uiState.value.mediaContext?.seriesId ?: return
        if (_uiState.value.seasonPanelSeasonId == seasonId && _uiState.value.seasonPanelEpisodes.isNotEmpty()) {
            return
        }
        loadSeasonEpisodes(seriesId, seasonId)
    }

    override fun setPlaybackSpeed(speed: Float) {
        _player?.setPlaybackSpeed(speed)
        _uiState.update { it.copy(playbackSpeed = speed) }
    }

    override fun toggleLock() {
        _uiState.update { state ->
            state.copy(
                isLocked = !state.isLocked,
                showControls = state.isLocked, // Show controls when unlocking
            )
        }
    }

    override fun setVolume(volume: Float) {
        _player?.volume = volume
        _uiState.update { it.copy(volume = volume) }
    }

    // Quality — re-fetch playback info with new max_height (like webapp)
    override fun setMaxQuality(height: Int) {
        val previous = _uiState.value.maxQuality
        _uiState.update { it.copy(maxQuality = height) }
        if (height != previous) {
            loadPlaybackInfo()
        }
    }

    // Aspect ratio
    override fun setAspectRatio(ratio: AspectRatioUi) {
        _uiState.update { it.copy(aspectRatio = ratio) }
    }

    // Repeat mode
    override fun setRepeatMode(mode: RepeatModeUi) {
        _uiState.update { it.copy(repeatMode = mode) }
        _player?.repeatMode = when (mode) {
            RepeatModeUi.None -> Player.REPEAT_MODE_OFF
            RepeatModeUi.One -> Player.REPEAT_MODE_ONE
            RepeatModeUi.All -> Player.REPEAT_MODE_ALL
        }
    }

    override fun cycleRepeatMode() {
        val current = _uiState.value.repeatMode
        val next = when (current) {
            RepeatModeUi.None -> RepeatModeUi.One
            RepeatModeUi.One -> RepeatModeUi.All
            RepeatModeUi.All -> RepeatModeUi.None
        }
        setRepeatMode(next)
    }

    // Subtitle delay
    override fun setSubtitleDelay(delay: Float) {
        _uiState.update { it.copy(subtitleDelay = delay) }
    }

    override fun adjustSubtitleDelay(delta: Float) {
        _uiState.update { it.copy(subtitleDelay = (it.subtitleDelay + delta).coerceIn(-10f, 10f)) }
    }

    // Subtitle appearance
    override fun setSubtitleSize(size: SubtitleSizeUi) {
        _uiState.update { it.copy(subtitleSize = size) }
    }

    override fun setSubtitleColor(color: String) {
        _uiState.update { it.copy(subtitleColor = color) }
    }

    override fun setSubtitleBackground(bg: SubtitleBackgroundUi) {
        _uiState.update { it.copy(subtitleBackground = bg) }
    }

    // Playback stats
    override fun togglePlaybackStats() {
        _uiState.update { it.copy(showPlaybackStats = !it.showPlaybackStats) }
    }

    // Seek feedback
    private var seekFeedbackJob: kotlinx.coroutines.Job? = null

    override fun showSeekFeedback(side: String, amount: Int) {
        _uiState.update { it.copy(showSeekFeedback = true, seekFeedbackSide = side, seekFeedbackAmount = amount) }
        seekFeedbackJob?.cancel()
        seekFeedbackJob = viewModelScope.launch {
            kotlinx.coroutines.delay(800)
            hideSeekFeedback()
        }
    }

    override fun hideSeekFeedback() {
        _uiState.update { it.copy(showSeekFeedback = false) }
    }

    // Cinema overlay - cinema items loaded from playback info or settings
    override fun loadCinemaItems(cinemaEnabled: Boolean) {
        // Cinema items would be loaded from API or settings
        // For now, leave empty - will be implemented when backend supports it
        if (!cinemaEnabled) {
            _uiState.update { it.copy(showCinemaOverlay = false) }
        }
    }

    override fun skipCinemaItem() {
        val state = _uiState.value
        val nextIndex = state.cinemaIndex + 1
        if (nextIndex >= state.cinemaItems.size) {
            _uiState.update { it.copy(showCinemaOverlay = false, cinemaIndex = 0) }
        } else {
            _uiState.update { it.copy(cinemaIndex = nextIndex) }
        }
    }

    override fun skipAllCinema() {
        _uiState.update { it.copy(showCinemaOverlay = false, cinemaIndex = 0) }
    }

    // Secondary subtitle
    override fun enableSecondarySubtitle(enabled: Boolean) {
        _uiState.update { it.copy(secondarySubtitleEnabled = enabled) }
    }

    // Up Next / Next Episode
    override fun showUpNext(nextEpisodeId: Int, nextEpisodeTitle: String) {
        _uiState.update { it.copy(showUpNext = true, upNextId = nextEpisodeId, upNextTitle = nextEpisodeTitle, upNextCountdown = 15) }
    }

    override fun dismissUpNext() {
        _uiState.update { it.copy(showUpNext = false, upNextCountdown = 0) }
    }

    override fun getUpNextId(): Int? = _uiState.value.upNextId

    private var upNextTriggered = false
    private var lastProgressSaveTime = 0L

    override fun updatePosition() {
        _player?.let { player ->
            val pos = player.currentPosition
            val dur = player.duration.takeIf { it > 0 } ?: return@let
            _uiState.update { it.copy(currentPosition = pos) }

            // Auto-save progress to server every 10 seconds (like webapp)
            val now = System.currentTimeMillis()
            if (now - lastProgressSaveTime >= 10_000 || pos.toFloat() / dur >= 0.95f) {
                lastProgressSaveTime = now
                saveProgress()
            }

            // Auto-trigger Up Next at credits start or last 30s (like Expo mobile)
            val state = _uiState.value
            if (!upNextTriggered && !state.showUpNext && state.nextEpisodeId != null && state.nextEpisodeTitle != null) {
                val creditsStart = state.skipSegments.find { it.type == "credits" }?.start
                val timeRemaining = dur - pos
                val inCredits = creditsStart != null && pos >= creditsStart
                val isNearEnd = timeRemaining in 1..30000 // last 30 seconds
                if (inCredits || isNearEnd) {
                    upNextTriggered = true
                    _uiState.update { it.copy(showUpNext = true, upNextId = state.nextEpisodeId, upNextTitle = state.nextEpisodeTitle, upNextCountdown = 15) }
                }
            }
        }
    }

    override fun saveProgress(completed: Boolean) {
        viewModelScope.launch {
            val position = (_player?.currentPosition ?: 0L) / 1000f
            val duration = (_player?.duration ?: 0L) / 1000f
            val isCompleted = completed || (duration > 0 && position / duration >= 0.95f)
            mediaRepository.updateProgress(mediaId, position, isCompleted)
        }
    }

    // Subtitle search
    override fun showSubtitleSearch() {
        _uiState.update { it.copy(showSubtitleSearch = true) }
    }

    override fun hideSubtitleSearch() {
        _uiState.update { it.copy(showSubtitleSearch = false) }
    }

    override fun searchSubtitles(lang: String) {
        _uiState.update { it.copy(isSearchingSubtitles = true, subtitleSearchLang = lang, subtitleSearchResults = emptyList()) }
        viewModelScope.launch {
            mediaRepository.searchSubtitles(mediaId, lang)
                .onSuccess { results ->
                    _uiState.update { state ->
                        state.copy(
                            isSearchingSubtitles = false,
                            subtitleSearchResults = results.map { result ->
                                SubtitleSearchResultUi(
                                    id = result.externalId,
                                    title = result.title,
                                    lang = result.language,
                                    hi = result.hi,
                                    aiTranslated = result.aiTranslated,
                                )
                            },
                        )
                    }
                }
                .onFailure {
                    _uiState.update { it.copy(isSearchingSubtitles = false) }
                }
        }
    }

    override fun downloadSubtitle(provider: String, externalId: String, language: String) {
        _uiState.update { it.copy(isDownloadingSubtitle = true) }
        viewModelScope.launch {
            mediaRepository.downloadSubtitle(mediaId, provider, externalId, language)
                .onSuccess {
                    _uiState.update { it.copy(isDownloadingSubtitle = false, showSubtitleSearch = false) }
                    // Refresh playback info to pick up new subtitle
                    loadPlaybackInfo()
                }
                .onFailure {
                    _uiState.update { it.copy(isDownloadingSubtitle = false) }
                }
        }
    }

    // Subtitle translate
    private val _isTranslatingSubtitle = MutableStateFlow(false)
    override val isTranslatingSubtitle: StateFlow<Boolean> = _isTranslatingSubtitle.asStateFlow()

    override fun translateSubtitle(subtitleId: Int, targetLanguage: String) {
        _isTranslatingSubtitle.value = true
        viewModelScope.launch {
            mediaRepository.translateSubtitle(subtitleId, targetLanguage)
                .onSuccess {
                    _isTranslatingSubtitle.value = false
                    // Refresh playback info to pick up translated subtitle
                    loadPlaybackInfo()
                }
                .onFailure {
                    _isTranslatingSubtitle.value = false
                }
        }
    }

    // Watched Warning
    override fun showWatchedWarning() {
        _uiState.update { it.copy(showWatchedWarning = true) }
    }

    override fun hideWatchedWarning() {
        _uiState.update { it.copy(showWatchedWarning = false) }
    }

    override fun resetProgressAndNavigate(targetMediaId: Int, navigate: (Int) -> Unit) {
        viewModelScope.launch {
            mediaRepository.updateProgress(targetMediaId, 0f, false)
            navigate(targetMediaId)
        }
    }

    override fun onCleared() {
        registeredNextHandler?.let { playbackManager.clearNextMediaHandlerIfMatches(it) }
        registeredNextHandler = null
        playbackManager.persistCurrentProgress()
        playbackManager.releasePlayer()
        playbackManager.detachUiPlayer(playerListener)
        _player = null
        super.onCleared()
    }
}
