package com.velox.app.presentation.tv.screens

import android.view.ViewGroup
import android.widget.FrameLayout
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.viewinterop.AndroidView
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.media3.ui.PlayerView
import com.velox.app.presentation.tv.components.TvPlayerOSD
import com.velox.app.presentation.viewmodel.PlayerViewModel

@Composable
fun TvPlayerScreen(
    mediaId: Int,
    onNavigateBack: () -> Unit,
    viewModel: PlayerViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val context = LocalContext.current

    DisposableEffect(Unit) {
        onDispose {
            viewModel.saveProgress(completed = false)
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black)
    ) {
        // Render ExoPlayer
        if (viewModel.player != null) {
            AndroidView(
                factory = {
                    PlayerView(context).apply {
                        this.player = viewModel.player
                        this.useController = false // Disable native ExoPlayer controls
                        this.layoutParams = FrameLayout.LayoutParams(
                            ViewGroup.LayoutParams.MATCH_PARENT,
                            ViewGroup.LayoutParams.MATCH_PARENT
                        )
                    }
                },
                modifier = Modifier.fillMaxSize()
            )
        }

        // TV Optimized OSD Overlay
        TvPlayerOSD(
            isPlaying = uiState.isPlaying,
            title = uiState.mediaTitle ?: uiState.mediaDetail?.title ?: "Playing...",
            currentPosition = uiState.currentPosition,
            duration = uiState.duration,
            onPlayPause = { viewModel.togglePlayPause() },
            onSeekForward = { viewModel.seekForward(10) },
            onSeekBackward = { viewModel.seekBackward(10) },
            onShowSettings = {
                // In a real app, this would toggle a side panel to select Tracks / Subtitles.
                // We'll leave the hooks ready.
            },
            onBack = onNavigateBack
        )
    }
}
