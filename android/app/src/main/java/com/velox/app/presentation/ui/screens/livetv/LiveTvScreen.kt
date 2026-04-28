package com.velox.app.presentation.ui.screens.livetv

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.GridItemSpan
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.items as lazyItems
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.velox.app.domain.model.DataResult
import com.velox.app.domain.model.livetv.LiveProgram
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.ui.components.NotificationBell
import com.velox.app.presentation.ui.components.livetv.LiveTvChannelCard
import com.velox.app.presentation.ui.components.livetv.LiveTvEpgTimeline
import com.velox.app.presentation.ui.components.livetv.LiveTvMiniPlayer
import com.velox.app.presentation.viewmodel.livetv.LiveTvViewModel
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import java.time.Instant

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LiveTvScreen(
    onChannelClick: (Int) -> Unit,
    onSearchClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    onSettingsClick: () -> Unit,
    viewModel: LiveTvViewModel = hiltViewModel(),
) {
    val groupsResult by viewModel.groups.collectAsStateWithLifecycle()
    val channelsResult by viewModel.channels.collectAsStateWithLifecycle()
    val selectedGroup by viewModel.selectedGroup.collectAsStateWithLifecycle()
    val playingChannel by viewModel.playingChannel.collectAsStateWithLifecycle()
    val epgPrograms by viewModel.epgPrograms.collectAsStateWithLifecycle()
    val isRefreshing by viewModel.isRefreshing.collectAsStateWithLifecycle()
    val miniPlayerStreamUrl = remember(playingChannel?.id) {
        playingChannel?.id?.let { viewModel.streamUrl(it) }.orEmpty()
    }

    val channels = (channelsResult as? DataResult.Success)?.data.orEmpty()
    val groupTitles = (groupsResult as? DataResult.Success)?.data.orEmpty().map { it.groupTitle }

    val currentProgram = remember(epgPrograms) { findCurrentProgram(epgPrograms) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Live TV", color = NetflixWhite, fontWeight = FontWeight.Bold) },
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
    ) { paddingValues ->
        PullToRefreshBox(
            isRefreshing = isRefreshing,
            onRefresh = { viewModel.refresh() },
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues),
        ) {
        LazyVerticalGrid(
            columns = GridCells.Adaptive(150.dp),
            contentPadding = PaddingValues(bottom = 32.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
            modifier = Modifier.fillMaxSize(),
        ) {
            // 1. Hero Mini-Player
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
                        .aspectRatio(16f / 9f),
                )
            }

            // 2. EPG Timeline (only when programs available)
            if (epgPrograms.isNotEmpty()) {
                item(span = { GridItemSpan(maxLineSpan) }) {
                    LiveTvEpgTimeline(
                        programs = epgPrograms,
                        channelName = playingChannel?.name,
                    )
                }
            }

            // 3. Category Pills: "All" + groups + "Hidden"
            item(span = { GridItemSpan(maxLineSpan) }) {
                CategoryPills(
                    groups = groupTitles,
                    selectedGroup = selectedGroup,
                    isLoading = groupsResult == null,
                    onSelectGroup = { viewModel.selectGroup(it) },
                )
            }

            // 4. Grid Header
            item(span = { GridItemSpan(maxLineSpan) }) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp),
                    verticalAlignment = Alignment.Bottom,
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    Column {
                        Text(
                            "LIVE CHANNELS",
                            color = NetflixLightGray,
                            fontSize = 10.sp,
                            fontWeight = FontWeight.Bold,
                            letterSpacing = 2.sp,
                        )
                        Spacer(modifier = Modifier.height(2.dp))
                        Text(
                            text = headerTitle(selectedGroup),
                            color = NetflixWhite,
                            fontSize = 20.sp,
                            fontWeight = FontWeight.Bold,
                        )
                    }
                    Text(
                        text = "${channels.size} ${if (channels.size == 1) "channel" else "channels"}",
                        color = NetflixLightGray,
                        fontSize = 12.sp,
                    )
                }
            }

            // 5. Grid Items (or loading/empty state)
            when {
                channelsResult == null -> {
                    item(span = { GridItemSpan(maxLineSpan) }) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(200.dp),
                            contentAlignment = Alignment.Center,
                        ) {
                            CircularProgressIndicator(color = NetflixRed)
                        }
                    }
                }
                channels.isEmpty() && channelsResult is DataResult.Success -> {
                    item(span = { GridItemSpan(maxLineSpan) }) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(200.dp)
                                .padding(horizontal = 16.dp)
                                .background(NetflixDark),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text(
                                text = "No channels available in this category.",
                                color = NetflixLightGray,
                                fontSize = 13.sp,
                            )
                        }
                    }
                }
                channelsResult is DataResult.Error -> {
                    item(span = { GridItemSpan(maxLineSpan) }) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(200.dp)
                                .padding(horizontal = 16.dp),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text("Failed to load channels.", color = NetflixLightGray)
                        }
                    }
                }
                else -> {
                    items(channels, key = { it.id }) { channel ->
                        Box(modifier = Modifier.padding(horizontal = 4.dp)) {
                            LiveTvChannelCard(
                                channel = channel,
                                isActive = channel.id == playingChannel?.id,
                                onClick = {
                                    viewModel.persistLastWatched(channel)
                                    onChannelClick(channel.id)
                                },
                                onToggleHidden = { viewModel.toggleHidden(channel) },
                            )
                        }
                    }
                }
            }
        }
        } // PullToRefreshBox
    }
}

