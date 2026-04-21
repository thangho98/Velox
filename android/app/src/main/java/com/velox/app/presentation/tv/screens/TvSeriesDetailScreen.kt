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
import com.velox.app.presentation.viewmodel.SeriesDetailViewModel

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun TvSeriesDetailScreen(
    seriesId: Int,
    onNavigateToPlayer: (Int) -> Unit,
    viewModel: SeriesDetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val context = LocalContext.current

    Box(modifier = Modifier.fillMaxSize()) {
        uiState.series?.let { series ->
            // Background Backdrop
            AsyncImage(
                model = ImageRequest.Builder(context)
                    .data(series.backdropPath ?: series.posterPath)
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
                        .data(series.posterPath)
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
                        text = series.title,
                        fontSize = 48.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color.White
                    )

                    Spacer(modifier = Modifier.height(16.dp))

                    Text(
                        text = "${series.seasons.size} Seasons",
                        fontSize = 18.sp,
                        color = Color.LightGray
                    )

                    Spacer(modifier = Modifier.height(24.dp))

                    Text(
                        text = series.overview ?: "No overview available.",
                        fontSize = 16.sp,
                        color = Color.White,
                        lineHeight = 24.sp
                    )

                    Spacer(modifier = Modifier.height(48.dp))

                    val firstEpisodeId = uiState.episodes.firstOrNull()?.mediaId
                    if (firstEpisodeId != null) {
                        Button(
                            onClick = { onNavigateToPlayer(firstEpisodeId) },
                            colors = ButtonDefaults.colors(
                                containerColor = Color.Red,
                                contentColor = Color.White
                            ),
                            modifier = Modifier
                                .fillMaxWidth(0.5f)
                                .height(56.dp)
                        ) {
                            Text(
                                text = "PLAY FIRST EPISODE",
                                fontSize = 20.sp,
                                fontWeight = FontWeight.Bold
                            )
                        }
                    } else if (uiState.areEpisodesLoading) {
                        Text(
                            text = "Loading episodes...",
                            color = Color.Gray,
                            fontSize = 16.sp,
                        )
                    } else {
                        Text(
                            text = "No episodes available.",
                            color = Color.Gray,
                            fontSize = 16.sp,
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
