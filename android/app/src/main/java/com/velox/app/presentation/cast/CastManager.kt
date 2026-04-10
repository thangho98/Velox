package com.velox.app.presentation.cast

import android.content.Context
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.remember
import androidx.compose.ui.platform.LocalContext
import com.google.android.gms.cast.CastDevice
import com.google.android.gms.cast.MediaInfo
import com.google.android.gms.cast.MediaLoadRequestData
import com.google.android.gms.cast.MediaMetadata
import com.google.android.gms.cast.MediaSeekOptions
import com.google.android.gms.cast.framework.CastContext
import com.google.android.gms.cast.framework.CastSession
import com.google.android.gms.cast.framework.SessionManager
import com.google.android.gms.cast.framework.SessionManagerListener
import com.google.android.gms.cast.framework.media.RemoteMediaClient
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class CastUiState(
    val isConnected: Boolean = false,
    val isConnecting: Boolean = false,
    val deviceName: String? = null,
    val device: CastDevice? = null,
)

class CastManager(private val context: Context) {
    private var castContext: CastContext? = null
    private var sessionManager: SessionManager? = null
    private var castSession: CastSession? = null
    private var remoteMediaClient: RemoteMediaClient? = null

    private val _uiState = MutableStateFlow(CastUiState())
    val uiState: StateFlow<CastUiState> = _uiState.asStateFlow()

    private val sessionManagerListener = object : SessionManagerListener<CastSession> {
        override fun onSessionStarted(session: CastSession, sessionId: String) {
            castSession = session
            remoteMediaClient = session.remoteMediaClient
            _uiState.value = _uiState.value.copy(
                isConnected = true,
                isConnecting = false,
                deviceName = session.castDevice?.friendlyName,
                device = session.castDevice,
            )
        }

        override fun onSessionResumed(session: CastSession, wasSuspended: Boolean) {
            castSession = session
            remoteMediaClient = session.remoteMediaClient
            _uiState.value = _uiState.value.copy(
                isConnected = true,
                isConnecting = false,
                deviceName = session.castDevice?.friendlyName,
                device = session.castDevice,
            )
        }

        override fun onSessionSuspended(session: CastSession, reason: Int) {
            _uiState.value = _uiState.value.copy(
                isConnected = false,
                isConnecting = false,
            )
        }

        override fun onSessionEnded(session: CastSession, error: Int) {
            castSession = null
            remoteMediaClient = null
            _uiState.value = _uiState.value.copy(
                isConnected = false,
                isConnecting = false,
                deviceName = null,
                device = null,
            )
        }

        override fun onSessionStarting(session: CastSession) {
            _uiState.value = _uiState.value.copy(isConnecting = true)
        }

        override fun onSessionResuming(session: CastSession, sessionId: String) {
            _uiState.value = _uiState.value.copy(isConnecting = true)
        }

        override fun onSessionResumeFailed(session: CastSession, error: Int) {
            _uiState.value = _uiState.value.copy(isConnecting = false)
        }

        override fun onSessionStartFailed(session: CastSession, error: Int) {
            _uiState.value = _uiState.value.copy(isConnecting = false)
        }

        override fun onSessionEnding(session: CastSession) {
            _uiState.value = _uiState.value.copy(isConnecting = false)
        }
    }

    fun initialize() {
        try {
            castContext = CastContext.getSharedInstance(context)
            sessionManager = castContext?.sessionManager
            sessionManager?.addSessionManagerListener(sessionManagerListener, CastSession::class.java)

            // Check for existing session
            castSession = sessionManager?.currentCastSession
            if (castSession != null && castSession?.isConnected == true) {
                remoteMediaClient = castSession?.remoteMediaClient
                _uiState.value = _uiState.value.copy(
                    isConnected = true,
                    deviceName = castSession?.castDevice?.friendlyName,
                    device = castSession?.castDevice,
                )
            }
        } catch (e: Exception) {
            // Cast not available
            _uiState.value = CastUiState()
        }
    }

    fun getCastSession(): CastSession? = castSession

    fun getRemoteMediaClient(): RemoteMediaClient? = remoteMediaClient

    @Suppress("UNUSED_PARAMETER")
    fun loadMediaAndPlay(
        streamUrl: String,
        title: String,
        subtitle: String? = null,
        imageUrl: String? = null, // Currently unused - Cast SDK shows video thumbnail differently
        positionMs: Long = 0,
        apiKey: String? = null,
    ) {
        val mediaClient = remoteMediaClient ?: return

        val urlWithKey = if (apiKey != null) {
            "$streamUrl?api_key=$apiKey"
        } else {
            streamUrl
        }

        val metadata = MediaMetadata(MediaMetadata.MEDIA_TYPE_MOVIE).apply {
            putString(MediaMetadata.KEY_TITLE, title)
            subtitle?.let { putString(MediaMetadata.KEY_SUBTITLE, it) }
        }

        val mediaInfo = MediaInfo.Builder(urlWithKey)
            .setStreamType(MediaInfo.STREAM_TYPE_BUFFERED)
            .setContentType("video/mp4")
            .setMetadata(metadata)
            .build()

        val loadRequest = MediaLoadRequestData.Builder()
            .setMediaInfo(mediaInfo)
            .setCurrentTime(positionMs)
            .build()

        mediaClient.load(loadRequest)
    }

    fun play() {
        remoteMediaClient?.play()
    }

    fun pause() {
        remoteMediaClient?.pause()
    }

    fun seekTo(positionMs: Long) {
        val mediaClient = remoteMediaClient ?: return
        val options = MediaSeekOptions.Builder()
            .setPosition(positionMs)
            .build()
        mediaClient.seek(options)
    }

    fun stop() {
        remoteMediaClient?.stop()
    }

    fun endSession() {
        sessionManager?.endCurrentSession(true)
    }

    fun release() {
        sessionManager?.removeSessionManagerListener(sessionManagerListener, CastSession::class.java)
        castSession = null
        remoteMediaClient = null
        castContext = null
        sessionManager = null
    }
}

@Composable
fun rememberCastManager(): CastManager {
    val context = LocalContext.current
    val castManager = remember { CastManager(context) }

    DisposableEffect(Unit) {
        castManager.initialize()
        onDispose {
            castManager.release()
        }
    }

    return castManager
}
