package com.velox.app.presentation.ui.components.player

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil.compose.AsyncImage
import com.velox.app.R
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.viewmodel.DetailPanelUi
import com.velox.app.presentation.viewmodel.PlayerUiState
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite

@Composable
internal fun WatchDetailPanel(
    uiState: PlayerUiState,
    onToggleDetailPanel: (DetailPanelUi) -> Unit,
    onSelectSeason: (Int) -> Unit,
    onSelectEpisode: (Int, Boolean) -> Unit,
    modifier: Modifier = Modifier,
) {
    val mediaDetail = uiState.mediaDetail
    val mediaContext = uiState.mediaContext
    val playbackInfo = uiState.playbackInfo
    val isEpisode = mediaContext?.mediaType == "episode"
    val selectedAudio = playbackInfo?.audioTracks?.find { it.selected }
        ?: playbackInfo?.audioTracks?.find { it.isDefault }
        ?: playbackInfo?.audioTracks?.firstOrNull()

    Column(modifier = modifier) {
        Row(
            horizontalArrangement = Arrangement.spacedBy(20.dp),
            verticalAlignment = Alignment.CenterVertically,
            modifier = Modifier.padding(start = 8.dp, bottom = 16.dp),
        ) {
            DetailTabButton(
                icon = LucideIcons.Info,
                label = "Info",
                isActive = uiState.activeDetailPanel == DetailPanelUi.Info,
                onClick = { onToggleDetailPanel(DetailPanelUi.Info) },
            )
            if (isEpisode && uiState.seasons.isNotEmpty()) {
                DetailTabButton(
                    icon = LucideIcons.ListIcon,
                    label = "Season",
                    isActive = uiState.activeDetailPanel == DetailPanelUi.Season,
                    onClick = { onToggleDetailPanel(DetailPanelUi.Season) },
                )
            }
        }

        when (uiState.activeDetailPanel) {
            DetailPanelUi.Info -> {
                Surface(
                    shape = RoundedCornerShape(28.dp),
                    color = Color.White.copy(alpha = 0.08f),
                    tonalElevation = 0.dp,
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(20.dp),
                        horizontalArrangement = Arrangement.spacedBy(18.dp),
                    ) {
                        Box(
                            modifier = Modifier
                                .width(180.dp)
                                .height(102.dp)
                                .clip(RoundedCornerShape(18.dp))
                                .background(Color.Black.copy(alpha = 0.4f)),
                        ) {
                            val artwork = mediaDetail?.backdrop ?: mediaDetail?.poster
                            val imageUrl = mediaDetail?.backdropPath ?: mediaDetail?.posterPath
                            if (artwork != null || imageUrl != null) {
                                if (artwork != null) {
                                    com.velox.app.presentation.ui.components.ResponsiveImage(
                                        data = artwork,
                                        contentDescription = mediaDetail?.title,
                                        modifier = Modifier.fillMaxSize(),
                                        contentScale = androidx.compose.ui.layout.ContentScale.Crop,
                                    )
                                } else {
                                    coil.compose.AsyncImage(
                                        model = imageUrl,
                                        contentDescription = mediaDetail?.title,
                                        modifier = Modifier.fillMaxSize(),
                                        contentScale = androidx.compose.ui.layout.ContentScale.Crop,
                                    )
                                }
                            } else {
                                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                                    Icon(
                                        LucideIcons.PlayArrow,
                                        contentDescription = null,
                                        tint = Color.White.copy(alpha = 0.2f),
                                        modifier = Modifier.size(28.dp),
                                    )
                                }
                            }
                        }

                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                text = mediaDetail?.title ?: uiState.mediaTitle ?: "Loading...",
                                color = NetflixWhite,
                                fontSize = 24.sp,
                                fontWeight = FontWeight.Bold,
                                maxLines = 2,
                                overflow = TextOverflow.Ellipsis,
                            )

                            val episodeLabel = buildEpisodeLabel(mediaContext)
                            if (episodeLabel != null) {
                                Spacer(modifier = Modifier.height(8.dp))
                                Surface(
                                    shape = RoundedCornerShape(999.dp),
                                    color = Color.White.copy(alpha = 0.08f),
                                ) {
                                    Text(
                                        text = episodeLabel,
                                        color = NetflixWhite.copy(alpha = 0.78f),
                                        fontSize = 11.sp,
                                        fontWeight = FontWeight.SemiBold,
                                        modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp),
                                    )
                                }
                            }

                            Spacer(modifier = Modifier.height(10.dp))

                            FlowRowLike(
                                items = buildList {
                                    mediaDetail?.releaseDate?.take(4)?.takeIf { it.isNotBlank() }?.let(::add)
                                    formatRuntimeLabel(playbackInfo?.duration ?: mediaDetail?.duration)?.let(::add)
                                    if ((playbackInfo?.subtitleTracks?.isNotEmpty() == true)) add("CC")
                                },
                            )

                            Spacer(modifier = Modifier.height(10.dp))

                            FlowRowLike(
                                items = buildList {
                                    formatResolutionLabel(playbackInfo?.height)?.let(::add)
                                    playbackInfo?.videoCodec?.uppercase()?.takeIf { it.isNotBlank() }?.let(::add)
                                    selectedAudio?.let { audio ->
                                        val lang = audio.language.ifBlank { "Audio" }
                                        add("$lang ${audio.codec.uppercase()} ${formatChannelLayout(audio.channels)}")
                                    }
                                },
                            )

                            mediaDetail?.overview?.takeIf { it.isNotBlank() }?.let { overview ->
                                Spacer(modifier = Modifier.height(12.dp))
                                Text(
                                    text = overview,
                                    color = NetflixWhite.copy(alpha = 0.68f),
                                    fontSize = 14.sp,
                                    lineHeight = 22.sp,
                                    maxLines = 4,
                                    overflow = TextOverflow.Ellipsis,
                                )
                            }
                        }
                    }
                }
            }

            DetailPanelUi.Season -> {
                Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
                    Row(
                        modifier = Modifier.horizontalScroll(rememberScrollState()),
                        horizontalArrangement = Arrangement.spacedBy(10.dp),
                    ) {
                        uiState.seasons.forEach { season ->
                            val isSelected = season.id == uiState.seasonPanelSeasonId
                            Surface(
                                shape = RoundedCornerShape(999.dp),
                                color = if (isSelected) NetflixWhite else Color.White.copy(alpha = 0.08f),
                                modifier = Modifier.clickable { onSelectSeason(season.id) },
                            ) {
                                Text(
                                    text = "Season ${season.seasonNumber}",
                                    color = if (isSelected) Color.Black else NetflixWhite.copy(alpha = 0.8f),
                                    fontSize = 14.sp,
                                    fontWeight = FontWeight.SemiBold,
                                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 10.dp),
                                )
                            }
                            Spacer(modifier = Modifier.width(10.dp))
                        }
                    }

                    Surface(
                        shape = RoundedCornerShape(28.dp),
                        color = Color.Black.copy(alpha = 0.45f),
                        modifier = Modifier.fillMaxWidth(),
                    ) {
                        when {
                            uiState.isSeasonPanelLoading -> {
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(vertical = 40.dp),
                                    contentAlignment = Alignment.Center,
                                ) {
                                    CircularProgressIndicator(color = NetflixWhite)
                                }
                            }

                            uiState.seasonPanelEpisodes.isEmpty() -> {
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(horizontal = 20.dp, vertical = 28.dp),
                                ) {
                                    Text(
                                        text = "No episodes found",
                                        color = NetflixWhite.copy(alpha = 0.48f),
                                        fontSize = 14.sp,
                                    )
                                }
                            }

                            else -> {
                                LazyRow(
                                    contentPadding = PaddingValues(horizontal = 20.dp, vertical = 18.dp),
                                    horizontalArrangement = Arrangement.spacedBy(14.dp),
                                ) {
                                    items(uiState.seasonPanelEpisodes, key = { it.id }) { episode ->
                                        val isPlayingEpisode = episode.mediaId == uiState.playbackInfo?.mediaId
                                        Surface(
                                            shape = RoundedCornerShape(22.dp),
                                            color = if (isPlayingEpisode) Color.White.copy(alpha = 0.14f) else Color.Black.copy(alpha = 0.55f),
                                            modifier = Modifier
                                                .width(220.dp)
                                                .clickable {
                                                    onSelectEpisode(episode.mediaId, isPlayingEpisode)
                                                },
                                        ) {
                                            Column {
                                                Box(
                                                    modifier = Modifier
                                                        .fillMaxWidth()
                                                        .height(124.dp)
                                                        .background(Color.Black.copy(alpha = 0.4f)),
                                                ) {
                                                    if (episode.still != null || episode.stillPath != null) {
                                                        if (episode.still != null) {
                                                            com.velox.app.presentation.ui.components.ResponsiveImage(
                                                                data = episode.still,
                                                                contentDescription = episode.title,
                                                                modifier = Modifier.fillMaxSize(),
                                                                contentScale = androidx.compose.ui.layout.ContentScale.Crop,
                                                            )
                                                        } else {
                                                            coil.compose.AsyncImage(
                                                                model = episode.stillPath,
                                                                contentDescription = episode.title,
                                                                modifier = Modifier.fillMaxSize(),
                                                                contentScale = androidx.compose.ui.layout.ContentScale.Crop,
                                                            )
                                                        }
                                                    } else {
                                                        Box(
                                                            modifier = Modifier.fillMaxSize(),
                                                            contentAlignment = Alignment.Center,
                                                        ) {
                                                            Icon(
                                                                LucideIcons.PlayArrow,
                                                                contentDescription = null,
                                                                tint = Color.White.copy(alpha = 0.25f),
                                                                modifier = Modifier.size(28.dp),
                                                            )
                                                        }
                                                    }

                                                    Row(
                                                        modifier = Modifier
                                                            .align(Alignment.TopStart)
                                                            .padding(12.dp),
                                                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                                                    ) {
                                                        Surface(
                                                            shape = RoundedCornerShape(999.dp),
                                                            color = Color.Black.copy(alpha = 0.58f),
                                                        ) {
                                                            Text(
                                                                text = "E${episode.episodeNumber}",
                                                                color = NetflixWhite.copy(alpha = 0.78f),
                                                                fontSize = 10.sp,
                                                                fontWeight = FontWeight.Bold,
                                                                modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp),
                                                            )
                                                        }
                                                        if (isPlayingEpisode) {
                                                            Surface(
                                                                shape = RoundedCornerShape(999.dp),
                                                                color = NetflixRed,
                                                            ) {
                                                                Text(
                                                                    text = "NOW PLAYING",
                                                                    color = NetflixWhite,
                                                                    fontSize = 10.sp,
                                                                    fontWeight = FontWeight.Black,
                                                                    modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp),
                                                                )
                                                            }
                                                        }
                                                    }
                                                }

                                                Column(
                                                    modifier = Modifier.padding(horizontal = 14.dp, vertical = 14.dp),
                                                    verticalArrangement = Arrangement.spacedBy(6.dp),
                                                ) {
                                                    Text(
                                                        text = episode.title,
                                                        color = NetflixWhite,
                                                        fontSize = 15.sp,
                                                        fontWeight = FontWeight.SemiBold,
                                                        maxLines = 2,
                                                        overflow = TextOverflow.Ellipsis,
                                                    )
                                                    Text(
                                                        text = buildEpisodeMetaLine(episode),
                                                        color = NetflixWhite.copy(alpha = 0.62f),
                                                        fontSize = 12.sp,
                                                        maxLines = 1,
                                                        overflow = TextOverflow.Ellipsis,
                                                    )
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }

            DetailPanelUi.None -> Unit
        }
    }
}

