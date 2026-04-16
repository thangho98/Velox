package com.velox.app.presentation.ui.screens.home

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.velox.app.presentation.ui.components.ResponsiveImage
import coil.compose.AsyncImage
import com.velox.app.R
import com.velox.app.domain.model.ContinueWatchingItem
import com.velox.app.domain.model.Library
import com.velox.app.domain.model.MediaItem
import com.velox.app.domain.model.NextUpItem
import com.velox.app.domain.model.User
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.ui.components.NotificationBell
import com.velox.app.presentation.viewmodel.HomeUiState
import com.velox.app.presentation.viewmodel.HomeViewModel
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixGray
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

@Composable
fun HomeScreen(
    onMediaClick: (Int) -> Unit,
    onPlayClick: (Int) -> Unit,
    onSeriesClick: (Int) -> Unit,
    onSearchClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    onSettingsClick: () -> Unit,
    onNavigateToMovies: () -> Unit,
    onNavigateToSeries: () -> Unit,
    viewModel: HomeViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    HomeContent(
        uiState = uiState,
        onMediaClick = onMediaClick,
        onPlayClick = onPlayClick,
        onSeriesClick = onSeriesClick,
        onSearchClick = onSearchClick,
        onNotificationsClick = onNotificationsClick,
        onSettingsClick = onSettingsClick,
        onNavigateToMovies = onNavigateToMovies,
        onNavigateToSeries = onNavigateToSeries,
        onDismissContinueWatching = { viewModel.dismissContinueWatching(it) }
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeContent(
    uiState: HomeUiState,
    onMediaClick: (Int) -> Unit,
    onPlayClick: (Int) -> Unit,
    onSeriesClick: (Int) -> Unit,
    onSearchClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    onSettingsClick: () -> Unit,
    onNavigateToMovies: () -> Unit,
    onNavigateToSeries: () -> Unit,
    onDismissContinueWatching: (Int) -> Unit,
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        text = "VELOX",
                        color = NetflixRed,
                        fontWeight = FontWeight.Bold,
                    )
                },
                actions = {
                    NotificationBell(onClick = onNotificationsClick)
                    IconButton(onClick = onSearchClick) {
                        Icon(LucideIcons.Search, contentDescription = "Search", tint = NetflixWhite)
                    }
                    IconButton(onClick = onSettingsClick) {
                        Icon(LucideIcons.Settings, contentDescription = "Settings", tint = NetflixWhite)
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = NetflixBlack,
                ),
            )
        },
        containerColor = NetflixBlack,
    ) { padding ->
        if (uiState.isLoading) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentAlignment = Alignment.Center,
            ) {
                CircularProgressIndicator(color = NetflixRed)
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentPadding = PaddingValues(vertical = 16.dp),
            ) {
                // Welcome section
                item {
                    val screenWidth = LocalConfiguration.current.screenWidthDp
                    Column(
                        modifier = Modifier.padding(horizontal = 16.dp),
                    ) {
                        val greeting = uiState.user?.displayName?.let { ", $it" } ?: ""
                        Text(
                            text = stringResource(R.string.home_welcome_back) + greeting,
                            color = NetflixWhite,
                            fontSize = if (screenWidth < 600) 22.sp else 26.sp,
                            fontWeight = FontWeight.Bold,
                        )
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = stringResource(R.string.home_personal_media_server),
                            color = NetflixLightGray,
                            fontSize = 14.sp,
                        )
                        Spacer(modifier = Modifier.height(20.dp))
                        Row(
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Button(
                                onClick = onNavigateToMovies,
                                modifier = Modifier.weight(1f).height(48.dp),
                                shape = androidx.compose.foundation.shape.RoundedCornerShape(8.dp),
                                colors = ButtonDefaults.buttonColors(
                                    containerColor = NetflixRed,
                                    contentColor = NetflixWhite
                                )
                            ) {
                                Text(stringResource(R.string.home_movies), fontSize = 15.sp, fontWeight = FontWeight.SemiBold)
                            }
                            OutlinedButton(
                                onClick = onNavigateToSeries,
                                modifier = Modifier.weight(1f).height(48.dp),
                                shape = androidx.compose.foundation.shape.RoundedCornerShape(8.dp),
                                border = androidx.compose.foundation.BorderStroke(1.dp, Color.White.copy(alpha = 0.2f)),
                                colors = ButtonDefaults.outlinedButtonColors(
                                    contentColor = NetflixWhite
                                )
                            ) {
                                Text(stringResource(R.string.home_series), fontSize = 15.sp, fontWeight = FontWeight.SemiBold)
                            }
                        }
                    }
                    Spacer(modifier = Modifier.height(24.dp))
                }

                // Continue Watching section
                if (uiState.continueWatching.isNotEmpty()) {
                    item {
                        SectionHeader(title = stringResource(R.string.home_continue_watching))
                    }
                    item {
                        LazyRow(
                            contentPadding = PaddingValues(horizontal = 16.dp),
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            items(uiState.continueWatching) { item ->
                                ContinueWatchingCard(
                                    item = item,
                                    onClick = { onPlayClick(item.mediaId) },
                                    onDismiss = { onDismissContinueWatching(item.mediaId) },
                                )
                            }
                        }
                        Spacer(modifier = Modifier.height(24.dp))
                    }
                }

                // Next Up section
                if (uiState.nextUp.isNotEmpty()) {
                    item {
                        SectionHeader(
                            title = stringResource(R.string.home_next_up),
                            onSeeAll = onNavigateToSeries
                        )
                    }
                    item {
                        LazyRow(
                            contentPadding = PaddingValues(horizontal = 16.dp),
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            items(uiState.nextUp) { item ->
                                NextUpCard(
                                    item = item,
                                    onClick = { onPlayClick(item.mediaId) },
                                )
                            }
                        }
                        Spacer(modifier = Modifier.height(24.dp))
                    }
                }

                // Recently Added - Movies
                if (uiState.recentlyAddedMovies.isNotEmpty()) {
                    item {
                        SectionHeader(
                            title = "Movies",
                            onSeeAll = onNavigateToMovies
                        )
                    }
                    item {
                        LazyRow(
                            contentPadding = PaddingValues(horizontal = 16.dp),
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            items(uiState.recentlyAddedMovies) { item ->
                                MediaCard(
                                    item = item,
                                    onClick = { onMediaClick(item.id) },
                                )
                            }
                        }
                        Spacer(modifier = Modifier.height(24.dp))
                    }
                }

                // Recently Added - Series
                if (uiState.recentlyAddedSeries.isNotEmpty()) {
                    item {
                        SectionHeader(
                            title = "Series",
                            onSeeAll = onNavigateToSeries
                        )
                    }
                    item {
                        LazyRow(
                            contentPadding = PaddingValues(horizontal = 16.dp),
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            items(uiState.recentlyAddedSeries) { item ->
                                SeriesCard(
                                    title = item.title,
                                    posterPath = item.posterPath,
                                    poster = item.poster,
                                    onClick = { onSeriesClick(item.id) },
                                )
                            }
                        }
                    }
                }

                // Show Libraries when no movies/series content
                if (uiState.recentlyAddedMovies.isEmpty() && uiState.recentlyAddedSeries.isEmpty() && uiState.libraries.isNotEmpty()) {
                    item {
                        Spacer(modifier = Modifier.height(16.dp))
                    }
                    item {
                        SectionHeader(title = "Your Libraries")
                    }
                    item {
                        LazyRow(
                            contentPadding = PaddingValues(horizontal = 16.dp),
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            items(uiState.libraries) { library ->
                                LibraryCard(
                                    library = library,
                                    onClick = { /* Navigate to browse with library */ },
                                )
                            }
                        }
                    }
                }

                // Error state
                uiState.error?.let { error ->
                    item {
                        Spacer(modifier = Modifier.height(16.dp))
                        Text(
                            text = error,
                            color = MaterialTheme.colorScheme.error,
                            modifier = Modifier.padding(horizontal = 16.dp),
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun SectionHeader(
    title: String,
    onSeeAll: (() -> Unit)? = null
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            text = title,
            color = NetflixWhite,
            fontSize = 18.sp,
            fontWeight = FontWeight.SemiBold,
        )
        if (onSeeAll != null) {
            Text(
                text = "See all",
                color = NetflixRed,
                fontSize = 14.sp,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.clickable(onClick = onSeeAll)
            )
        }
    }
}

@Composable
fun MediaCard(
    item: MediaItem,
    onClick: () -> Unit,
) {
    val screenWidth = LocalConfiguration.current.screenWidthDp
    val cardWidth = if (screenWidth < 600) 100.dp else 120.dp
    val cardHeight = if (screenWidth < 600) 140.dp else 168.dp

    Surface(
        modifier = Modifier
            .width(cardWidth)
            .height(cardHeight)
            .clickable(onClick = onClick),
        color = Color(0xFF1F1F1F),
        shape = androidx.compose.foundation.shape.RoundedCornerShape(8.dp),
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
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
                    modifier = Modifier.fillMaxSize().background(Color(0xFF2A2A2A))
                ) {
                    Text(text = item.title.take(1).uppercase(), color = Color(0xFF444444), fontSize = 24.sp, fontWeight = FontWeight.Bold)
                }
            }

            // Bottom gradient with title
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .align(Alignment.BottomCenter)
                    .background(
                        brush = androidx.compose.ui.graphics.Brush.verticalGradient(
                            colors = listOf(Color.Transparent, Color(0x99000000), Color(0xCC000000))
                        )
                    )
                    .padding(horizontal = 10.dp, vertical = 10.dp)
            ) {
                Text(
                    text = item.title,
                    color = Color.White,
                    fontSize = 13.sp,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 2,
                    lineHeight = 16.sp
                )
                // Progress bar
                val position = item.position ?: 0f
                val duration = item.duration ?: 0f
                val hasProgress = position > 0 && duration > 0 && item.completed != true
                val isCompleted = item.completed == true
                if (hasProgress || isCompleted) {
                    val progressPercent = if (isCompleted) 100f else (position / duration * 100f).coerceIn(0f, 100f)
                    Spacer(modifier = Modifier.height(4.dp))
                    LinearProgressIndicator(
                        progress = { progressPercent / 100f },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(3.dp),
                        color = NetflixRed,
                        trackColor = Color(0xFF4B5563),
                        drawStopIndicator = {}
                    )
                    Spacer(modifier = Modifier.height(2.dp))
                    Text(
                        text = "${progressPercent.toInt()}% watched",
                        color = Color(0xFFD1D5DB),
                        fontSize = 10.sp,
                    )
                }
            }
        }
    }
}

