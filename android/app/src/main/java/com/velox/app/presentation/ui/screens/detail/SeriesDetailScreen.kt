package com.velox.app.presentation.ui.screens.detail

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.FavoriteBorder
import androidx.compose.material.icons.filled.Link
import androidx.compose.material.icons.filled.Lock
import androidx.compose.material.icons.filled.MoreVert
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Tv
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.Dialog
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import coil.compose.AsyncImage
import kotlinx.coroutines.launch
import com.velox.app.presentation.ui.components.ResponsiveImage
import com.velox.app.R
import com.velox.app.domain.model.Episode
import com.velox.app.domain.model.Season
import com.velox.app.domain.model.SeriesDetail
import com.velox.app.domain.model.WatchProgress
import com.velox.app.presentation.ui.components.ActionMenu
import com.velox.app.presentation.ui.components.ActionMenuButton
import com.velox.app.presentation.ui.components.ActionMenuItem
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.ui.components.YouTubePlayer
import com.velox.app.presentation.viewmodel.SeriesDetailUiState
import com.velox.app.presentation.viewmodel.SeriesDetailViewModel
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixGray
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SeriesDetailScreen(
    seriesId: Int,
    onBackClick: () -> Unit,
    onEpisodeClick: (Int) -> Unit,
    viewModel: SeriesDetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    SeriesDetailContent(
        uiState = uiState,
        onBackClick = onBackClick,
        onEpisodeClick = onEpisodeClick,
        onRefresh = { viewModel.refresh() },
        onToggleFavorite = { viewModel.toggleFavorite() },
        onSeasonSelect = { viewModel.selectSeason(it) },
        onAutoDownloadSubtitle = { mediaId, onResult ->
            viewModel.autoDownloadSubtitle(mediaId, onResult)
        },
        onDownloadSeries = { onResult -> viewModel.downloadSeries(onResult) },
        onDeleteSeriesDownload = { onResult -> viewModel.deleteSeriesDownload(onResult) },
        onDownloadEpisode = { mediaId, onResult -> viewModel.downloadEpisode(mediaId, onResult) },
        onDeleteEpisodeDownload = { mediaId, onResult -> viewModel.deleteEpisodeDownload(mediaId, onResult) },
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SeriesDetailContent(
    uiState: SeriesDetailUiState,
    onBackClick: () -> Unit,
    onEpisodeClick: (Int) -> Unit,
    onRefresh: () -> Unit,
    onToggleFavorite: () -> Unit,
    onSeasonSelect: (Season) -> Unit,
    onEditEpisode: ((Episode) -> Unit)? = null,
    onAutoDownloadSubtitle: (Int, (Boolean) -> Unit) -> Unit = { _, _ -> },
    onDownloadSeries: ((Boolean) -> Unit) -> Unit = {},
    onDeleteSeriesDownload: ((Boolean) -> Unit) -> Unit = {},
    onDownloadEpisode: (Int, (Boolean) -> Unit) -> Unit = { _, _ -> },
    onDeleteEpisodeDownload: (Int, (Boolean) -> Unit) -> Unit = { _, _ -> },
) {
    var currentTrailerIndex by remember { mutableIntStateOf(0) }
    var showTrailer by remember { mutableStateOf(false) }

    // YouTube trailer overlay
    if (showTrailer) {
        val trailers = uiState.cinema?.trailers ?: emptyList()
        if (trailers.isNotEmpty() && currentTrailerIndex < trailers.size) {
            val trailer = trailers[currentTrailerIndex]
            val videoKey = trailer.key ?: extractYouTubeKeyFromUrl(trailer.url)
            if (videoKey != null) {
                YouTubePlayer(
                    videoKey = videoKey,
                    title = trailer.name,
                    onClose = {
                        showTrailer = false
                        currentTrailerIndex = 0
                    },
                    onSkip = {
                        showTrailer = false
                        currentTrailerIndex = 0
                    },
                )
            }
        }
    }

    // Episode edit dialog
    var editingEpisode by remember { mutableStateOf<Episode?>(null) }
    editingEpisode?.let { episode ->
        EpisodeEditDialog(
            episode = episode,
            onClose = { editingEpisode = null },
            onSave = { episodeId, title, overview, episodeNumber, airDate ->
                // Build updated episode
                val updated = episode.copy(
                    title = title,
                    overview = overview,
                    episodeNumber = episodeNumber ?: episode.episodeNumber,
                    airDate = airDate
                )
                onEditEpisode?.invoke(updated)
                editingEpisode = null
            },
        )
    }

    Box(modifier = Modifier.fillMaxSize().background(NetflixBlack)) {
        val snackbarHostState = remember { SnackbarHostState() }
        val coroutineScope = rememberCoroutineScope()
        // Fixed Background Backdrop
        val series = uiState.series
        if (series != null && (series.backdropPath != null || series.backdrop != null)) {
            val imageResource = series.backdrop ?: series.poster
            if (imageResource != null) {
                ResponsiveImage(
                    data = imageResource,
                    contentDescription = null,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Crop,
                )
            } else {
                AsyncImage(
                    model = series.backdropPath,
                    contentDescription = null,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Crop,
                    alignment = Alignment.TopCenter
                )
            }
            // Vertical Gradient Overlay (Darkens the bottom)
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(
                        Brush.verticalGradient(
                            0.0f to Color(0x4D000000), // 30% black
                            0.5f to Color(0xCC000000), // 80% black
                            1.0f to NetflixBlack
                        )
                    )
            )
            // Horizontal Gradient Overlay (Cinematic left darkening)
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(
                        Brush.horizontalGradient(
                            0.0f to NetflixBlack,
                            0.4f to Color(0x80000000), // 50% black
                            1.0f to Color.Transparent
                        )
                    )
            )
        }

        Scaffold(
            snackbarHost = { SnackbarHost(hostState = snackbarHostState) },
            topBar = {
                TopAppBar(
                    title = { },
                    navigationIcon = {
                        Surface(
                            onClick = onBackClick,
                            shape = androidx.compose.foundation.shape.CircleShape,
                            color = Color(0x80000000), // Black 50%
                            modifier = Modifier.size(40.dp).padding(4.dp)
                        ) {
                            Box(contentAlignment = Alignment.Center) {
                                Icon(LucideIcons.ChevronLeft, contentDescription = "Back", tint = NetflixWhite, modifier = Modifier.size(24.dp))
                            }
                        }
                    },
                    colors = TopAppBarDefaults.topAppBarColors(containerColor = Color.Transparent),
                )
            },
            containerColor = Color.Transparent,
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
        } else if (uiState.error != null) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentAlignment = Alignment.Center,
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(
                        text = uiState.error,
                        color = MaterialTheme.colorScheme.error,
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    Button(
                        onClick = onRefresh,
                        colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                    ) {
                        Text(stringResource(R.string.action_retry))
                    }
                }
            }
        } else {
            val series = uiState.series ?: return@Scaffold

            PullToRefreshBox(
                isRefreshing = uiState.isRefreshing,
                onRefresh = onRefresh,
                modifier = Modifier.fillMaxSize()
            ) {
                LazyColumn(
                    modifier = Modifier.fillMaxSize()
                ) {
                    item {
                        Spacer(modifier = Modifier.height(padding.calculateTopPadding() + 40.dp))
                }

                // Central Poster Overlay
                item {
                    val screenWidth = LocalConfiguration.current.screenWidthDp
                    val posterWidth = if (screenWidth < 600) 180.dp else 260.dp

                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(bottom = 24.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        if (series.poster != null || series.posterPath != null) {
                            Surface(
                                modifier = Modifier
                                    .align(Alignment.Center)
                                    .padding(top = 20.dp)
                                    .width(posterWidth)
                                    .aspectRatio(2f / 3f),
                                shape = RoundedCornerShape(8.dp),
                                color = NetflixDark,
                                shadowElevation = 12.dp,
                            ) {
                                if (series.poster != null) {
                                    ResponsiveImage(
                                        data = series.poster,
                                        contentDescription = null,
                                        modifier = Modifier.fillMaxSize(),
                                        contentScale = ContentScale.Crop,
                                    )
                                } else {
                                    AsyncImage(
                                        model = series.posterPath,
                                        contentDescription = null,
                                        modifier = Modifier.fillMaxSize(),
                                        contentScale = ContentScale.Crop,
                                    )
                                }
                            }
                        }
                    }
                }

                // Info
                item {
                    val screenWidth = LocalConfiguration.current.screenWidthDp
                    val infoPadding = if (screenWidth < 600) 16.dp else 32.dp
                    Column(modifier = Modifier.padding(horizontal = infoPadding, vertical = 16.dp)) {
                        // Title row with badges
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                        ) {
                            Text(
                                text = series.title,
                                color = NetflixWhite,
                                fontSize = if (screenWidth < 600) 20.sp else 24.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.weight(1f),
                            )
                            // Lock badge (gold/yellow matching mobile)
                            if (series.metadataLocked) {
                                Surface(
                                    color = Color(0x26FBBF24), // rgba(251,191,36,0.15)
                                    shape = RoundedCornerShape(10.dp),
                                ) {
                                    Icon(
                                        LucideIcons.Lock,
                                        contentDescription = "Locked",
                                        tint = Color(0xFFFBBF24), // #fbbf24 gold
                                        modifier = Modifier
                                            .size(24.dp)
                                            .padding(4.dp),
                                    )
                                }
                            }
                            // Edit badge (matching mobile - clickable with text)
                            if (uiState.isAdmin == true) {
                                Surface(
                                    onClick = { /* TODO: Open metadata editor */ },
                                    color = Color(0x1AFFFFFF), // rgba(255,255,255,0.1)
                                    shape = RoundedCornerShape(4.dp),
                                ) {
                                    Row(
                                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 3.dp),
                                        verticalAlignment = Alignment.CenterVertically,
                                    ) {
                                        Icon(
                                            LucideIcons.Edit,
                                            contentDescription = null,
                                            tint = Color(0xFFCCCCCC),
                                            modifier = Modifier.size(11.dp),
                                        )
                                        Spacer(modifier = Modifier.width(4.dp))
                                        Text(
                                            text = "Edit",
                                            color = Color(0xFFCCCCCC),
                                            fontSize = 11.sp,
                                            fontWeight = FontWeight.Medium,
                                        )
                                    }
                                }
                            }
                        }

                        Spacer(modifier = Modifier.height(8.dp))

                        // Meta row: Year, Status, Network (matching mobile order)
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            series.firstAirDate?.take(4)?.let { year ->
                                Text(
                                    text = year,
                                    color = NetflixLightGray,
                                    fontSize = 14.sp,
                                )
                            }
                            // Status badge (matching mobile purple)
                            series.status?.let { status ->
                                Surface(
                                    color = Color(0x338A57D4), // rgba(168,85,247,0.2)
                                    shape = RoundedCornerShape(4.dp),
                                ) {
                                    Text(
                                        text = status,
                                        color = Color(0xFFC084FC), // #c084fc
                                        fontSize = 11.sp,
                                        fontWeight = FontWeight.SemiBold,
                                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 3.dp),
                                    )
                                }
                            }
                            // Network info with TV icon
                            series.network?.let { network ->
                                Row(verticalAlignment = Alignment.CenterVertically) {
                                    Icon(
                                        LucideIcons.Tv,
                                        contentDescription = null,
                                        tint = NetflixLightGray,
                                        modifier = Modifier.size(14.dp),
                                    )
                                    Spacer(modifier = Modifier.width(4.dp))
                                    Text(
                                        text = network,
                                        color = NetflixLightGray,
                                        fontSize = 14.sp,
                                    )
                                }
                            }

                            val firstEpisode = uiState.episodes.firstOrNull()
                            val topLevelSourceName = when {
                                firstEpisode?.filePath?.startsWith("ophim://", ignoreCase = true) == true -> "OPhim"
                                firstEpisode?.filePath?.startsWith("fshare://", ignoreCase = true) == true -> "Fshare"
                                firstEpisode != null -> "Local"
                                else -> "Unknown"
                            }
                            if (firstEpisode != null) {
                                Surface(
                                    color = Color(0x334B5563), // a discrete badge color
                                    shape = RoundedCornerShape(4.dp),
                                ) {
                                    Text(
                                        text = topLevelSourceName,
                                        color = Color(0xFF9CA3AF),
                                        fontSize = 11.sp,
                                        fontWeight = FontWeight.SemiBold,
                                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 3.dp),
                                    )
                                }
                            }
                        }
                        Spacer(modifier = Modifier.height(12.dp))

                        // Overview (Full text like WebApp)
                        if (series.overview != null) {
                            Text(
                                text = series.overview,
                                color = NetflixLightGray,
                                fontSize = if (screenWidth < 600) 14.sp else 16.sp,
                                lineHeight = if (screenWidth < 600) 20.sp else 24.sp,
                            )
                        }
                        Spacer(modifier = Modifier.height(24.dp))

                        // Play Button Block
                        if (uiState.episodes.isNotEmpty()) {
                            val firstEp = uiState.episodes.first()
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(12.dp)
                            ) {
                                Button(
                                    onClick = { onEpisodeClick(firstEp.mediaId) },
                                    colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                                    shape = RoundedCornerShape(4.dp),
                                    modifier = Modifier.height(44.dp)
                                ) {
                                    Icon(
                                        LucideIcons.PlayArrow,
                                        contentDescription = null,
                                        modifier = Modifier.size(18.dp)
                                    )
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text(
                                        text = "Resume",
                                        fontWeight = FontWeight.Bold,
                                        fontSize = 14.sp
                                    )
                                }

                                Text(
                                    text = "Continue ${series.title} - ${firstEp.title}",
                                    color = NetflixLightGray,
                                    fontSize = 14.sp
                                )

                                Box {
                                    val hasCloudMedia = uiState.episodes.any { ep ->
                                        ep.filePath?.contains("://") == true && !ep.filePath.startsWith("http")
                                    }
                                    val hasDownloadedFiles = uiState.episodes.any { ep ->
                                        ep.filePath?.contains("library/downloads") == true && !ep.filePath.contains("://")
                                    }
                                    var showSeriesActionMenu by remember { mutableStateOf(false) }
                                    var isDownloadingSeries by remember { mutableStateOf(false) }
                                    ActionMenuButton(
                                        expanded = showSeriesActionMenu,
                                        onClick = { showSeriesActionMenu = true },
                                    )
                                    ActionMenu(
                                        expanded = showSeriesActionMenu,
                                        onDismiss = { showSeriesActionMenu = false },
                                        items = buildList {
                                            if (uiState.isAdmin) {
                                                add(ActionMenuItem.RefreshMetadata(onClick = { showSeriesActionMenu = false; onRefresh() }))
                                            }
                                            if (hasCloudMedia && uiState.isAdmin) {
                                                add(ActionMenuItem.Custom(
                                                    label = "Download Series to NAS",
                                                    icon = LucideIcons.Download,
                                                    isLoading = isDownloadingSeries,
                                                    autoDismiss = false,
                                                    onClick = {
                                                        if (isDownloadingSeries) return@Custom
                                                        isDownloadingSeries = true
                                                        coroutineScope.launch {
                                                            snackbarHostState.showSnackbar("Starting download to NAS...")
                                                        }
                                                        onDownloadSeries { success ->
                                                            isDownloadingSeries = false
                                                            showSeriesActionMenu = false
                                                            coroutineScope.launch {
                                                                snackbarHostState.showSnackbar(
                                                                    if (success) "Download started!"
                                                                    else "Failed to start download."
                                                                )
                                                            }
                                                        }
                                                    }
                                                ))
                                            }
                                            if (hasDownloadedFiles && uiState.isAdmin) {
                                                var isDeletingSeriesDownload by remember { mutableStateOf(false) }
                                                add(ActionMenuItem.Custom(
                                                    label = "Delete Series on NAS",
                                                    icon = LucideIcons.Delete,
                                                    isLoading = isDeletingSeriesDownload,
                                                    autoDismiss = false,
                                                    onClick = {
                                                        if (isDeletingSeriesDownload) return@Custom
                                                        isDeletingSeriesDownload = true
                                                        coroutineScope.launch {
                                                            snackbarHostState.showSnackbar("Deleting series downloads...")
                                                        }
                                                        onDeleteSeriesDownload { success ->
                                                            isDeletingSeriesDownload = false
                                                            showSeriesActionMenu = false
                                                            coroutineScope.launch {
                                                                snackbarHostState.showSnackbar(
                                                                    if (success) "Deleted successfully!"
                                                                    else "Failed to delete series."
                                                                )
                                                            }
                                                        }
                                                    }
                                                ))
                                            }
                                        }
                                    )
                                }
                            }
                        }
                    }
                }

                // Seasons selector
                if (series.seasons.isNotEmpty()) {
                    item {
                        Text(
                            text = "Episodes",
                            color = NetflixWhite,
                            fontSize = 18.sp,
                            fontWeight = FontWeight.SemiBold,
                            modifier = Modifier.padding(horizontal = 32.dp),
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        LazyRow(
                            contentPadding = PaddingValues(horizontal = 32.dp),
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                        ) {
                            items(series.seasons) { season ->
                                SeasonChip(
                                    season = season,
                                    isSelected = uiState.selectedSeason?.id == season.id,
                                    onClick = { onSeasonSelect(season) },
                                )
                            }
                        }
                        Spacer(modifier = Modifier.height(16.dp))
                    }
                }

                // Episodes
                if (uiState.areEpisodesLoading) {
                    item {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(200.dp),
                            contentAlignment = Alignment.Center,
                        ) {
                            CircularProgressIndicator(color = NetflixRed)
                        }
                    }
                } else {
                    items(uiState.episodes) { episode ->
                        EpisodeCard(
                            episode = episode,
                            progress = uiState.episodeProgress[episode.mediaId],
                            seriesTitle = series.title,
                            onClick = { onEpisodeClick(episode.mediaId) },
                            onEditClick = if (uiState.isAdmin) { { editingEpisode = episode } } else null,
                            onAutoDownloadSubtitle = { mediaId, onResult ->
                                coroutineScope.launch {
                                    snackbarHostState.showSnackbar("Searching and downloading subtitle...")
                                }
                                onAutoDownloadSubtitle(mediaId) { success ->
                                    onResult(success)
                                    coroutineScope.launch {
                                        snackbarHostState.showSnackbar(if (success) "Subtitle downloaded successfully!" else "Could not find a subtitle match.")
                                    }
                                }
                            },
                            onDownloadEpisode = { mediaId, onResult ->
                                coroutineScope.launch {
                                    snackbarHostState.showSnackbar("Starting download to NAS...")
                                }
                                onDownloadEpisode(mediaId) { success ->
                                    onResult(success)
                                    coroutineScope.launch {
                                        snackbarHostState.showSnackbar(
                                            if (success) "Download started!"
                                            else "Failed to start download."
                                        )
                                    }
                                }
                            },
                            onDeleteEpisodeDownload = { mediaId, onResult ->
                                coroutineScope.launch {
                                    snackbarHostState.showSnackbar("Deleting local download...")
                                }
                                onDeleteEpisodeDownload(mediaId) { success ->
                                    onResult(success)
                                    coroutineScope.launch {
                                        snackbarHostState.showSnackbar(
                                            if (success) "Deleted successfully!"
                                            else "Failed to delete download."
                                        )
                                    }
                                }
                            },
                            isAdmin = uiState.isAdmin
                        )
                    }
                } // closes else
            } // closes LazyColumn
        } // closes PullToRefreshBox
    } // closes else (isLoading/error/etc)
    } // closes Scaffold padding block
    } // closes Box
} // closes fun SeriesDetailContent

