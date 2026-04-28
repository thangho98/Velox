package com.velox.app.presentation.tv.screens

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.GridItemSpan
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.tv.material3.Border
import androidx.tv.material3.ClickableSurfaceDefaults
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.Surface
import androidx.tv.material3.Text
import coil.compose.AsyncImage
import com.velox.app.domain.model.DataResult
import com.velox.app.domain.model.livetv.LiveChannel
import com.velox.app.domain.model.livetv.LiveProgram
import com.velox.app.presentation.ui.components.livetv.LiveTvEpgTimeline
import com.velox.app.presentation.ui.components.livetv.LiveTvMiniPlayer
import com.velox.app.presentation.viewmodel.livetv.LiveTvViewModel
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import java.time.Instant

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun TvLiveTvScreen(
    onChannelClick: (Int) -> Unit,
    viewModel: LiveTvViewModel = hiltViewModel(),
) {
    val groupsResult by viewModel.groups.collectAsStateWithLifecycle()
    val channelsResult by viewModel.channels.collectAsStateWithLifecycle()
    val selectedGroup by viewModel.selectedGroup.collectAsStateWithLifecycle()
    val playingChannel by viewModel.playingChannel.collectAsStateWithLifecycle()
    val epgPrograms by viewModel.epgPrograms.collectAsStateWithLifecycle()
    val miniPlayerStreamUrl = remember(playingChannel?.id) {
        playingChannel?.id?.let { viewModel.streamUrl(it) }.orEmpty()
    }

    val channels = (channelsResult as? DataResult.Success)?.data.orEmpty()
    val groupTitles = (groupsResult as? DataResult.Success)?.data.orEmpty().map { it.groupTitle }

    val currentProgram = remember(epgPrograms) { findCurrentProgram(epgPrograms) }

    Row(modifier = Modifier.fillMaxSize().background(NetflixBlack)) {
        TvSidebar(
            selectedGroup = selectedGroup,
            groups = groupTitles,
            isLoading = groupsResult == null,
            onSelectGroup = { viewModel.selectGroup(it) },
            modifier = Modifier
                .weight(0.22f)
                .fillMaxHeight(),
        )

        LazyVerticalGrid(
            columns = GridCells.Fixed(5),
            contentPadding = PaddingValues(horizontal = 24.dp, vertical = 20.dp),
            horizontalArrangement = Arrangement.spacedBy(16.dp),
            verticalArrangement = Arrangement.spacedBy(20.dp),
            modifier = Modifier
                .weight(0.78f)
                .fillMaxHeight(),
        ) {
            item(span = { GridItemSpan(maxLineSpan) }) {
                LiveTvMiniPlayer(
                    channel = playingChannel,
                    currentProgram = currentProgram,
                    streamUrl = miniPlayerStreamUrl,
                    onWatch = { channel, _ ->
                        viewModel.persistLastWatched(channel)
                        onChannelClick(channel.id)
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(320.dp),
                )
            }

            if (epgPrograms.isNotEmpty()) {
                item(span = { GridItemSpan(maxLineSpan) }) {
                    LiveTvEpgTimeline(
                        programs = epgPrograms,
                        channelName = playingChannel?.name,
                    )
                }
            }

            item(span = { GridItemSpan(maxLineSpan) }) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.Bottom,
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    Column {
                        Text(
                            "LIVE CHANNELS",
                            color = NetflixLightGray,
                            fontSize = 11.sp,
                            fontWeight = FontWeight.Bold,
                            letterSpacing = 3.sp,
                        )
                        Spacer(modifier = Modifier.height(2.dp))
                        Text(
                            text = sidebarLabel(selectedGroup),
                            color = NetflixWhite,
                            fontSize = 22.sp,
                            fontWeight = FontWeight.Bold,
                        )
                    }
                    Text(
                        text = "${channels.size} ${if (channels.size == 1) "channel" else "channels"}",
                        color = NetflixLightGray,
                        fontSize = 13.sp,
                    )
                }
            }

            when {
                channelsResult == null -> {
                    item(span = { GridItemSpan(maxLineSpan) }) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(200.dp),
                            contentAlignment = Alignment.Center,
                        ) { Text("Loading…", color = NetflixLightGray) }
                    }
                }
                channels.isEmpty() && channelsResult is DataResult.Success -> {
                    item(span = { GridItemSpan(maxLineSpan) }) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(200.dp)
                                .background(NetflixDark),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text("No channels available in this category.", color = NetflixLightGray)
                        }
                    }
                }
                channelsResult is DataResult.Error -> {
                    item(span = { GridItemSpan(maxLineSpan) }) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(200.dp),
                            contentAlignment = Alignment.Center,
                        ) { Text("Failed to load channels.", color = NetflixLightGray) }
                    }
                }
                else -> {
                    items(channels, key = { it.id }) { channel ->
                        TvChannelCard(
                            channel = channel,
                            isActive = channel.id == playingChannel?.id,
                            onClick = {
                                viewModel.persistLastWatched(channel)
                                onChannelClick(channel.id)
                            },
                            onLongClick = { viewModel.toggleHidden(channel) },
                        )
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun TvSidebar(
    selectedGroup: String,
    groups: List<String>,
    isLoading: Boolean,
    onSelectGroup: (String) -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .background(Color(0xFF0B0B0B))
            .padding(horizontal = 20.dp, vertical = 28.dp),
    ) {
        Text(
            text = "LIVE TV",
            color = NetflixLightGray,
            fontSize = 11.sp,
            fontWeight = FontWeight.Bold,
            letterSpacing = 4.sp,
        )
        Spacer(modifier = Modifier.height(6.dp))
        Text(
            text = "Categories",
            color = NetflixWhite,
            fontSize = 22.sp,
            fontWeight = FontWeight.Bold,
        )
        Spacer(modifier = Modifier.height(20.dp))

        if (isLoading) {
            Text("Loading…", color = NetflixLightGray, fontSize = 13.sp)
            return@Column
        }

        val displayGroups = remember(groups) {
            buildList {
                add(LiveTvViewModel.GROUP_ALL to "All")
                groups.filter { it.isNotBlank() }.forEach { add(it to it) }
                add(LiveTvViewModel.GROUP_HIDDEN to "Hidden")
            }
        }

        LazyColumn(
            verticalArrangement = Arrangement.spacedBy(6.dp),
            modifier = Modifier.fillMaxWidth(),
        ) {
            items(displayGroups, key = { it.first }) { (key, label) ->
                val isSelected = selectedGroup == key
                Surface(
                    onClick = { onSelectGroup(key) },
                    colors = ClickableSurfaceDefaults.colors(
                        containerColor = if (isSelected) NetflixRed.copy(alpha = 0.22f) else Color.Transparent,
                        contentColor = if (isSelected) NetflixWhite else NetflixLightGray,
                        focusedContainerColor = NetflixRed,
                        focusedContentColor = NetflixWhite,
                        pressedContainerColor = NetflixRed,
                        pressedContentColor = NetflixWhite,
                    ),
                    shape = ClickableSurfaceDefaults.shape(shape = RoundedCornerShape(4.dp)),
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    Text(
                        text = label,
                        fontSize = 15.sp,
                        fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Medium,
                        modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
                    )
                }
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun TvChannelCard(
    channel: LiveChannel,
    isActive: Boolean,
    onClick: () -> Unit,
    onLongClick: () -> Unit,
) {
    val bg = remember(channel.id) { seedTvColor(channel) }
    val cleanName = remember(channel.name) { stripDecorations(channel.name) }

    Surface(
        onClick = onClick,
        onLongClick = onLongClick,
        colors = ClickableSurfaceDefaults.colors(
            containerColor = Color(0xFF181818),
            contentColor = NetflixWhite,
            focusedContainerColor = Color(0xFF1F1F1F),
            focusedContentColor = NetflixWhite,
            pressedContainerColor = Color(0xFF1F1F1F),
            pressedContentColor = NetflixWhite,
        ),
        shape = ClickableSurfaceDefaults.shape(shape = RoundedCornerShape(4.dp)),
        border = ClickableSurfaceDefaults.border(
            border = if (isActive) Border(androidx.compose.foundation.BorderStroke(2.dp, NetflixRed), shape = RoundedCornerShape(4.dp)) else Border.None,
            focusedBorder = Border(androidx.compose.foundation.BorderStroke(3.dp, NetflixWhite), shape = RoundedCornerShape(4.dp)),
        ),
        modifier = Modifier.fillMaxWidth(),
    ) {
        Column(modifier = Modifier.fillMaxWidth()) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .aspectRatio(4f / 3f)
                    .background(bg),
                contentAlignment = Alignment.Center,
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            Brush.verticalGradient(
                                0f to Color.White.copy(alpha = 0.06f),
                                1f to Color.Black.copy(alpha = 0.32f),
                            ),
                        ),
                )

                if (!channel.logo.isNullOrBlank()) {
                    AsyncImage(
                        model = channel.logo,
                        contentDescription = channel.name,
                        contentScale = ContentScale.Fit,
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(horizontal = 24.dp, vertical = 20.dp),
                    )
                } else {
                    Text(
                        text = cleanName,
                        color = NetflixWhite.copy(alpha = 0.92f),
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Black,
                        textAlign = TextAlign.Center,
                        maxLines = 3,
                        overflow = TextOverflow.Ellipsis,
                        modifier = Modifier.padding(horizontal = 8.dp),
                    )
                }

                Row(
                    modifier = Modifier
                        .align(Alignment.TopStart)
                        .padding(8.dp)
                        .clip(RoundedCornerShape(2.dp))
                        .background(NetflixBlack.copy(alpha = 0.7f))
                        .padding(horizontal = 6.dp, vertical = 3.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    com.velox.app.presentation.ui.components.livetv.LivePulseDot(
                        color = NetflixRed,
                        size = 5.dp,
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = "LIVE",
                        color = NetflixWhite,
                        fontSize = 10.sp,
                        fontWeight = FontWeight.Bold,
                        letterSpacing = 2.sp,
                    )
                }

                if (channel.isHidden) {
                    Row(
                        modifier = Modifier
                            .align(Alignment.TopEnd)
                            .padding(8.dp)
                            .clip(RoundedCornerShape(2.dp))
                            .background(Color.Black.copy(alpha = 0.75f))
                            .padding(horizontal = 6.dp, vertical = 3.dp),
                    ) {
                        Text(
                            text = "HIDDEN",
                            color = NetflixLightGray,
                            fontSize = 9.sp,
                            fontWeight = FontWeight.Bold,
                            letterSpacing = 2.sp,
                        )
                    }
                }

                if (isActive) {
                    Row(
                        modifier = Modifier
                            .align(Alignment.BottomEnd)
                            .padding(8.dp)
                            .clip(RoundedCornerShape(2.dp))
                            .background(NetflixRed)
                            .padding(horizontal = 6.dp, vertical = 3.dp),
                    ) {
                        Text(
                            text = "PLAYING",
                            color = NetflixWhite,
                            fontSize = 10.sp,
                            fontWeight = FontWeight.Bold,
                            letterSpacing = 2.sp,
                        )
                    }
                }
            }

            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 10.dp, vertical = 8.dp),
            ) {
                Text(
                    text = channel.name,
                    color = NetflixWhite,
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Bold,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = channel.groupTitle.ifBlank { "Live channel" },
                    color = NetflixLightGray,
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Medium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
    }
}

private val tvCardColors = listOf(
    Color(0xFF12234D),
    Color(0xFF3D141A),
    Color(0xFF2C1C08),
    Color(0xFF083023),
    Color(0xFF112933),
    Color(0xFF1F1540),
    Color(0xFF33152A),
    Color(0xFF1A1A1A),
)

private fun seedTvColor(channel: LiveChannel): Color {
    val seed = "${channel.name}:${channel.groupTitle}"
    var hash = 0
    for (ch in seed) hash = (hash * 31 + ch.code) % 997
    return tvCardColors[kotlin.math.abs(hash) % tvCardColors.size]
}

private val tvDecorationRegex = Regex("\\s*[(\\[][^)\\]]*[)\\]]")
private fun stripDecorations(name: String): String =
    name.replace(tvDecorationRegex, "").trim().ifEmpty { name }

private fun findCurrentProgram(programs: List<LiveProgram>): LiveProgram? {
    if (programs.isEmpty()) return null
    val now = System.currentTimeMillis()
    return programs.firstOrNull { p ->
        val s = runCatching { Instant.parse(p.startTimeIso).toEpochMilli() }.getOrDefault(0L)
        val e = runCatching { Instant.parse(p.endTimeIso).toEpochMilli() }.getOrDefault(0L)
        s <= now && now < e
    }
}

private fun sidebarLabel(selectedGroup: String): String = when (selectedGroup) {
    LiveTvViewModel.GROUP_ALL -> "All Channels"
    else -> selectedGroup
}