@Composable
fun SeriesCard(
    title: String,
    posterPath: String?,
    poster: com.velox.app.domain.model.ImageResource? = null,
    onClick: () -> Unit,
) {
    val screenWidth = LocalConfiguration.current.screenWidthDp
    val cardWidth = if (screenWidth < 600) 100.dp else 120.dp
    val cardHeight = if (screenWidth < 600) 140.dp else 168.dp

    Surface(
        modifier = Modifier
            .width(cardWidth)
            .height(cardHeight)
            .clickable(onClick = onClick),
        color = Color(0xFF1F1F1F),
        shape = androidx.compose.foundation.shape.RoundedCornerShape(8.dp),
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
            if (poster != null || posterPath != null) {
                if (poster != null) {
                    ResponsiveImage(
                        data = poster,
                        contentDescription = title,
                        modifier = Modifier.fillMaxSize(),
                        contentScale = ContentScale.Crop,
                    )
                } else {
                    AsyncImage(
                        model = posterPath,
                        contentDescription = title,
                        modifier = Modifier.fillMaxSize(),
                        contentScale = ContentScale.Crop,
                    )
                }
            } else {
                Box(
                    contentAlignment = Alignment.Center,
                    modifier = Modifier.fillMaxSize().background(Color(0xFF2A2A2A))
                ) {
                    Text(text = title.take(1).uppercase(), color = Color(0xFF444444), fontSize = 24.sp, fontWeight = FontWeight.Bold)
                }
            }

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .align(Alignment.BottomCenter)
                    .background(
                        brush = androidx.compose.ui.graphics.Brush.verticalGradient(
                            colors = listOf(Color.Transparent, Color(0x99000000), Color(0xCC000000))
                        )
                    )
                    .padding(horizontal = 10.dp, vertical = 10.dp)
            ) {
                Text(
                    text = title,
                    color = Color.White,
                    fontSize = 13.sp,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 2,
                    lineHeight = 16.sp
                )
            }
        }
    }
}

