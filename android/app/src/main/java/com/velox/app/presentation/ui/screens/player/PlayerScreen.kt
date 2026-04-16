package com.velox.app.presentation.ui.screens.player

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.velox.app.presentation.ui.components.VideoPlayer
import com.velox.app.presentation.viewmodel.PlayerViewModel
import com.velox.app.ui.theme.VeloxTheme

@Composable
fun PlayerScreen(
    onBackClick: () -> Unit,
    onNavigateToMedia: (Int) -> Unit = {},
    viewModel: PlayerViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(viewModel) {
        // onCleared of the outgoing ViewModel already persists progress via
        // PlaybackManager.persistCurrentProgress — no need to save here.
        viewModel.nextMediaRequested.collect { nextId ->
            onNavigateToMedia(nextId)
        }
    }

    VideoPlayer(
        onBackClick = onBackClick,
        onNavigateToMedia = onNavigateToMedia,
        actions = viewModel,
        uiState = uiState,
        modifier = Modifier.fillMaxSize(),
    )
}

@Preview(showBackground = true)
@Composable
fun PlayerScreenPreview() {
    VeloxTheme {
        PlayerScreen(
            onBackClick = {},
        )
    }
}