@Composable
internal fun FlowRowLike(items: List<String>) {
    if (items.isEmpty()) return
    Row(
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        items.forEachIndexed { index, item ->
            Text(
                text = item,
                color = NetflixWhite.copy(alpha = 0.78f),
                fontSize = 14.sp,
            )
            if (index != items.lastIndex) {
                Text(
                    text = "•",
                    color = NetflixWhite.copy(alpha = 0.4f),
                    fontSize = 12.sp,
                )
            }
        }
    }
}

private fun buildEpisodeLabel(mediaContext: com.velox.app.domain.model.MediaWithFilesInfo?): String? {
    if (mediaContext?.mediaType != "episode") return null
    val season = mediaContext.seasonNumber ?: return null
    val episode = mediaContext.episodeNumber ?: return null
    return "Episode S$season E$episode"
}

private fun formatRuntimeLabel(durationSeconds: Float?): String? {
    val seconds = durationSeconds?.toInt() ?: return null
    if (seconds <= 0) return null
    val hours = seconds / 3600
    val minutes = (seconds % 3600) / 60
    return if (hours > 0) "${hours}h ${minutes}m" else "${minutes}m"
}

private fun formatResolutionLabel(height: Int?): String? {
    val value = height ?: return null
    if (value <= 0) return null
    return "${value}p"
}

private fun buildEpisodeMetaLine(episode: com.velox.app.domain.model.Episode): String {
    val parts = mutableListOf("Episode ${episode.episodeNumber}")
    formatRuntimeLabel(episode.duration)?.let(parts::add)
    return parts.joinToString(" • ")
}
