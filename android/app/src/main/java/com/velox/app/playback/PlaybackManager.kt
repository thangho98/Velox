package com.velox.app.playback

import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.media3.common.AudioAttributes
import androidx.media3.common.C
import androidx.media3.common.ForwardingPlayer
import androidx.media3.common.MediaItem
import androidx.media3.common.MediaMetadata
import androidx.media3.common.Player
import androidx.media3.datasource.DefaultDataSource
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.DefaultMediaSourceFactory
import com.velox.app.domain.repository.MediaRepository
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import timber.log.Timber

data class PlaybackRuntimeState(
    val isPlaying: Boolean = false,
    val isBuffering: Boolean = false,
    val duration: Long = 0L,
    val currentPosition: Long = 0L,
    val videoWidth: Int = 0,
    val videoHeight: Int = 0,
    val playbackState: Int = Player.STATE_IDLE,
    val currentMediaUrl: String? = null,
    val currentMediaId: Int? = null,
)

data class PlaybackMetadata(
    val title: String? = null,
    val subtitle: String? = null,
    val artwork: com.velox.app.domain.model.ImageResource? = null,
)

@Singleton
class PlaybackManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val mediaRepository: MediaRepository,
) {
    private val _state = MutableStateFlow(PlaybackRuntimeState())
    val state: StateFlow<PlaybackRuntimeState> = _state.asStateFlow()

    private val managerScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var playerInternal: ExoPlayer? = null
    private var sessionPlayerInternal: Player? = null
    private var serviceStarted = false
    private val uiListeners = mutableSetOf<Player.Listener>()
    private var currentMediaId: Int? = null
    private var nextMediaHandler: (() -> Unit)? = null

    private val runtimeListener = object : Player.Listener {
        override fun onIsPlayingChanged(isPlaying: Boolean) {
            _state.update { it.copy(isPlaying = isPlaying) }
        }

        override fun onPlaybackStateChanged(playbackState: Int) {
            _state.update {
                it.copy(
                    playbackState = playbackState,
                    isBuffering = playbackState == Player.STATE_BUFFERING,
                    duration = playerInternal?.duration ?: it.duration,
                )
            }
        }

        override fun onPositionDiscontinuity(
            oldPosition: Player.PositionInfo,
            newPosition: Player.PositionInfo,
            reason: Int,
        ) {
            _state.update { it.copy(currentPosition = newPosition.positionMs) }
        }

        override fun onVideoSizeChanged(videoSize: androidx.media3.common.VideoSize) {
            _state.update {
                it.copy(
                    videoWidth = videoSize.width,
                    videoHeight = videoSize.height,
                )
            }
        }
    }

    val playerOrNull: ExoPlayer?
        get() = playerInternal

    val currentMediaIdOrNull: Int?
        get() = currentMediaId

    fun ensurePlayer(): ExoPlayer {
        playerInternal?.let { return it }

        return ExoPlayer.Builder(context)
            .setAudioAttributes(
                AudioAttributes.Builder()
                    .setUsage(C.USAGE_MEDIA)
                    .setContentType(C.AUDIO_CONTENT_TYPE_MOVIE)
                    .build(),
                true,
            )
            .setSeekBackIncrementMs(SEEK_INCREMENT_MS)
            .setSeekForwardIncrementMs(SEEK_INCREMENT_MS)
            .build()
            .apply {
                addListener(runtimeListener)
                trackSelectionParameters = trackSelectionParameters
                    .buildUpon()
                    .setTrackTypeDisabled(C.TRACK_TYPE_TEXT, true)
                    .build()
                playerInternal = this
            }
    }

    fun buildSessionPlayer(): Player {
        sessionPlayerInternal?.let { return it }
        val base = ensurePlayer()
        return object : ForwardingPlayer(base) {
            override fun hasNextMediaItem(): Boolean =
                nextMediaHandler != null || super.hasNextMediaItem()

            override fun getAvailableCommands(): Player.Commands {
                val original = super.getAvailableCommands()
                return if (nextMediaHandler != null) {
                    original.buildUpon().add(Player.COMMAND_SEEK_TO_NEXT_MEDIA_ITEM).build()
                } else {
                    original
                }
            }

            override fun seekToNextMediaItem() {
                val handler = nextMediaHandler
                if (handler != null) {
                    handler.invoke()
                } else {
                    super.seekToNextMediaItem()
                }
            }
        }.also { sessionPlayerInternal = it }
    }

    fun setNextMediaHandler(handler: (() -> Unit)?) {
        nextMediaHandler = handler
    }

    // Identity-based clear: avoids a race during rapid nav (A→B) where the outgoing
    // screen's onCleared would otherwise wipe the incoming screen's freshly registered
    // handler. Only clears if the current handler reference matches `expected`.
    fun clearNextMediaHandlerIfMatches(expected: () -> Unit) {
        if (nextMediaHandler === expected) {
            nextMediaHandler = null
        }
    }

    fun attachUiPlayer(listener: Player.Listener): ExoPlayer {
        val player = ensurePlayer()
        if (uiListeners.add(listener)) {
            player.addListener(listener)
        }
        return player
    }

    fun detachUiPlayer(listener: Player.Listener) {
        if (uiListeners.remove(listener)) {
            playerInternal?.removeListener(listener)
        }
        releaseIfUnused()
    }

    fun ensureServiceStarted() {
        if (serviceStarted) return

        runCatching {
            context.startService(Intent(context, VeloxPlaybackService::class.java))
            serviceStarted = true
        }.onFailure { error ->
            Timber.e(error, "Failed to start playback service")
        }
    }

    fun markServiceStarted() {
        serviceStarted = true
    }

    fun markServiceStopped() {
        serviceStarted = false
        releaseIfUnused()
    }

    fun prepare(
        mediaId: Int,
        url: String,
        resumePosition: Long = 0L,
        apiKey: String? = null,
        metadata: PlaybackMetadata? = null,
    ) {
        val previousMediaId = currentMediaId
        if (previousMediaId != null && previousMediaId != mediaId) {
            persistCurrentProgress()
        }

        val player = ensurePlayer()
        currentMediaId = mediaId
        val authenticatedUrl = appendApiKey(url, apiKey)
        player.setMediaSource(buildMediaSource(mediaId, authenticatedUrl, metadata))
        if (resumePosition > 0) {
            player.seekTo(resumePosition)
        }
        player.prepare()
        player.playWhenReady = true
        _state.update {
            it.copy(
                currentMediaUrl = authenticatedUrl,
                currentMediaId = mediaId,
                currentPosition = resumePosition,
            )
        }
    }

    fun pause() {
        playerInternal?.pause()
    }

    fun persistCurrentProgress(completed: Boolean = false) {
        val mediaId = currentMediaId ?: return
        val player = playerInternal ?: return
        val position = player.currentPosition / 1000f
        val duration = player.duration
            .takeIf { it > 0 }
            ?.div(1000f)
            ?: 0f
        val isCompleted = completed || (duration > 0f && position / duration >= 0.95f)

        managerScope.launch {
            mediaRepository.updateProgress(mediaId, position, isCompleted)
                .onFailure { error ->
                    Timber.e(error, "Failed to persist playback progress for mediaId=%s", mediaId)
                }
        }
    }

    fun releaseIfUnused(force: Boolean = false) {
        val player = playerInternal ?: return
        if (!force && (serviceStarted || uiListeners.isNotEmpty() || player.isPlaying)) {
            return
        }
        releasePlayer()
    }

    fun releasePlayer() {
        persistCurrentProgress()
        playerInternal?.removeListener(runtimeListener)
        playerInternal?.release()
        playerInternal = null
        sessionPlayerInternal = null
        nextMediaHandler = null
        uiListeners.clear()
        currentMediaId = null
        _state.value = PlaybackRuntimeState()
    }

    private fun appendApiKey(url: String, apiKey: String?): String {
        if (apiKey.isNullOrBlank()) return url
        val separator = if ('?' in url) '&' else '?'
        return "$url${separator}api_key=$apiKey"
    }

    private fun buildMediaSource(mediaId: Int, url: String, metadata: PlaybackMetadata?) =
        DefaultMediaSourceFactory(
            DefaultDataSource.Factory(
                context,
                DefaultHttpDataSource.Factory()
                    .setAllowCrossProtocolRedirects(true)
                    .setConnectTimeoutMs(45_000)
                    .setReadTimeoutMs(45_000),
            ),
        ).createMediaSource(
            MediaItem.Builder()
                .setMediaId(mediaId.toString())
                .setUri(url)
                .apply { metadata?.let { setMediaMetadata(buildMediaMetadata(it)) } }
                .build()
        )

    private fun buildMediaMetadata(metadata: PlaybackMetadata): MediaMetadata =
        MediaMetadata.Builder()
            .setTitle(metadata.title)
            .setArtist(metadata.subtitle)
            .setDisplayTitle(metadata.title)
            .setSubtitle(metadata.subtitle)
            .setIsBrowsable(false)
            .setIsPlayable(true)
            .setMediaType(MediaMetadata.MEDIA_TYPE_MOVIE)
            .apply {
                metadata.artwork?.url
                    ?.takeIf { it.isNotBlank() }
                    ?.let { setArtworkUri(Uri.parse(it)) }
            }
            .build()

    companion object {
        private const val SEEK_INCREMENT_MS = 10_000L
    }
}