@Composable
fun SeasonChip(
    season: Season,
    isSelected: Boolean,
    onClick: () -> Unit,
) {
    Surface(
        onClick = onClick,
        color = if (isSelected) NetflixRed else NetflixDark,
        shape = RoundedCornerShape(8.dp),
    ) {
        Text(
            text = if (!season.title.isNullOrBlank()) season.title else "Season ${season.seasonNumber}",
            color = NetflixWhite,
            fontSize = 14.sp,
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
        )
    }
}

@Composable
fun ReadMoreOverview(overview: String) {
    var expanded by remember { mutableStateOf(false) }
    val minLines = 3

    Column(modifier = Modifier.clickable { expanded = !expanded }) {
        Text(
            text = overview,
            color = NetflixLightGray,
            fontSize = 14.sp,
            lineHeight = 20.sp,
            maxLines = if (expanded) Int.MAX_VALUE else minLines,
            overflow = TextOverflow.Ellipsis,
        )
        if (!expanded && overview.length > 120) {
            Text(
                text = "Read more",
                color = NetflixRed,
                fontSize = 14.sp,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.padding(top = 4.dp),
            )
        }
    }
}

@Composable
fun EpisodeCard(
    episode: Episode,
    progress: WatchProgress?,
    seriesTitle: String,
    onClick: () -> Unit,
    onEditClick: (() -> Unit)? = null,
    onAutoDownloadSubtitle: (Int, (Boolean) -> Unit) -> Unit = { _, _ -> },
    onDownloadEpisode: (Int, (Boolean) -> Unit) -> Unit = { _, _ -> },
    onDeleteEpisodeDownload: (Int, (Boolean) -> Unit) -> Unit = { _, _ -> },
    isAdmin: Boolean = false,
) {
    var showActionMenu by remember { mutableStateOf(false) }
    var isDownloadingSubtitle by remember { mutableStateOf(false) }
    var isDownloadingEpisode by remember { mutableStateOf(false) }
    var isDeletingEpisodeDownload by remember { mutableStateOf(false) }

    val isCloudMedia = episode.filePath?.contains("://") == true &&
        episode.filePath.startsWith("http").not()
    val isDownloaded = episode.filePath?.contains("library/downloads") == true &&
        episode.filePath.contains("://").not()

    val screenWidth = LocalConfiguration.current.screenWidthDp
    val cardPadding = if (screenWidth < 600) 16.dp else 32.dp

    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = cardPadding, vertical = 8.dp)
            .clickable(onClick = onClick),
        color = Color(0xFF1C1C1C), // Slightly brighter than Pitch Black
        shape = RoundedCornerShape(8.dp),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            // Thumbnail
            Surface(
                modifier = Modifier
                    .width(128.dp)
                    .height(80.dp),
                color = NetflixGray,
                shape = RoundedCornerShape(4.dp),
            ) {
                if (episode.still != null || episode.stillPath != null) {
                    if (episode.still != null) {
                        ResponsiveImage(
                            data = episode.still,
                            contentDescription = episode.title,
                            modifier = Modifier.fillMaxSize(),
                            contentScale = ContentScale.Crop,
                        )
                    } else {
                        AsyncImage(
                            model = episode.stillPath,
                            contentDescription = episode.title,
                            modifier = Modifier.fillMaxSize(),
                            contentScale = ContentScale.Crop,
                        )
                    }
                } else {
                    Box(contentAlignment = Alignment.Center) {
                        Column(horizontalAlignment = Alignment.CenterHorizontally) {
                            Text(
                                text = "S${episode.seasonNumber ?: 1}E${episode.episodeNumber}",
                                color = NetflixRed,
                                fontWeight = FontWeight.Bold,
                                fontSize = 12.sp,
                            )
                        }
                    }
                }
            }
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        text = episode.episodeNumber.toString(),
                        color = Color.Gray,
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Bold,
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = episode.title,
                        color = NetflixWhite,
                        fontSize = 16.sp,
                        fontWeight = FontWeight.SemiBold,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                    if (progress?.completed == true) {
                        Spacer(modifier = Modifier.width(8.dp))
                        Icon(
                            imageVector = LucideIcons.Check,
                            contentDescription = null,
                            tint = Color(0xFF22C55E), // green-500
                            modifier = Modifier.size(16.dp)
                        )
                    }
                }
                Spacer(modifier = Modifier.height(4.dp))
                episode.overview?.let { overview ->
                    Text(
                        text = overview,
                        color = NetflixLightGray,
                        fontSize = 14.sp,
                        maxLines = 2,
                        overflow = TextOverflow.Ellipsis,
                    )
                }

                Spacer(modifier = Modifier.height(4.dp))
                val sourceName = when {
                    episode.filePath?.startsWith("ophim://", ignoreCase = true) == true -> "OPhim"
                    episode.filePath?.startsWith("fshare://", ignoreCase = true) == true -> "Fshare"
                    else -> "Local"
                }
                Text(
                    text = "Source: $sourceName",
                    color = Color.Gray,
                    fontSize = 12.sp,
                )
                if (progress != null && progress.position > 0 && !progress.completed) {
                    val duration = episode.duration ?: 0f
                    if (duration > 0) {
                        val progressPercent = minOf(100f, (progress.position / duration) * 100f)
                        Spacer(modifier = Modifier.height(8.dp))
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(4.dp)
                                .background(Color(0xFF374151), RoundedCornerShape(2.dp)) // gray-700
                        ) {
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth(progressPercent / 100f)
                                    .height(4.dp)
                                    .background(NetflixRed, RoundedCornerShape(2.dp))
                            )
                        }
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = "${progressPercent.toInt()}% đã xem",
                            color = Color.Gray,
                            fontSize = 12.sp
                        )
                    }
                }
            }
            // Action menu
            Box {
                val coroutineScope = rememberCoroutineScope()
                ActionMenuButton(
                    expanded = showActionMenu,
                    onClick = { showActionMenu = true },
                )
                ActionMenu(
                    expanded = showActionMenu,
                    onDismiss = { showActionMenu = false },
                    items = buildList {
                        add(ActionMenuItem.MarkWatched(
                            isWatched = progress?.completed == true,
                            onClick = { showActionMenu = false },
                        ))
                        add(ActionMenuItem.StreamUrl(onClick = { showActionMenu = false }))
                        add(ActionMenuItem.Custom(
                            label = "Auto-download subtitles",
                            icon = LucideIcons.Download,
                            isLoading = isDownloadingSubtitle,
                            autoDismiss = false,
                            onClick = {
                                if (isDownloadingSubtitle) return@Custom
                                isDownloadingSubtitle = true
                                onAutoDownloadSubtitle(episode.mediaId) {
                                    isDownloadingSubtitle = false
                                    showActionMenu = false
                                }
                            }
                        ))
                        if (isCloudMedia && isAdmin) {
                            add(ActionMenuItem.Custom(
                                label = "Download to NAS",
                                icon = LucideIcons.Download,
                                isLoading = isDownloadingEpisode,
                                autoDismiss = false,
                                onClick = {
                                    if (isDownloadingEpisode) return@Custom
                                    isDownloadingEpisode = true
                                    onDownloadEpisode(episode.mediaId) {
                                        isDownloadingEpisode = false
                                        showActionMenu = false
                                    }
                                }
                            ))
                        }
                        if (isDownloaded && isAdmin) {
                            add(ActionMenuItem.Custom(
                                label = "Delete Local Download",
                                icon = LucideIcons.Delete,
                                isLoading = isDeletingEpisodeDownload,
                                autoDismiss = false,
                                onClick = {
                                    if (isDeletingEpisodeDownload) return@Custom
                                    isDeletingEpisodeDownload = true
                                    onDeleteEpisodeDownload(episode.mediaId) {
                                        isDeletingEpisodeDownload = false
                                        showActionMenu = false
                                    }
                                }
                            ))
                        }
                        if (onEditClick != null) {
                            add(ActionMenuItem.Separator)
                            add(ActionMenuItem.Edit(onClick = { showActionMenu = false; onEditClick() }))
                        }
                    },
                )
            }
        }
    }
}

