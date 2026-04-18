package com.velox.app.presentation.ui.components.player

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material.icons.outlined.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.velox.app.R
import com.velox.app.presentation.ui.components.SubtitleMenuPopup
import com.velox.app.presentation.cast.CastUiState
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.ui.components.SubtitleSearchModal
import com.velox.app.presentation.viewmodel.AspectRatioUi
import com.velox.app.presentation.viewmodel.DetailPanelUi
import com.velox.app.presentation.viewmodel.PlayerUiState
import com.velox.app.presentation.viewmodel.PlayerActions
import com.velox.app.presentation.viewmodel.QualityOptionUi
import com.velox.app.presentation.viewmodel.RepeatModeUi
import com.velox.app.presentation.viewmodel.SkipSegmentUi
import com.velox.app.presentation.viewmodel.SubtitleBackgroundUi
import com.velox.app.presentation.viewmodel.SubtitleSizeUi
import com.velox.app.presentation.viewmodel.SubtitleTrackUi
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite

@Suppress("UNUSED_PARAMETER")
@Composable
internal fun PlayerControlsOverlay(
    uiState: PlayerUiState,
    letterboxDp: androidx.compose.ui.unit.Dp,
    onBackClick: () -> Unit,
    onPlayPauseClick: () -> Unit,
    onSeekForward: () -> Unit,
    onSeekBackward: () -> Unit,
    onSeek: (Long) -> Unit,
    onSkipSegment: (SkipSegmentUi) -> Unit,
    onToggleControls: () -> Unit,
    onSetPlaybackSpeed: (Float) -> Unit,
    onToggleLock: () -> Unit,
    onSetVolume: (Float) -> Unit,
    onEnterPiP: () -> Unit,
    onSetAspectRatio: (AspectRatioUi) -> Unit,
    onSetQuality: (Int) -> Unit,
    onSetSubtitleDelay: (Float) -> Unit,
    onAdjustSubtitleDelay: (Float) -> Unit,
    onSetSubtitleSize: (SubtitleSizeUi) -> Unit,
    onSetSubtitleColor: (String) -> Unit,
    onSetSubtitleBackground: (SubtitleBackgroundUi) -> Unit,
    onSetRepeatMode: (RepeatModeUi) -> Unit,
    onTogglePlaybackStats: () -> Unit,
    onPlayNextEpisode: () -> Unit,
    onDismissUpNext: () -> Unit,
    supportsPiP: Boolean,
    isFullscreen: Boolean,
    onToggleFullscreen: () -> Unit,
    actions: PlayerActions,
    castState: CastUiState,
    onCastClick: () -> Unit,
    onMenuOpenChange: (Boolean) -> Unit = {},
    onToggleDetailPanel: (DetailPanelUi) -> Unit,
    onSelectSeasonPanelSeason: (Int) -> Unit,
    onSelectEpisode: (Int, Boolean) -> Unit,
) {
    var showSubtitleMenu by remember { mutableStateOf(false) }
    var showSubtitleTranslate by remember { mutableStateOf(false) }
    var showAudioMenu by remember { mutableStateOf(false) }
    var showSpeedMenu by remember { mutableStateOf(false) }
    var showSettingsMenu by remember { mutableStateOf(false) }
    var showRemainingTime by remember { mutableStateOf(true) }
    val isTranslatingSubtitle by actions.isTranslatingSubtitle.collectAsStateWithLifecycle()

    LaunchedEffect(showSubtitleMenu, showSubtitleTranslate, showAudioMenu, showSpeedMenu, showSettingsMenu) {
        onMenuOpenChange(showSubtitleMenu || showSubtitleTranslate || showAudioMenu || showSpeedMenu || showSettingsMenu)
    }

    LaunchedEffect(uiState.activeDetailPanel) {
        if (uiState.activeDetailPanel != DetailPanelUi.None) {
            showSubtitleMenu = false
            showSubtitleTranslate = false
            showAudioMenu = false
            showSpeedMenu = false
            showSettingsMenu = false
        }
    }

    if (uiState.isLocked) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .clickable { onToggleLock() },
            contentAlignment = Alignment.TopEnd
        ) {
            IconButton(onClick = onToggleLock, modifier = Modifier.padding(16.dp)) {
                Icon(LucideIcons.Lock, contentDescription = "Unlock", tint = NetflixWhite)
            }
        }
        return
    }

    AnimatedVisibility(
        visible = uiState.showControls,
        enter = fadeIn(),
        exit = fadeOut(),
    ) {
        Box(
            modifier = Modifier.fillMaxSize(),
        ) {
            // Dark gradient overlay strictly at the bottom
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .align(Alignment.BottomCenter)
                    .height(200.dp)
                    .background(
                        androidx.compose.ui.graphics.Brush.verticalGradient(
                            colors = listOf(Color.Transparent, Color.Black.copy(alpha = 0.9f))
                        )
                    )
            )

            // Top Bar — back button only
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .statusBarsPadding()
                    .padding(start = 12.dp, top = 12.dp)
                    .align(Alignment.TopStart),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier
                        .background(Color.Black.copy(alpha = 0.4f), RoundedCornerShape(20.dp))
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onBackClick,
                        )
                        .padding(horizontal = 12.dp, vertical = 6.dp),
                ) {
                    Icon(LucideIcons.ChevronLeft, contentDescription = "Back", tint = NetflixWhite, modifier = Modifier.size(24.dp))
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(text = "Back", color = NetflixWhite, fontSize = 14.sp)
                }
            }

            AnimatedVisibility(
                visible = uiState.activeDetailPanel == DetailPanelUi.None,
                enter = fadeIn(),
                exit = fadeOut(),
                modifier = Modifier
                    .fillMaxWidth()
                    .align(Alignment.BottomCenter),
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 32.dp, vertical = 24.dp),
                ) {
                    // Title and Buttons Row
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(
                            text = uiState.mediaTitle ?: "Loading...",
                            color = NetflixWhite,
                            fontSize = 18.sp,
                            fontWeight = FontWeight.Bold,
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis,
                            modifier = Modifier.weight(1f).padding(end = 16.dp)
                        )

                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                            // CC Button
                            Box {
                                MetadataIconButton(LucideIcons.Subtitles, { showSubtitleMenu = true }, uiState.selectedSubtitleIndex != -1)
                                if (showSubtitleMenu) {
                                    SubtitleMenuPopup(
                                        uiState = uiState,
                                        onDismiss = { showSubtitleMenu = false },
                                        onSelectPrimary = { actions.selectSubtitleTrack(it) },
                                        onSelectSecondary = { actions.selectSecondarySubtitleTrack(it) },
                                        onSearchSubtitles = { actions.showSubtitleSearch() },
                                        onTranslateSubtitles = { showSubtitleTranslate = true }
                                    )
                                }
                            }

                            // Audio
                            if (uiState.audioTracks.isNotEmpty()) {
                                MetadataIconButton(LucideIcons.MusicTrack, { showAudioMenu = true })
                                DropdownMenu(expanded = showAudioMenu, onDismissRequest = { showAudioMenu = false }) {
                                    uiState.audioTracks.forEachIndexed { index, track ->
                                        DropdownMenuItem(
                                            text = {
                                                Text(buildAnnotatedString {
                                                    append(track.label)
                                                    if (track.codec.isNotBlank()) {
                                                        append(" ")
                                                        withStyle(SpanStyle(color = Color.Gray, fontSize = 12.sp)) {
                                                            append(track.codec.uppercase())
                                                            if (track.channels > 0) append(" ${track.channels}ch")
                                                        }
                                                    }
                                                    if (!track.isDefault) {
                                                        append(" ")
                                                        withStyle(SpanStyle(color = Color(0xFFFFD700), fontSize = 12.sp)) {
                                                            append("(HLS)")
                                                        }
                                                    }
                                                })
                                            },
                                            onClick = { actions.selectAudioTrack(index); showAudioMenu = false },
                                            leadingIcon = { if (index == uiState.selectedAudioTrackIndex) Icon(LucideIcons.Check, null) }
                                        )
                                    }
                                }
                            }

                            // Speed
                            val isSpeedActive = uiState.playbackSpeed != 1.0f
                            val speedText = if (uiState.playbackSpeed % 1.0f == 0.0f) "${uiState.playbackSpeed.toInt()}x" else "${uiState.playbackSpeed}x"
                            Box(
                                modifier = Modifier
                                    .clickable { showSpeedMenu = true }
                                    .background(Color.White.copy(alpha = if (isSpeedActive) 0.15f else 0.0f), RoundedCornerShape(6.dp))
                                    .border(1.dp, Color.White.copy(alpha = if (isSpeedActive) 0.8f else 0.3f), RoundedCornerShape(6.dp))
                                    .size(36.dp)
                            ) {
                                Text(
                                    text = speedText,
                                    color = NetflixWhite,
                                    fontSize = 13.sp,
                                    fontWeight = FontWeight.Bold,
                                    modifier = Modifier.align(Alignment.Center)
                                )
                                DropdownMenu(
                                    expanded = showSpeedMenu,
                                    onDismissRequest = { showSpeedMenu = false },
                                    modifier = Modifier.background(Color(0xFF1C1C1E))
                                ) {
                                    Text(
                                        text = "PLAYBACK SPEED",
                                        color = Color.White.copy(alpha = 0.5f),
                                        fontSize = 11.sp,
                                        fontWeight = FontWeight.SemiBold,
                                        letterSpacing = 1.sp,
                                        modifier = Modifier.padding(start = 24.dp, end = 24.dp, top = 8.dp, bottom = 12.dp)
                                    )

                                    listOf(0.25f, 0.5f, 0.75f, 1.0f, 1.25f, 1.5f, 1.75f, 2.0f).forEach { speed ->
                                        val itemText = if (speed == 1.0f) "Normal" else if (speed % 1.0f == 0.0f) "${speed.toInt()}x" else "${speed}x"
                                        val isSelected = speed == uiState.playbackSpeed

                                        DropdownMenuItem(
                                            text = {
                                                Text(
                                                    text = itemText,
                                                    color = if (isSelected) Color.White else Color.White.copy(alpha = 0.8f),
                                                    fontSize = 15.sp
                                                )
                                            },
                                            onClick = { onSetPlaybackSpeed(speed); showSpeedMenu = false },
                                            leadingIcon = {
                                                if (isSelected) {
                                                    Icon(LucideIcons.Check, null, tint = Color.White)
                                                } else {
                                                    Spacer(modifier = Modifier.width(24.dp))
                                                }
                                            },
                                            contentPadding = PaddingValues(horizontal = 16.dp, vertical = 4.dp)
                                        )
                                    }
                                }
                            }

                            // Settings
                            MetadataIconButton(LucideIcons.Settings, { showSettingsMenu = true })
                            SettingsMenu(
                                expanded = showSettingsMenu, onDismiss = { showSettingsMenu = false },
                                aspectRatio = uiState.aspectRatio, onSetAspectRatio = onSetAspectRatio,
                                maxQuality = uiState.maxQuality, qualityOptions = uiState.qualityOptions,
                                onSetQuality = onSetQuality, subtitleDelay = uiState.subtitleDelay,
                                onAdjustSubtitleDelay = onAdjustSubtitleDelay,
                                subtitleSize = uiState.subtitleSize, onSetSubtitleSize = onSetSubtitleSize,
                                subtitleColor = uiState.subtitleColor, onSetSubtitleColor = onSetSubtitleColor,
                                subtitleBackground = uiState.subtitleBackground, onSetSubtitleBackground = onSetSubtitleBackground,
                                repeatMode = uiState.repeatMode, onSetRepeatMode = onSetRepeatMode,
                                onTogglePlaybackStats = onTogglePlaybackStats
                            )

                            // Picture-in-Picture
                            if (supportsPiP) {
                                MetadataIconButton(
                                    icon = Icons.Outlined.PictureInPictureAlt,
                                    onClick = onEnterPiP,
                                    contentDescription = "Picture in Picture",
                                )
                            }

                            // Only show SkipNext for series episodes with a next episode
                            if (uiState.nextEpisodeId != null) {
                                MetadataIconButton(LucideIcons.SkipNext, onPlayNextEpisode)
                            }
                            MetadataIconButton(LucideIcons.LockOpen, onToggleLock)
                            MetadataIconButton(if (isFullscreen) LucideIcons.FullscreenExit else LucideIcons.Fullscreen, onToggleFullscreen)
                        }
                    }

                    Spacer(modifier = Modifier.height(16.dp))

                    // Seek bar Match Web
                    PlayerSeekBar(
                        currentPosition = uiState.currentPosition,
                        duration = uiState.effectiveDuration,
                        onSeek = onSeek,
                    )

                    Spacer(modifier = Modifier.height(2.dp))

                    // Bottom row with precise aligned play controls and timing details
                    Row(
                        modifier = Modifier.fillMaxWidth().height(48.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        val displayPos = if (uiState.effectiveDuration > 0 || !uiState.isLoading) formatTime(uiState.currentPosition) else "--:--"
                        Text(text = displayPos, color = NetflixWhite, fontSize = 14.sp)
                        Spacer(modifier = Modifier.width(32.dp))

                        IconButton(onClick = onSeekBackward) {
                            Icon(LucideIcons.Replay5, contentDescription = "Rewind", tint = NetflixWhite, modifier = Modifier.size(24.dp))
                        }
                        Spacer(modifier = Modifier.width(20.dp))

                        IconButton(onClick = onPlayPauseClick, modifier = Modifier.size(48.dp)) {
                            Icon(
                                if (uiState.isPlaying) LucideIcons.Pause else LucideIcons.PlayArrow,
                                contentDescription = if (uiState.isPlaying) "Pause" else "Play",
                                tint = NetflixWhite,
                                modifier = Modifier.size(36.dp)
                            )
                        }

                        Spacer(modifier = Modifier.width(20.dp))
                        IconButton(onClick = onSeekForward) {
                            Icon(LucideIcons.Forward5, contentDescription = "Forward", tint = NetflixWhite, modifier = Modifier.size(24.dp))
                        }

                        Spacer(modifier = Modifier.weight(1f))

                        val displayEnd = if (uiState.effectiveDuration > 0 || !uiState.isLoading) {
                            if (showRemainingTime) "-${formatTime((uiState.effectiveDuration - uiState.currentPosition).coerceAtLeast(0L))}"
                            else formatTime(uiState.effectiveDuration)
                        } else {
                            "--:--"
                        }
                        Text(
                            text = displayEnd,
                            color = NetflixWhite,
                            fontSize = 14.sp,
                            modifier = Modifier.clickable { showRemainingTime = !showRemainingTime }
                        )
                    }

                    Spacer(modifier = Modifier.height(4.dp))

                    Row(
                        horizontalArrangement = Arrangement.spacedBy(20.dp),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        DetailTabButton(
                            icon = LucideIcons.Info,
                            label = "Info",
                            isActive = false,
                            onClick = { onToggleDetailPanel(DetailPanelUi.Info) },
                        )
                        if (uiState.mediaContext?.mediaType == "episode" && uiState.seasons.isNotEmpty()) {
                            DetailTabButton(
                                icon = LucideIcons.ListIcon,
                                label = "Season",
                                isActive = false,
                                onClick = { onToggleDetailPanel(DetailPanelUi.Season) },
                            )
                        }
                    }
                }
            }

            AnimatedVisibility(
                visible = uiState.activeDetailPanel != DetailPanelUi.None,
                enter = fadeIn(),
                exit = fadeOut(),
                modifier = Modifier
                    .fillMaxWidth()
                    .align(Alignment.BottomCenter),
            ) {
                WatchDetailPanel(
                    uiState = uiState,
                    onToggleDetailPanel = onToggleDetailPanel,
                    onSelectSeason = onSelectSeasonPanelSeason,
                    onSelectEpisode = onSelectEpisode,
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 24.dp, vertical = 24.dp),
                )
            }
        }

        if (showSubtitleTranslate) {
            SubtitleTranslateSection(
                subtitles = uiState.subtitleTracks.filter { it.id > 0 },
                onTranslate = { subtitleId, targetLanguage ->
                    actions.translateSubtitle(subtitleId, targetLanguage)
                },
                onDismiss = { showSubtitleTranslate = false },
                isTranslating = isTranslatingSubtitle,
                modifier = Modifier.fillMaxSize(),
            )
        }
    }
}

