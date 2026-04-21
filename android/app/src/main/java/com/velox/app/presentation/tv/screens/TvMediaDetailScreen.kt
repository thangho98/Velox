package com.velox.app.presentation.tv.screens

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.tv.material3.Button
import androidx.tv.material3.ButtonDefaults
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.Text
import coil.compose.AsyncImage
import coil.request.ImageRequest
import com.velox.app.presentation.viewmodel.MediaDetailUiState
import com.velox.app.presentation.viewmodel.MediaDetailViewModel

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun TvMediaDetailScreen(
    mediaId: Int,
    onNavigateToPlayer: (Int) -> Unit,
    viewModel: MediaDetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val context = LocalContext.current



    Box(modifier = Modifier.fillMaxSize()) {
        uiState.media?.let { media ->
            // Background Backdrop
            AsyncImage(
                model = ImageRequest.Builder(context)
                    .data(media.backdropPath)
                    .crossfade(true)
                    .build(),
                contentDescription = null,
                contentScale = ContentScale.Crop,
                modifier = Modifier.fillMaxSize(),
                alpha = 0.3f
            )

            Row(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(48.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                // Poster
                AsyncImage(
                    model = ImageRequest.Builder(context)
                        .data(media.posterPath)
                        .crossfade(true)
                        .build(),
                    contentDescription = "Poster",
                    modifier = Modifier
                        .width(300.dp)
                        .height(450.dp)
                )

                Spacer(modifier = Modifier.width(48.dp))

                // Info & Actions
                Column(
                    modifier = Modifier.weight(1f)
                ) {
                    Text(
                        text = media.title,
                        fontSize = 48.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color.White
                    )

                    Spacer(modifier = Modifier.height(16.dp))

                    Text(
                        text = "${media.releaseDate?.take(4) ?: ""} | ${media.duration?.let { (it / 60).toLong() } ?: "0"} min | ⭐ ${String.format("%.1f", media.rating ?: 0f)}",
                        fontSize = 18.sp,
                        color = Color.LightGray
                    )

                    Spacer(modifier = Modifier.height(24.dp))

                    Text(
                        text = media.overview ?: "No overview available.",
                        fontSize = 16.sp,
                        color = Color.White,
                        lineHeight = 24.sp
                    )

                    Spacer(modifier = Modifier.height(48.dp))

                    Button(
                        onClick = { onNavigateToPlayer(media.id) },
                        colors = ButtonDefaults.colors(
                            containerColor = Color.Red,
                            contentColor = Color.White
                        ),
                        modifier = Modifier
                            .fillMaxWidth(0.5f)
                            .height(56.dp)
                    ) {
                        Text(
                            text = "PLAY",
                            fontSize = 20.sp,
                            fontWeight = FontWeight.Bold
                        )
                    }
                }
            }
        }

        if (uiState.isLoading) {
            Text(
                text = "Loading...",
                color = Color.White,
                modifier = Modifier.align(Alignment.Center)
            )
        }
    }
}
