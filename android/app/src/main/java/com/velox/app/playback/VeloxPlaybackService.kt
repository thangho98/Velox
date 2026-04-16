package com.velox.app.playback

import android.content.Intent
import androidx.media3.session.MediaSession
import androidx.media3.session.MediaSessionService
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class VeloxPlaybackService : MediaSessionService() {

    @Inject
    lateinit var playbackManager: PlaybackManager

    private var mediaSession: MediaSession? = null

    override fun onCreate() {
        super.onCreate()
        playbackManager.markServiceStarted()
        val player = playbackManager.buildSessionPlayer()
        mediaSession = MediaSession.Builder(this, player)
            .setCallback(
                object : MediaSession.Callback {
                    override fun onConnect(
                        session: MediaSession,
                        controller: MediaSession.ControllerInfo,
                    ): MediaSession.ConnectionResult {
                        val isAllowedController =
                            controller.isTrusted || controller.packageName == packageName

                        return if (isAllowedController) {
                            MediaSession.ConnectionResult.AcceptedResultBuilder(session).build()
                        } else {
                            MediaSession.ConnectionResult.reject()
                        }
                    }
                }
            )
            .build()
    }

    override fun onGetSession(controllerInfo: MediaSession.ControllerInfo): MediaSession? = mediaSession

    override fun onTaskRemoved(rootIntent: Intent?) {
        playbackManager.pause()
        playbackManager.persistCurrentProgress()
        stopSelf()
    }

    override fun onDestroy() {
        mediaSession?.release()
        mediaSession = null
        playbackManager.markServiceStopped()
        super.onDestroy()
    }
}