@Composable
fun ContinueWatchingCard(
    item: ContinueWatchingItem,
    onClick: () -> Unit,
    onDismiss: () -> Unit,
) {
    val screenWidth = LocalConfiguration.current.screenWidthDp
    val cardWidth = if (screenWidth < 600) 200.dp else 240.dp
    val cardHeight = if (screenWidth < 600) 112.dp else 135.dp

    Box(
        modifier = Modifier
            .width(cardWidth)
            .height(cardHeight)
    ) {
        Surface(
            modifier = Modifier
                .fillMaxSize()
                .clickable(onClick = onClick),
            color = Color(0xFF1F1F1F),
            shape = androidx.compose.foundation.shape.RoundedCornerShape(8.dp),
        ) {
            // Prefer backdrop for horizontal 16:9 ratio
            val imagePath = item.backdropPath ?: item.posterPath
            Box(modifier = Modifier.fillMaxSize()) {
                val imageResource = item.backdrop ?: item.poster

                if (imageResource != null || imagePath != null) {
                    if (imageResource != null) {
                        ResponsiveImage(
                            data = imageResource,
                            contentDescription = item.title,
                            modifier = Modifier.fillMaxSize(),
                            contentScale = ContentScale.Crop,
                        )
                    } else {
                        AsyncImage(
                            model = imagePath,
                            contentDescription = item.title,
                            modifier = Modifier.fillMaxSize(),
                            contentScale = ContentScale.Crop,
                        )
                    }
                } else {
                    Box(
                        contentAlignment = Alignment.Center,
                        modifier = Modifier.fillMaxSize().background(Color(0xFF2A2A2A))
                    ) {
                        Text(text = item.title.take(1).uppercase(), color = Color(0xFF444444), fontSize = 24.sp, fontWeight = FontWeight.Bold)
                    }
                }

                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .align(Alignment.BottomCenter)
                        .background(
                            brush = androidx.compose.ui.graphics.Brush.verticalGradient(
                                colors = listOf(Color.Transparent, Color(0x99000000), Color(0xCC000000))
                            )
                        )
                ) {
                    Column(
                        modifier = Modifier.padding(10.dp)
                    ) {
                        val displayTitle = if (item.mediaType == "episode" && item.seriesTitle != null) {
                            "S${item.seasonNumber}E${item.episodeNumber} • ${item.seriesTitle}"
                        } else {
                            item.title
                        }

                        Text(
                            text = displayTitle,
                            color = Color.White,
                            fontSize = 12.sp,
                            fontWeight = FontWeight.SemiBold,
                            maxLines = 2,
                            lineHeight = 14.sp
                        )
                        val remainingMinutes = if (item.duration != null && item.duration > 0) {
                            Math.ceil(((item.duration - item.position) / 60).toDouble()).toInt()
                        } else 0
                        if (remainingMinutes > 0) {
                            Text(
                                text = "${remainingMinutes}m remaining",
                                color = Color(0xFFAAAAAA),
                                fontSize = 10.sp,
                            )
                        }
                    }

                    val progress = if (item.duration != null && item.duration > 0) item.position / item.duration else 0f
                    if (progress > 0f) {
                        LinearProgressIndicator(
                            progress = { progress },
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(3.dp)
                                .align(Alignment.BottomCenter),
                            color = NetflixRed,
                            trackColor = Color(0x33FFFFFF),
                            drawStopIndicator = {}
                        )
                    }
                }
            }
        }

        // Dismiss button
        IconButton(
            onClick = onDismiss,
            modifier = Modifier
                .align(Alignment.TopEnd)
                .size(24.dp)
                .padding(4.dp),
        ) {
            Surface(
                shape = CircleShape,
                color = Color(0xB3000000),
            ) {
                Icon(
                    LucideIcons.Close,
                    contentDescription = "Dismiss",
                    tint = NetflixWhite,
                    modifier = Modifier.padding(2.dp),
                )
            }
        }
    }
}

