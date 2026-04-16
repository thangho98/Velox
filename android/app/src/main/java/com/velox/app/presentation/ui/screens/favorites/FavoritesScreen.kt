package com.velox.app.presentation.ui.screens.favorites

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.velox.app.presentation.ui.components.ResponsiveImage
import com.velox.app.R
import com.velox.app.domain.model.MediaItem
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.ui.components.NotificationBell
import com.velox.app.presentation.viewmodel.FavoritesUiState
import com.velox.app.presentation.viewmodel.FavoritesViewModel
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixGray
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.FavoritePink
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

@Composable
fun FavoritesScreen(
    onMediaClick: (Int) -> Unit,
    onSeriesClick: (Int) -> Unit,
    onSearchClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    onSettingsClick: () -> Unit,
    viewModel: FavoritesViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    FavoritesScreenContent(
        uiState = uiState,
        onMediaClick = onMediaClick,
        onSeriesClick = onSeriesClick,
        onSearchClick = onSearchClick,
        onNotificationsClick = onNotificationsClick,
        onSettingsClick = onSettingsClick,
        onRetry = { viewModel.refresh() },
        onToggleFavorite = { viewModel.toggleFavorite(it) }
    )
}

@Composable
private fun FavoritesScreenContent(
    uiState: FavoritesUiState,
    onMediaClick: (Int) -> Unit,
    onSeriesClick: (Int) -> Unit,
    onSearchClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    onSettingsClick: () -> Unit,
    onRetry: () -> Unit,
    onToggleFavorite: (Int) -> Unit,
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.nav_favorites), color = NetflixWhite, fontWeight = FontWeight.Bold) },
                actions = {
                    NotificationBell(onClick = onNotificationsClick)
                    IconButton(onClick = onSearchClick) {
                        Icon(LucideIcons.Search, contentDescription = "Search", tint = NetflixWhite)
                    }
                    IconButton(onClick = onSettingsClick) {
                        Icon(LucideIcons.Settings, contentDescription = "Settings", tint = NetflixWhite)
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = NetflixBlack),
            )
        },
        containerColor = NetflixBlack,
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .background(NetflixBlack)
                .padding(padding),
        ) {
            if (uiState.isLoading) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator(color = NetflixRed)
                }
            } else if (uiState.error != null) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            text = uiState.error ?: "An error occurred",
                            color = MaterialTheme.colorScheme.error,
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Button(
                            onClick = onRetry,
                            colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                        ) {
                            Text(stringResource(R.string.action_retry))
                        }
                    }
                }
            } else if (uiState.favorites.isEmpty()) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            text = "No favorites yet",
                            color = NetflixWhite,
                            fontSize = 18.sp,
                            fontWeight = FontWeight.Medium,
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = "Add movies and series to your favorites\nto see them here",
                            color = NetflixLightGray,
                            fontSize = 14.sp,
                        )
                    }
                }
            } else {
                LazyVerticalGrid(
                    columns = GridCells.Adaptive(minSize = 100.dp),
                    contentPadding = PaddingValues(16.dp),
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                    verticalArrangement = Arrangement.spacedBy(16.dp),
                ) {
                    items(uiState.favorites) { item ->
                        FavoriteCard(
                            item = item,
                            onClick = { onMediaClick(item.id) },
                            onRemove = { onToggleFavorite(item.id) },
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun FavoriteCard(
    item: MediaItem,
    onClick: () -> Unit,
    onRemove: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
    ) {
        Box {
            Surface(
                modifier = Modifier
                    .fillMaxWidth()
                    .aspectRatio(2f / 3f),
                color = NetflixGray,
                shape = RoundedCornerShape(8.dp),
            ) {
                if (item.poster != null || item.posterPath != null) {
                    if (item.poster != null) {
                        ResponsiveImage(
                            data = item.poster,
                            contentDescription = item.title,
                            modifier = Modifier.fillMaxSize(),
                            contentScale = ContentScale.Crop,
                        )
                    } else {
                        AsyncImage(
                            model = item.posterPath,
                            contentDescription = item.title,
                            modifier = Modifier.fillMaxSize(),
                            contentScale = ContentScale.Crop,
                        )
                    }
                } else {
                    Box(
                        contentAlignment = Alignment.Center,
                    ) {
                        Text(
                            text = item.title.take(2).uppercase(),
                            color = NetflixLightGray,
                            fontSize = 24.sp,
                        )
                    }
                }
            }
            // Media Type Badge
            com.velox.app.presentation.ui.components.MediaBadge(
                mediaType = item.mediaType,
                modifier = Modifier.align(Alignment.TopStart)
            )
            // Favorite indicator
            Surface(
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .padding(8.dp)
                    .size(32.dp),
                color = FavoritePink,
                shape = RoundedCornerShape(16.dp),
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Icon(
                        imageVector = LucideIcons.Heart,
                        contentDescription = "Favorite",
                        tint = NetflixWhite,
                        modifier = Modifier.size(16.dp)
                    )
                }
            }
        }
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = item.title,
            color = NetflixWhite,
            fontSize = 14.sp,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        item.year?.let {
            Text(
                text = it.toString(),
                color = NetflixLightGray,
                fontSize = 12.sp,
            )
        }
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun FavoritesScreenLoadingPreview() {
    VeloxTheme {
        FavoritesScreenContent(
            uiState = FavoritesUiState(isLoading = true, favorites = emptyList(), error = null),
            onMediaClick = {},
            onSeriesClick = {},
            onSearchClick = {},
            onNotificationsClick = {},
            onSettingsClick = {},
            onRetry = {},
            onToggleFavorite = {}
        )
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun FavoritesScreenErrorPreview() {
    VeloxTheme {
        FavoritesScreenContent(
            uiState = FavoritesUiState(isLoading = false, favorites = emptyList(), error = "Failed to load favorites"),
            onMediaClick = {},
            onSeriesClick = {},
            onSearchClick = {},
            onNotificationsClick = {},
            onSettingsClick = {},
            onRetry = {},
            onToggleFavorite = {}
        )
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun FavoritesScreenEmptyPreview() {
    VeloxTheme {
        FavoritesScreenContent(
            uiState = FavoritesUiState(isLoading = false, favorites = emptyList(), error = null),
            onMediaClick = {},
            onSeriesClick = {},
            onSearchClick = {},
            onNotificationsClick = {},
            onSettingsClick = {},
            onRetry = {},
            onToggleFavorite = {}
        )
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun FavoritesScreenWithItemsPreview() {
    VeloxTheme {
        FavoritesScreenContent(
            uiState = FavoritesUiState(
                isLoading = false,
                favorites = listOf(
                    MediaItem(id = 1, title = "Inception", posterPath = null, backdropPath = null, year = 2010, rating = 8.8f, mediaType = "movie", overview = "A thief who steals corporate secrets."),
                    MediaItem(id = 2, title = "The Matrix", posterPath = null, backdropPath = null, year = 1999, rating = 8.7f, mediaType = "movie", overview = "A computer hacker."),
                    MediaItem(id = 3, title = "Interstellar", posterPath = null, backdropPath = null, year = 2014, rating = 8.6f, mediaType = "movie", overview = "A team of explorers."),
                ),
                error = null
            ),
            onMediaClick = {},
            onSeriesClick = {},
            onSearchClick = {},
            onNotificationsClick = {},
            onSettingsClick = {},
            onRetry = {},
            onToggleFavorite = {}
        )
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun FavoriteCardPreview() {
    VeloxTheme {
        Box(modifier = Modifier.background(NetflixBlack)) {
            FavoriteCard(
                item = MediaItem(
                    id = 1,
                    title = "Inception",
                    posterPath = null,
                    backdropPath = null,
                    year = 2010,
                    rating = 8.8f,
                    mediaType = "movie",
                    overview = "A thief who steals corporate secrets.",
                ),
                onClick = {},
                onRemove = {}
            )
        }
    }
}