@Composable
fun MetadataIconButton(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    onClick: () -> Unit,
    isActive: Boolean = false,
    contentDescription: String? = null,
) {
    IconButton(
        onClick = onClick,
        modifier = Modifier
            .background(Color.White.copy(alpha = if (isActive) 0.15f else 0.0f), RoundedCornerShape(6.dp))
            .border(1.dp, Color.White.copy(alpha = if (isActive) 0.8f else 0.3f), RoundedCornerShape(6.dp))
            .size(36.dp)
    ) {
        Icon(
            icon,
            contentDescription = contentDescription,
            tint = NetflixWhite,
            modifier = Modifier.size(18.dp)
        )
    }
}

@Composable
internal fun DetailTabButton(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    label: String,
    isActive: Boolean,
    onClick: () -> Unit,
) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(6.dp),
        modifier = Modifier.clickable(onClick = onClick),
    ) {
        Icon(
            icon,
            contentDescription = null,
            tint = if (isActive) NetflixWhite else NetflixWhite.copy(alpha = 0.45f),
            modifier = Modifier.size(16.dp),
        )
        Text(
            text = label,
            color = if (isActive) NetflixWhite else NetflixWhite.copy(alpha = 0.45f),
            fontSize = 14.sp,
            fontWeight = FontWeight.SemiBold,
        )
    }
}