@Composable
fun LibraryCard(
    library: Library,
    onClick: () -> Unit,
) {
    val screenWidth = LocalConfiguration.current.screenWidthDp
    val cardWidth = if (screenWidth < 600) 140.dp else 160.dp

    Column(
        modifier = Modifier
            .width(cardWidth)
            .clickable(onClick = onClick),
    ) {
        Surface(
            modifier = Modifier
                .fillMaxWidth()
                .height(100.dp),
            color = NetflixDark,
            shape = MaterialTheme.shapes.small,
        ) {
            Box(
                contentAlignment = Alignment.Center,
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(text = if (library.type == "movie") "🎬" else "📺", fontSize = 24.sp)
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = library.name,
                        color = NetflixWhite,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Medium,
                    )
                }
            }
        }
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = "${library.itemCount ?: 0} items",
            color = NetflixLightGray,
            fontSize = 12.sp,
        )
    }
}

@Composable
fun NextUpCard(
    item: NextUpItem,
    onClick: () -> Unit,
) {
    val screenWidth = LocalConfiguration.current.screenWidthDp
    val cardWidth = if (screenWidth < 600) 200.dp else 240.dp
    val cardHeight = if (screenWidth < 600) 112.dp else 135.dp

    Surface(
        modifier = Modifier
            .width(cardWidth)
            .height(cardHeight)
            .clickable(onClick = onClick),
        color = Color(0xFF1F1F1F),
        shape = androidx.compose.foundation.shape.RoundedCornerShape(8.dp),
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
            // Prefer stillPath (episode thumb) or backdropPath for horizontal ratio
            val imagePath = item.stillPath ?: item.backdropPath ?: item.seriesPoster
            val imageResource = item.still ?: item.backdrop ?: item.poster

            if (imageResource != null || imagePath != null) {
                if (imageResource != null) {
                    ResponsiveImage(
                        data = imageResource,
                        contentDescription = item.title,
                        modifier = Modifier.fillMaxSize(),
                        contentScale = ContentScale.Crop,
                    )
                } else {
                    AsyncImage(
                        model = imagePath,
                        contentDescription = item.title,
                        modifier = Modifier.fillMaxSize(),
                        contentScale = ContentScale.Crop,
                    )
                }
            } else {
                Box(
                    contentAlignment = Alignment.Center,
                    modifier = Modifier.fillMaxSize().background(Color(0xFF2A2A2A))
                ) {
                    Text(text = (item.seriesTitle ?: item.title).take(1).uppercase(), color = Color(0xFF444444), fontSize = 24.sp, fontWeight = FontWeight.Bold)
                }
            }

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .align(Alignment.BottomCenter)
                    .background(
                        brush = androidx.compose.ui.graphics.Brush.verticalGradient(
                            colors = listOf(Color.Transparent, Color(0x99000000), Color(0xCC000000))
                        )
                    )
                    .padding(horizontal = 10.dp, vertical = 10.dp)
            ) {
                Column {
                    Text(
                        text = "S${item.seasonNumber}E${item.episodeNumber} • ${item.seriesTitle ?: item.title}",
                        color = Color.White,
                        fontSize = 12.sp,
                        fontWeight = FontWeight.SemiBold,
                        maxLines = 2,
                        lineHeight = 14.sp
                    )
                }
            }
        }
    }
}