@Composable
fun EpisodeEditDialog(
    episode: Episode,
    onClose: () -> Unit,
    onSave: (episodeId: Int, title: String, overview: String?, episodeNumber: Int?, airDate: String?) -> Unit,
    isSaving: Boolean = false,
) {
    var title by remember { mutableStateOf(episode.title) }
    var overview by remember { mutableStateOf(episode.overview ?: "") }
    var episodeNumber by remember { mutableStateOf(episode.episodeNumber.toString()) }
    var airDate by remember { mutableStateOf(episode.airDate ?: "") }

    Dialog(onDismissRequest = onClose) {
        Surface(
            shape = RoundedCornerShape(16.dp),
            color = NetflixDark,
        ) {
            Column(
                modifier = Modifier
                    .padding(24.dp)
                    .fillMaxWidth(),
            ) {
                // Header
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        text = "Edit Episode ${episode.episodeNumber}",
                        color = NetflixWhite,
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Bold,
                    )
                    IconButton(onClick = onClose) {
                        Icon(
                            LucideIcons.Close,
                            contentDescription = "Close",
                            tint = NetflixLightGray,
                        )
                    }
                }
                Spacer(modifier = Modifier.height(16.dp))

                // Title field
                OutlinedTextField(
                    value = title,
                    onValueChange = { title = it },
                    label = { Text(stringResource(R.string.detail_title)) },
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = NetflixRed,
                        unfocusedBorderColor = NetflixGray,
                        focusedLabelColor = NetflixRed,
                        unfocusedLabelColor = NetflixLightGray,
                        cursorColor = NetflixRed,
                        focusedTextColor = NetflixWhite,
                        unfocusedTextColor = NetflixWhite,
                    ),
                    modifier = Modifier.fillMaxWidth(),
                )
                Spacer(modifier = Modifier.height(12.dp))

                // Episode Number field
                OutlinedTextField(
                    value = episodeNumber,
                    onValueChange = { value ->
                        // Only allow digits
                        episodeNumber = value.filter { it.isDigit() }
                    },
                    label = { Text(stringResource(R.string.detail_episode_number)) },
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = NetflixRed,
                        unfocusedBorderColor = NetflixGray,
                        focusedLabelColor = NetflixRed,
                        unfocusedLabelColor = NetflixLightGray,
                        cursorColor = NetflixRed,
                        focusedTextColor = NetflixWhite,
                        unfocusedTextColor = NetflixWhite,
                    ),
                    modifier = Modifier.fillMaxWidth(),
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                )
                Spacer(modifier = Modifier.height(12.dp))

                // Air Date field
                OutlinedTextField(
                    value = airDate,
                    onValueChange = { airDate = it },
                    label = { Text(stringResource(R.string.detail_air_date)) },
                    placeholder = { Text(stringResource(R.string.detail_date_placeholder), color = NetflixGray) },
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = NetflixRed,
                        unfocusedBorderColor = NetflixGray,
                        focusedLabelColor = NetflixRed,
                        unfocusedLabelColor = NetflixLightGray,
                        cursorColor = NetflixRed,
                        focusedTextColor = NetflixWhite,
                        unfocusedTextColor = NetflixWhite,
                    ),
                    modifier = Modifier.fillMaxWidth(),
                )
                Spacer(modifier = Modifier.height(12.dp))

                // Overview field
                OutlinedTextField(
                    value = overview,
                    onValueChange = { overview = it },
                    label = { Text(stringResource(R.string.detail_overview)) },
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = NetflixRed,
                        unfocusedBorderColor = NetflixGray,
                        focusedLabelColor = NetflixRed,
                        unfocusedLabelColor = NetflixLightGray,
                        cursorColor = NetflixRed,
                        focusedTextColor = NetflixWhite,
                        unfocusedTextColor = NetflixWhite,
                    ),
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(120.dp),
                    maxLines = 5,
                )
                Spacer(modifier = Modifier.height(24.dp))

                // Actions
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    OutlinedButton(
                        onClick = onClose,
                        modifier = Modifier.weight(1f),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = NetflixWhite),
                    ) {
                        Text(stringResource(R.string.action_cancel))
                    }
                    Button(
                        onClick = {
                            val epNum = episodeNumber.toIntOrNull()
                            onSave(
                                episode.id,
                                title,
                                overview.ifBlank { null },
                                if (epNum != null && epNum != episode.episodeNumber) epNum else null,
                                airDate.ifBlank { null }
                            )
                        },
                        modifier = Modifier.weight(2f),
                        enabled = title.isNotBlank() && !isSaving,
                        colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                    ) {
                        if (isSaving) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(16.dp),
                                color = NetflixWhite,
                                strokeWidth = 2.dp,
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                        }
                        Text(if (isSaving) "Saving..." else "Save")
                    }
                }
            }
        }
    }
}