private fun findCurrentProgram(programs: List<LiveProgram>): LiveProgram? {
    if (programs.isEmpty()) return null
    val now = System.currentTimeMillis()
    return programs.firstOrNull { program ->
        val start = runCatching { Instant.parse(program.startTimeIso).toEpochMilli() }.getOrDefault(0L)
        val end = runCatching { Instant.parse(program.endTimeIso).toEpochMilli() }.getOrDefault(0L)
        start <= now && now < end
    }
}

private fun headerTitle(selectedGroup: String): String = when (selectedGroup) {
    LiveTvViewModel.GROUP_ALL -> "All Channels"
    else -> selectedGroup
}

@Composable
private fun CategoryPills(
    groups: List<String>,
    selectedGroup: String,
    isLoading: Boolean,
    onSelectGroup: (String) -> Unit,
) {
    if (isLoading) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(56.dp),
            contentAlignment = Alignment.Center,
        ) {
            CircularProgressIndicator(
                color = NetflixRed,
                modifier = Modifier.padding(8.dp),
            )
        }
        return
    }

    LazyRow(
        contentPadding = PaddingValues(horizontal = 16.dp, vertical = 4.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
        modifier = Modifier.fillMaxWidth(),
    ) {
        item(key = LiveTvViewModel.GROUP_ALL) {
            CategoryPill(
                label = "All",
                isSelected = selectedGroup == LiveTvViewModel.GROUP_ALL,
                onClick = { onSelectGroup(LiveTvViewModel.GROUP_ALL) },
            )
        }
        lazyItems(groups.filter { it.isNotBlank() }, key = { it }) { group ->
            CategoryPill(
                label = group,
                isSelected = selectedGroup == group,
                onClick = { onSelectGroup(group) },
            )
        }
        item(key = LiveTvViewModel.GROUP_HIDDEN) {
            CategoryPill(
                label = "Hidden",
                isSelected = selectedGroup == LiveTvViewModel.GROUP_HIDDEN,
                onClick = { onSelectGroup(LiveTvViewModel.GROUP_HIDDEN) },
            )
        }
    }
}

@Composable
private fun CategoryPill(label: String, isSelected: Boolean, onClick: () -> Unit) {
    Box(
        modifier = Modifier
            .clip(RoundedCornerShape(2.dp))
            .background(if (isSelected) NetflixWhite else NetflixDark)
            .clickable(onClick = onClick)
            .padding(horizontal = 16.dp, vertical = 8.dp),
    ) {
        Text(
            text = label,
            color = if (isSelected) NetflixBlack else NetflixLightGray,
            fontSize = 13.sp,
            fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Medium,
        )
    }
}
