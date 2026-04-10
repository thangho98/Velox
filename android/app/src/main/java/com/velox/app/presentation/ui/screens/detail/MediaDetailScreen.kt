package com.velox.app.presentation.ui.screens.detail

import java.util.Locale
import com.velox.app.presentation.ui.components.LucideIcons
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.FavoriteBorder
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import coil.compose.AsyncImage
import com.velox.app.data.model.CinemaDto
import com.velox.app.data.model.TrailerDto
import com.velox.app.domain.model.CastMember
import com.velox.app.domain.model.Credits
import com.velox.app.domain.model.Genre
import com.velox.app.domain.model.MediaDetail
import com.velox.app.domain.model.MediaFile
import com.velox.app.domain.model.MediaItem
import com.velox.app.domain.model.WatchProgress
import com.velox.app.presentation.ui.components.ActionMenu
import com.velox.app.presentation.ui.components.ActionMenuButton
import com.velox.app.presentation.ui.components.ActionMenuItem
import com.velox.app.presentation.ui.components.YouTubePlayer
import com.velox.app.presentation.viewmodel.MediaDetailUiState
import com.velox.app.presentation.viewmodel.MediaDetailViewModel
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixGray
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MediaDetailScreen(
    mediaId: Int,
    onBackClick: () -> Unit,
    onPlayClick: () -> Unit,
    viewModel: MediaDetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()
    val clipboardManager = androidx.compose.ui.platform.LocalClipboardManager.current

    MediaDetailContent(
        uiState = uiState,
        onBackClick = onBackClick,
        onPlayClick = onPlayClick,
        onFavoriteClick = { viewModel.toggleFavorite() },
        onWatchedClick = { viewModel.toggleWatched() },
        onRetryClick = { viewModel.refresh() },
        onSubtitleSelect = { viewModel.selectSubtitleLanguage(it) },
        onCopyStreamUrl = {
            viewModel.copyStreamUrl { url ->
                if (url != null) {
                    clipboardManager.setText(androidx.compose.ui.text.AnnotatedString(url))
                }
            }
        },
        onRefreshMetadata = { viewModel.refreshMetadata() },
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun MediaDetailContent(
    uiState: MediaDetailUiState,
    onBackClick: () -> Unit,
    onPlayClick: () -> Unit,
    onFavoriteClick: () -> Unit,
    onWatchedClick: () -> Unit,
    onRetryClick: () -> Unit,
    onSubtitleSelect: (String?) -> Unit = {},
    onCopyStreamUrl: () -> Unit = {},
    onRefreshMetadata: () -> Unit = {},
) {
    var currentTrailerIndex by remember { mutableIntStateOf(0) }
    var showTrailer by remember { mutableStateOf(false) }

    // YouTube trailer overlay
    if (showTrailer) {
        val trailers = uiState.cinema?.trailers ?: emptyList()
        if (trailers.isNotEmpty() && currentTrailerIndex < trailers.size) {
            val trailer = trailers[currentTrailerIndex]
            val videoKey = trailer.key ?: extractYouTubeKey(trailer.url)
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
                        onPlayClick()
                    },
                )
            }
        }
    }

    Box(modifier = Modifier.fillMaxSize().background(NetflixBlack)) {
        // Fixed Background Backdrop (like SeriesDetailScreen)
        val media = uiState.media
        if (media != null && (media.backdropPath != null || media.posterPath != null)) {
            AsyncImage(
                model = media.backdropPath ?: media.posterPath,
                contentDescription = null,
                modifier = Modifier.fillMaxSize(),
                contentScale = ContentScale.Crop,
                alignment = Alignment.TopCenter,
            )
            // Vertical Gradient Overlay
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(
                        Brush.verticalGradient(
                            0.0f to Color(0x4D000000),
                            0.5f to Color(0xCC000000),
                            1.0f to NetflixBlack,
                        ),
                    ),
            )
            // Horizontal Gradient Overlay (cinematic left darkening)
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(
                        Brush.horizontalGradient(
                            0.0f to NetflixBlack,
                            0.4f to Color(0x80000000),
                            1.0f to Color.Transparent,
                        ),
                    ),
            )
        }

        Scaffold(
            topBar = {
                TopAppBar(
                    title = { },
                    navigationIcon = {
                        Surface(
                            onClick = onBackClick,
                            shape = androidx.compose.foundation.shape.CircleShape,
                            color = Color(0x80000000),
                            modifier = Modifier.size(40.dp).padding(4.dp),
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
                            onClick = onRetryClick,
                            colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                        ) {
                            Text("Retry")
                        }
                    }
                }
            } else {
                val mediaDetail = uiState.media ?: return@Scaffold
                val primaryFile = uiState.primaryFile

                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .verticalScroll(rememberScrollState()),
                ) {
                    Spacer(modifier = Modifier.height(padding.calculateTopPadding() + 40.dp))

                    // Centered Poster (like SeriesDetailScreen)
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(bottom = 24.dp),
                        contentAlignment = Alignment.Center,
                    ) {
                        if (mediaDetail.posterPath != null) {
                            Surface(
                                modifier = Modifier
                                    .padding(top = 20.dp)
                                    .width(260.dp)
                                    .aspectRatio(2f / 3f),
                                shape = RoundedCornerShape(8.dp),
                                color = NetflixDark,
                                shadowElevation = 12.dp,
                            ) {
                                AsyncImage(
                                    model = mediaDetail.posterPath,
                                    contentDescription = mediaDetail.title,
                                    modifier = Modifier.fillMaxSize(),
                                    contentScale = ContentScale.Crop,
                                )
                            }
                        } else {
                            Surface(
                                modifier = Modifier
                                    .padding(top = 20.dp)
                                    .width(260.dp)
                                    .aspectRatio(2f / 3f),
                                shape = RoundedCornerShape(8.dp),
                                color = NetflixDark,
                                shadowElevation = 12.dp,
                            ) {
                                Box(contentAlignment = Alignment.Center) {
                                    Text(
                                        text = mediaDetail.title.take(2).uppercase(),
                                        color = NetflixLightGray,
                                        fontSize = 48.sp,
                                        fontWeight = FontWeight.Bold,
                                    )
                                }
                            }
                        }
                    }

                    // Info section
                    Column(modifier = Modifier.padding(horizontal = 32.dp)) {
                        MediaDetailInfo(
                            media = mediaDetail,
                            uiState = uiState,
                            primaryFile = primaryFile,
                            onPlayClick = onPlayClick,
                            onFavoriteClick = onFavoriteClick,
                            onWatchedClick = onWatchedClick,
                            showTrailer = showTrailer,
                            onShowTrailer = { showTrailer = true },
                            onSubtitleSelect = onSubtitleSelect,
                            onCopyStreamUrl = onCopyStreamUrl,
                            onRefreshMetadata = onRefreshMetadata,
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun MediaDetailInfo(
    media: MediaDetail,
    uiState: MediaDetailUiState,
    primaryFile: MediaFile?,
    onPlayClick: () -> Unit,
    onFavoriteClick: () -> Unit,
    onWatchedClick: () -> Unit,
    showTrailer: Boolean,
    onShowTrailer: () -> Unit,
    onSubtitleSelect: (String?) -> Unit = {},
    onCopyStreamUrl: () -> Unit = {},
    onRefreshMetadata: () -> Unit = {},
) {
    // Title
    Text(
        text = media.title,
        color = NetflixWhite,
        fontSize = 28.sp,
        fontWeight = FontWeight.Bold,
    )

    Spacer(modifier = Modifier.height(8.dp))

    // Meta info row: Year · Duration · Ends at · Rating
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        media.releaseDate?.take(4)?.let { year ->
            Text(text = year, color = NetflixLightGray, fontSize = 14.sp)
        }
        if (media.duration != null) {
            val hours = (media.duration / 60).toInt()
            val mins = (media.duration % 60).toInt()
            val durationText = if (hours > 0) "${hours}h ${mins}m" else "${mins}m"
            Text(text = durationText, color = NetflixLightGray, fontSize = 14.sp)
        }
        // Ends at
        val progress = uiState.progress
        if (progress != null && !progress.completed && progress.position > 0) {
            val durationSeconds = (primaryFile?.duration ?: media.duration ?: 0f).toInt()
            if (durationSeconds > 0) {
                val remainingSeconds = durationSeconds - progress.position.toInt()
                if (remainingSeconds > 0) {
                    val cal = java.util.Calendar.getInstance()
                    cal.add(java.util.Calendar.SECOND, remainingSeconds)
                    val endsAtText = String.format(
                        Locale.getDefault(),
                        "Ends at %d:%02d %s",
                        if (cal.get(java.util.Calendar.HOUR) == 0) 12 else cal.get(java.util.Calendar.HOUR),
                        cal.get(java.util.Calendar.MINUTE),
                        if (cal.get(java.util.Calendar.AM_PM) == 0) "AM" else "PM",
                    )
                    Text(text = endsAtText, color = NetflixLightGray, fontSize = 14.sp)
                }
            }
        }
        // Rating
        if (media.rating != null) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    LucideIcons.Star,
                    contentDescription = null,
                    tint = Color(0xFFEAB308),
                    modifier = Modifier.size(14.dp),
                )
                Spacer(modifier = Modifier.width(2.dp))
                Text(
                    text = String.format(Locale.getDefault(), "%.1f", media.rating),
                    color = NetflixLightGray,
                    fontSize = 14.sp,
                )
            }
        }
    }

    Spacer(modifier = Modifier.height(8.dp))

    // Tech specs row with labels (matching webapp: Video 2076p HEVC · Audio EAC3 · Container MATROSKA · Size 34.62 GB)
    if (primaryFile != null) {
        Row(
            horizontalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            if (primaryFile.height != null && primaryFile.height > 0) {
                val codec = primaryFile.videoCodec?.uppercase() ?: ""
                TechSpecLabeled(label = "Video", value = "${primaryFile.height}p $codec".trim())
            }
            if (primaryFile.audioCodec != null) {
                TechSpecLabeled(label = "Audio", value = primaryFile.audioCodec.uppercase())
            }
            if (primaryFile.container != null) {
                TechSpecLabeled(label = "Container", value = primaryFile.container.uppercase())
            }
            if (primaryFile.fileSize > 0) {
                val sizeGB = primaryFile.fileSize / (1024.0 * 1024.0 * 1024.0)
                val sizeText = if (sizeGB >= 1) {
                    String.format(Locale.getDefault(), "%.2f GB", sizeGB)
                } else {
                    val sizeMB = primaryFile.fileSize / (1024.0 * 1024.0)
                    String.format(Locale.getDefault(), "%.0f MB", sizeMB)
                }
                TechSpecLabeled(label = "Size", value = sizeText)
            }
        }
        Spacer(modifier = Modifier.height(16.dp))
    }

    // Overview (before buttons, like webapp)
    if (media.overview != null) {
        Text(
            text = media.overview,
            color = NetflixLightGray,
            fontSize = 16.sp,
            lineHeight = 24.sp,
        )
        Spacer(modifier = Modifier.height(24.dp))
    }

    // Actions row: Resume, From Beginning, Watched, Favorite, Trailer, Menu
    val progress = uiState.progress
    val canResume = progress != null && progress.position > 0 && !progress.completed
    Row(
        horizontalArrangement = Arrangement.spacedBy(10.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        // Play/Resume button
        Button(
            onClick = onPlayClick,
            colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
            shape = RoundedCornerShape(4.dp),
            contentPadding = PaddingValues(horizontal = 20.dp, vertical = 10.dp),
        ) {
            Icon(
                LucideIcons.PlayArrow,
                contentDescription = null,
                modifier = Modifier.size(18.dp),
            )
            Spacer(modifier = Modifier.width(6.dp))
            Text(
                text = if (canResume) "Resume" else "Play",
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
            )
        }

        // From Beginning button (when resuming)
        if (canResume) {
            OutlinedButton(
                onClick = onPlayClick, // TODO: play from beginning
                shape = RoundedCornerShape(4.dp),
                contentPadding = PaddingValues(horizontal = 16.dp, vertical = 10.dp),
                colors = ButtonDefaults.outlinedButtonColors(contentColor = NetflixWhite),
            ) {
                Icon(
                    LucideIcons.PlayArrow,
                    contentDescription = null,
                    modifier = Modifier.size(16.dp),
                )
                Spacer(modifier = Modifier.width(6.dp))
                Text(
                    text = "From Beginning",
                    fontSize = 13.sp,
                    fontWeight = FontWeight.Medium,
                )
            }
        }

        // Watched/Check button
        OutlinedButton(
            onClick = onWatchedClick,
            contentPadding = PaddingValues(8.dp),
            colors = ButtonDefaults.outlinedButtonColors(
                contentColor = if (uiState.isWatched) NetflixWhite else NetflixGray,
            ),
        ) {
            Icon(
                LucideIcons.Check,
                contentDescription = "Watched",
                modifier = Modifier.size(20.dp),
            )
        }

        // Favorite button
        OutlinedButton(
            onClick = onFavoriteClick,
            contentPadding = PaddingValues(8.dp),
            colors = ButtonDefaults.outlinedButtonColors(
                contentColor = if (uiState.isFavorite) NetflixRed else NetflixWhite,
            ),
        ) {
            Icon(
                if (uiState.isFavorite) LucideIcons.Favorite else LucideIcons.FavoriteBorder,
                contentDescription = "Favorite",
                modifier = Modifier.size(20.dp),
            )
        }

        // Action menu
        Box {
            var showActionMenu by remember { mutableStateOf(false) }
            ActionMenuButton(
                expanded = showActionMenu,
                onClick = { showActionMenu = true },
            )
            ActionMenu(
                expanded = showActionMenu,
                onDismiss = { showActionMenu = false },
                items = buildList {
                    add(ActionMenuItem.StreamUrl(onClick = { showActionMenu = false; onCopyStreamUrl() }))
                    if (uiState.isAdmin) {
                        add(ActionMenuItem.Edit(onClick = { showActionMenu = false }))
                        add(ActionMenuItem.RefreshMetadata(onClick = { showActionMenu = false; onRefreshMetadata() }))
                    }
                },
            )
        }
    }

    // Subtitle selector
    if (uiState.subtitleTracks.isNotEmpty()) {
        Spacer(modifier = Modifier.height(16.dp))
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text(
                text = "Subtitles",
                color = Color(0xFF9CA3AF), // gray-400
                fontSize = 14.sp,
            )
            var subtitleExpanded by remember { mutableStateOf(false) }
            Box {
                Surface(
                    onClick = { subtitleExpanded = true },
                    color = Color(0xFF2A2A2A),
                    shape = RoundedCornerShape(20.dp),
                ) {
                    Text(
                        text = if (uiState.selectedSubtitleLanguage != null) {
                            uiState.subtitleTracks.find { it.language == uiState.selectedSubtitleLanguage }?.label
                                ?: uiState.selectedSubtitleLanguage
                        } else "Off",
                        color = NetflixWhite,
                        fontSize = 14.sp,
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                    )
                }
                DropdownMenu(
                    expanded = subtitleExpanded,
                    onDismissRequest = { subtitleExpanded = false },
                ) {
                    DropdownMenuItem(
                        text = { Text("Off", color = NetflixWhite) },
                        onClick = {
                            onSubtitleSelect(null)
                            subtitleExpanded = false
                        },
                    )
                    uiState.subtitleTracks.forEach { track ->
                        DropdownMenuItem(
                            text = {
                                Text(
                                    text = track.label.ifBlank { "${track.language} (${track.format?.uppercase() ?: ""})" },
                                    color = NetflixWhite,
                                )
                            },
                            onClick = {
                                onSubtitleSelect(track.language)
                                subtitleExpanded = false
                            },
                        )
                    }
                }
            }
        }
    }

    Spacer(modifier = Modifier.height(16.dp))

    // Progress bar (inline with remaining text, matching webapp)
    if (progress != null && progress.position > 0) {
        val totalDuration = progress.duration
            ?: primaryFile?.duration
            ?: media.duration
            ?: 0f
        val progressPercent = if (totalDuration > 0) {
            (progress.position / totalDuration).coerceIn(0f, 1f)
        } else 0f
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            modifier = Modifier.widthIn(max = 520.dp),
        ) {
            LinearProgressIndicator(
                progress = { progressPercent },
                modifier = Modifier
                    .weight(1f)
                    .height(4.dp),
                color = Color(0xFF22C55E),
                trackColor = Color(0xFF374151),
                drawStopIndicator = {}
            )
            Text(
                text = if (progress.completed) {
                    "Watched"
                } else if (totalDuration > 0) {
                    val remainSec = (totalDuration - progress.position).coerceAtLeast(0f).toInt()
                    val remainHr = remainSec / 3600
                    val remainM = (remainSec % 3600) / 60
                    if (remainHr > 0) "${remainHr}h ${remainM}m remaining" else "${remainM}m remaining"
                } else "",
                color = NetflixLightGray,
                fontSize = 14.sp,
            )
        }
        Spacer(modifier = Modifier.height(16.dp))
    }

    // Genres
    if (media.genres.isNotEmpty()) {
        Row(
            modifier = Modifier.horizontalScroll(rememberScrollState()),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            media.genres.forEach { genre ->
                Surface(
                    color = NetflixDark,
                    shape = RoundedCornerShape(16.dp),
                ) {
                    Text(
                        text = genre.name,
                        color = NetflixLightGray,
                        fontSize = 12.sp,
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
                    )
                }
            }
        }
        Spacer(modifier = Modifier.height(16.dp))
    }

    // Cast
    if (media.credits?.cast?.isNotEmpty() == true) {
        Text(
            text = "Cast",
            color = NetflixWhite,
            fontSize = 18.sp,
            fontWeight = FontWeight.SemiBold,
        )
        Spacer(modifier = Modifier.height(12.dp))
        LazyRow(
            horizontalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            items(media.credits.cast.take(10)) { castMember ->
                CastCard(castMember = castMember)
            }
        }
        Spacer(modifier = Modifier.height(24.dp))
    }

    // Similar
    if (media.similar.isNotEmpty()) {
        Text(
            text = "More Like This",
            color = NetflixWhite,
            fontSize = 18.sp,
            fontWeight = FontWeight.SemiBold,
        )
        Spacer(modifier = Modifier.height(12.dp))
        LazyRow(
            horizontalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            items(media.similar) { item ->
                SimilarCard(
                    item = item,
                    onClick = { /* Navigate to similar */ },
                )
            }
        }
    }
}