@Preview(showBackground = true)
@Composable
fun HomeScreenPreview() {
    VeloxTheme {
        HomeContent(
            uiState = SampleHomeUiState,
            onMediaClick = {},
            onPlayClick = {},
            onSeriesClick = {},
            onSearchClick = {},
            onNotificationsClick = {},
            onSettingsClick = {},
            onNavigateToMovies = {},
            onNavigateToSeries = {},
            onDismissContinueWatching = {},
        )
    }
}

@Preview(showBackground = true)
@Composable
fun MediaCardPreview() {
    VeloxTheme {
        MediaCard(
            item = SampleMediaItem,
            onClick = {},
        )
    }
}

@Preview(showBackground = true)
@Composable
fun SeriesCardPreview() {
    VeloxTheme {
        SeriesCard(
            title = "The Last of Us",
            posterPath = null,
            poster = null,
            onClick = {},
        )
    }
}

@Preview(showBackground = true)
@Composable
fun ContinueWatchingCardPreview() {
    VeloxTheme {
        ContinueWatchingCard(
            item = SampleContinueWatchingItem,
            onClick = {},
            onDismiss = {},
        )
    }
}

@Preview(showBackground = true)
@Composable
fun NextUpCardPreview() {
    VeloxTheme {
        NextUpCard(
            item = SampleNextUpItem,
            onClick = {},
        )
    }
}

