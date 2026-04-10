package com.velox.app.presentation.ui.screens.player

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import androidx.hilt.navigation.compose.hiltViewModel
import com.velox.app.presentation.ui.components.VideoPlayer
import com.velox.app.presentation.viewmodel.PlayerViewModel
import com.velox.app.ui.theme.VeloxTheme

@Composable
fun PlayerScreen(
    onBackClick: () -> Unit,
    onNavigateToMedia: (Int) -> Unit = {},
    viewModel: PlayerViewModel = hiltViewModel(),
) {
    VideoPlayer(
        onBackClick = onBackClick,
        onNavigateToMedia = onNavigateToMedia,
        viewModel = viewModel,
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