@Composable
private fun TechSpecLabeled(label: String, value: String) {
    Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
        Text(
            text = label,
            color = Color(0xFF9CA3AF), // gray-400
            fontSize = 13.sp,
        )
        Text(
            text = value,
            color = NetflixWhite,
            fontSize = 13.sp,
        )
    }
}

@Composable
fun CastCard(castMember: CastMember) {
    Column(
        modifier = Modifier.width(80.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Surface(
            modifier = Modifier
                .size(72.dp)
                .clip(CircleShape),
            color = NetflixGray,
        ) {
            if (castMember.profilePath != null) {
                AsyncImage(
                    model = castMember.profilePath,
                    contentDescription = castMember.name,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Crop,
                )
            } else {
                Box(contentAlignment = Alignment.Center) {
                    Text(
                        text = castMember.name.take(2).uppercase(),
                        color = NetflixLightGray,
                        fontSize = 18.sp,
                    )
                }
            }
        }
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = castMember.name,
            color = NetflixWhite,
            fontSize = 12.sp,
            maxLines = 2,
        )
        castMember.character?.let {
            Text(
                text = it,
                color = NetflixLightGray,
                fontSize = 10.sp,
                maxLines = 1,
            )
        }
    }
}

@Composable
fun SimilarCard(
    item: MediaItem,
    onClick: () -> Unit,
) {
    Column(
        modifier = Modifier
            .width(120.dp)
            .clickable(onClick = onClick),
    ) {
        Surface(
            modifier = Modifier
                .width(120.dp)
                .height(160.dp),
            color = NetflixGray,
            shape = RoundedCornerShape(8.dp),
        ) {
            if (item.posterPath != null) {
                AsyncImage(
                    model = item.posterPath,
                    contentDescription = item.title,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Crop,
                )
            } else {
                Box(contentAlignment = Alignment.Center) {
                    Text(
                        text = item.title.take(2).uppercase(),
                        color = NetflixLightGray,
                        fontSize = 24.sp,
                    )
                }
            }
        }
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = item.title,
            color = NetflixWhite,
            fontSize = 14.sp,
            maxLines = 2,
        )
    }
}

