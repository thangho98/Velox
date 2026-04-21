package com.velox.app.presentation.tv.screens

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.Text
import com.velox.app.presentation.tv.components.TvMediaCard
import com.velox.app.presentation.viewmodel.HomeViewModel

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun TvHomeScreen(
    onNavigateToMedia: (mediaId: Int) -> Unit,
    onNavigateToSeries: (seriesId: Int) -> Unit,
    viewModel: HomeViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(bottom = 32.dp)
    ) {
        item {
            Column(modifier = Modifier.padding(start = 32.dp, top = 32.dp, bottom = 16.dp)) {
                Text(
                    text = "VELOX",
                    fontSize = 32.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color.Red
                )
            }
        }

        if (uiState.continueWatching.isNotEmpty()) {
            item {
                TvCarouselSection(
                    title = "Continue Watching",
                    items = uiState.continueWatching,
                    onItemClick = { onNavigateToMedia(it.mediaId) }
                )
            }
        }

        if (uiState.recentlyAddedMovies.isNotEmpty()) {
            item {
                TvCarouselSection(
                    title = "Recently Added Movies",
                    items = uiState.recentlyAddedMovies,
                    onItemClick = { onNavigateToMedia(it.id) }
                )
            }
        }

        if (uiState.recentlyAddedSeries.isNotEmpty()) {
            item {
                TvCarouselSection(
                    title = "Recently Added Series",
                    items = uiState.recentlyAddedSeries,
                    onItemClick = { onNavigateToSeries(it.id) },
                    isLandscape = false
                )
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun <T> TvCarouselSection(
    title: String,
    items: List<T>,
    onItemClick: (T) -> Unit,
    isLandscape: Boolean = true
) {
    Column(modifier = Modifier.padding(bottom = 24.dp)) {
        Text(
            text = title,
            fontSize = 20.sp,
            fontWeight = FontWeight.SemiBold,
            color = Color.White,
            modifier = Modifier.padding(start = 32.dp, bottom = 8.dp)
        )
        LazyRow(
            contentPadding = PaddingValues(horizontal = 24.dp),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(items) { item ->
                when (item) {
                    is com.velox.app.domain.model.ContinueWatchingItem -> {
                        val progress = if (item.duration != null && item.duration > 0) item.position / item.duration else 0f
                        TvMediaCard(
                            title = item.seriesTitle ?: item.title,
                            posterPath = item.backdropPath ?: item.posterPath,
                            onClick = { onItemClick(item as T) },
                            isLandscape = true
                        )
                    }
                    is com.velox.app.domain.model.MediaItem -> {
                        TvMediaCard(
                            title = item.title,
                            posterPath = item.backdropPath ?: item.posterPath,
                            onClick = { onItemClick(item as T) },
                            isLandscape = isLandscape
                        )
                    }
                    is com.velox.app.domain.model.SeriesItem -> {
                        TvMediaCard(
                            title = item.title,
                            posterPath = item.posterPath,
                            onClick = { onItemClick(item as T) },
                            isLandscape = false
                        )
                    }
                }
            }
        }
    }
}