@Preview(showBackground = true)
@Composable
fun LibraryCardPreview() {
    VeloxTheme {
        LibraryCard(
            library = SampleLibrary,
            onClick = {},
        )
    }
}

private val SampleUser = User(
    id = 1,
    username = "johndoe",
    displayName = "John Doe",
    isAdmin = false,
)

private val SampleMediaItem = MediaItem(
    id = 1,
    title = "Inception",
    posterPath = null,
    backdropPath = null,
    year = 2010,
    rating = 8.8f,
    mediaType = "movie",
    overview = "A thief who steals corporate secrets through the use of dream-sharing technology.",
)

private val SampleContinueWatchingItem = ContinueWatchingItem(
    mediaId = 1,
    seriesId = null,
    position = 3600f,
    completed = false,
    title = "Inception",
    posterPath = null,
    backdropPath = null,
    mediaType = "movie",
    duration = 8880f,
    seriesTitle = null,
    seasonNumber = null,
    episodeNumber = null,
)

private val SampleNextUpItem = NextUpItem(
    mediaId = 2,
    seriesId = 1,
    title = "The Last of Us",
    episodeTitle = "Long, Long Time",
    stillPath = null,
    backdropPath = null,
    duration = 3600f,
    seasonNumber = 1,
    episodeNumber = 3,
    seriesTitle = "The Last of Us",
    seriesPoster = null,
)

private val SampleLibrary = Library(
    id = 1,
    name = "Movies",
    type = "movie",
    paths = listOf("/movies"),
    itemCount = 120,
)

private val SampleHomeUiState = HomeUiState(
    isLoading = false,
    user = SampleUser,
    continueWatching = listOf(SampleContinueWatchingItem),
    nextUp = listOf(SampleNextUpItem),
    recentlyAddedMovies = listOf(SampleMediaItem),
    recentlyAddedSeries = emptyList(),
    libraries = listOf(SampleLibrary),

)