private fun extractYouTubeKey(url: String): String? {
    // Handle formats: https://www.youtube.com/embed/VIDEO_KEY, https://youtu.be/VIDEO_KEY
    val embedMatch = Regex("embed/([^?]+)").find(url)
    if (embedMatch != null) return embedMatch.groupValues[1]

    val shortMatch = Regex("youtu\\.be/([^?]+)").find(url)
    if (shortMatch != null) return shortMatch.groupValues[1]

    return null
}

@Preview(showBackground = true)
@Composable
private fun MediaDetailScreenPreview() {
    VeloxTheme {
        MediaDetailContent(
            uiState = SampleMediaDetailUiState,
            onBackClick = {},
            onPlayClick = {},
            onFavoriteClick = {},
            onWatchedClick = {},
            onRetryClick = {},
        )
    }
}

private val SampleMediaDetail = MediaDetail(
    id = 1,
    title = "The Dark Knight",
    overview = "When the menace known as the Joker wreaks havoc and chaos on the people of Gotham, Batman must accept one of the greatest psychological and physical tests of his ability to fight injustice.",
    posterPath = "https://image.tmdb.org/t/p/w500/qJ2tW6ixuS1nuO3SJznSrSyTGr6.jpg",
    backdropPath = "https://image.tmdb.org/t/p/original/nMK9tii768SjHSuO6YnUvOq677A.jpg",
    rating = 9.0f,
    duration = 152f,
    releaseDate = "2008-07-18",
    genres = listOf(Genre(1, "Action"), Genre(2, "Crime"), Genre(3, "Drama")),
    credits = Credits(
        cast = listOf(
            CastMember(1, "Christian Bale", "Bruce Wayne / Batman", null, 0),
            CastMember(2, "Heath Ledger", "Joker", null, 1),
            CastMember(3, "Aaron Eckhart", "Harvey Dent", null, 2),
        ),
        crew = emptyList(),
    ),
    similar = listOf(
        MediaItem(2, "Batman Begins", null, null, 2005, 8.2f, "movie", null),
        MediaItem(3, "The Dark Knight Rises", null, null, 2012, 8.4f, "movie", null),
    ),
)

private val SampleMediaDetailUiState = MediaDetailUiState(
    isLoading = false,
    media = SampleMediaDetail,
    progress = WatchProgress(1, 3600f, false, true, 9120f),
    isFavorite = true,
    cinema = CinemaDto(
        trailers = listOf(
            TrailerDto(1, "Official Trailer", "https://www.youtube.com/watch?v=EXeTwQW6GQg", "youtube", "EXeTwQW6GQg"),
        ),
    ),
)