private fun extractYouTubeKeyFromUrl(url: String): String? {
    // Handle formats: https://www.youtube.com/embed/VIDEO_KEY, https://youtu.be/VIDEO_KEY
    val embedMatch = Regex("embed/([^?]+)").find(url)
    if (embedMatch != null) return embedMatch.groupValues[1]

    val shortMatch = Regex("youtu\\.be/([^?]+)").find(url)
    if (shortMatch != null) return shortMatch.groupValues[1]

    return null
}

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
fun SeriesDetailScreenPreview() {
    val sampleSeason = com.velox.app.domain.model.Season(
        id = 1,
        seriesId = 1,
        seasonNumber = 1,
        title = "Season 1",
        overview = "The first season.",
        posterPath = null,
        episodeCount = 10
    )
    val sampleEpisode = com.velox.app.domain.model.Episode(
        id = 1,
        mediaId = 1,
        seriesId = 1,
        seasonId = 1,
        seasonNumber = 1,
        episodeNumber = 1,
        title = "Pilot",
        overview = "The first episode.",
        stillPath = null,
        airDate = "2023-01-01",
        duration = 3000f
    )
    val sampleSeries = com.velox.app.domain.model.SeriesDetail(
        id = 1,
        title = "Sample Series",
        overview = "This is a sample series overview.",
        posterPath = null,
        backdropPath = null,
        status = "Returning Series",
        network = "Netflix",
        firstAirDate = "2023-01-01",
        genres = emptyList(),
        seasons = listOf(sampleSeason),
        credits = null
    )
    val uiState = SeriesDetailUiState(
        isLoading = false,
        series = sampleSeries,
        selectedSeason = sampleSeason,
        episodes = listOf(sampleEpisode),
        isFavorite = false,
        isAdmin = true
    )

    VeloxTheme {
        SeriesDetailContent(
            uiState = uiState,
            onBackClick = {},
            onEpisodeClick = {},
            onRefresh = {},
            onToggleFavorite = {},
            onSeasonSelect = {},
            onEditEpisode = {},
        )
    }
}

@Preview
@Composable
fun SeasonChipPreview() {
    val sampleSeason = com.velox.app.domain.model.Season(
        id = 1,
        seriesId = 1,
        seasonNumber = 1,
        title = "Season 1",
        overview = null,
        posterPath = null,
        episodeCount = 10
    )
    VeloxTheme {
        SeasonChip(
            season = sampleSeason,
            isSelected = true,
            onClick = {}
        )
    }
}

@Preview
@Composable
fun EpisodeCardPreview() {
    val sampleEpisode = com.velox.app.domain.model.Episode(
        id = 1,
        mediaId = 1,
        seriesId = 1,
        seasonId = 1,
        seasonNumber = 1,
        episodeNumber = 1,
        title = "Pilot",
        overview = "The first episode.",
        stillPath = null,
        airDate = "2023-01-01",
        duration = 3000f
    )
    VeloxTheme {
        EpisodeCard(
            episode = sampleEpisode,
            progress = null,
            seriesTitle = "Sample Series",
            onClick = {}
        )
    }
}
