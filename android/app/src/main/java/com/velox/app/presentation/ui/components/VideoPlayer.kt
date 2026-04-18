package com.velox.app.presentation.ui.components

import android.app.Activity
import android.app.PictureInPictureParams
import android.content.pm.ActivityInfo
import android.content.pm.PackageManager
import android.os.Build
import android.util.Rational
import android.view.ViewGroup
import android.view.WindowManager
import androidx.activity.compose.BackHandler
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectDragGestures
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
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
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.media3.common.util.UnstableApi
import androidx.media3.ui.AspectRatioFrameLayout
import androidx.media3.ui.PlayerView
import com.velox.app.R
import com.velox.app.presentation.cast.CastManager
import com.velox.app.presentation.cast.CastUiState
import com.velox.app.presentation.cast.rememberCastManager
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.ui.components.SubtitleSearchModal
import com.velox.app.presentation.viewmodel.AspectRatioUi
import com.velox.app.presentation.viewmodel.DetailPanelUi
import com.velox.app.presentation.viewmodel.PlayerUiState
import com.velox.app.presentation.viewmodel.PlayerViewModel
import com.velox.app.presentation.viewmodel.QualityOptionUi
import com.velox.app.presentation.viewmodel.RepeatModeUi
import com.velox.app.presentation.viewmodel.SkipSegmentUi
import com.velox.app.presentation.viewmodel.SubtitleBackgroundUi
import com.velox.app.presentation.viewmodel.SubtitleSizeUi
import com.velox.app.presentation.viewmodel.SubtitleTrackUi
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import com.velox.app.presentation.ui.components.player.PlayerControlsOverlay
import com.velox.app.presentation.ui.components.player.PlaybackStatsOverlay
import com.velox.app.presentation.ui.components.player.UpNextCard

@androidx.annotation.OptIn(UnstableApi::class)
@Composable
fun VideoPlayer(
    onBackClick: () -> Unit,
    onNavigateToMedia: (Int) -> Unit = {},
    uiState: PlayerUiState,
    actions: com.velox.app.presentation.viewmodel.PlayerActions,
    modifier: Modifier = Modifier,
) {

    val context = LocalContext.current
    val activity = remember(context) {
        var ctx = context
        while (ctx is android.content.ContextWrapper) {
            if (ctx is Activity) return@remember ctx
            ctx = ctx.baseContext
        }
        null
    }

    // Cast Manager
    val castManager = rememberCastManager()
    val castState by castManager.uiState.collectAsStateWithLifecycle()

    var isAnyMenuOpen by remember { mutableStateOf(false) }

    var isUserTouching by remember { mutableStateOf(false) }
    var lastInteractionTime by remember { mutableLongStateOf(0L) }

    // Auto-hide controls — never hide while buffering or loading
    LaunchedEffect(uiState.showControls, uiState.isPlaying, uiState.isLocked, isAnyMenuOpen, isUserTouching, lastInteractionTime, uiState.isBuffering, uiState.isLoading) {
        if (uiState.showControls && uiState.isPlaying && !uiState.isLocked && !isAnyMenuOpen && !isUserTouching && !uiState.isBuffering && !uiState.isLoading) {
            delay(3000)
            actions.toggleControls()
        }
    }

    // Update position periodically during playback
    LaunchedEffect(uiState.isPlaying) {
        while (uiState.isPlaying) {
            actions.updatePosition()
            delay(500)
        }
    }

    BackHandler {
        actions.saveProgress()
        onBackClick()
    }

    // Keep screen on while playing — handled via PlayerView.keepScreenOn in AndroidView update block

    // Double-tap seek handling (YouTube-style)
    var lastTapTime by remember { mutableLongStateOf(0L) }
    var lastTapX by remember { mutableFloatStateOf(0f) }
    var seekAccumulator by remember { mutableIntStateOf(0) }
    val coroutineScope = rememberCoroutineScope()
    var resetSeekJob by remember { mutableStateOf< kotlinx.coroutines.Job? >(null) }

    // Pan gesture state for volume/brightness feedback
    var panGestureVolume by remember { mutableFloatStateOf(0f) }
    var panGestureBrightness by remember { mutableFloatStateOf(0f) }
    var showVolumeFeedback by remember { mutableStateOf(false) }
    var showBrightnessFeedback by remember { mutableStateOf(false) }

    // Fullscreen state - controls orientation lock
    var isFullscreen by remember { mutableStateOf(false) }

    // Handle fullscreen orientation changes
    LaunchedEffect(isFullscreen) {
        val act = activity ?: run {
            android.util.Log.w("VideoPlayer", "toggleFullscreen: activity is null")
            return@LaunchedEffect
        }
        if (isFullscreen) {
            act.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_SENSOR_LANDSCAPE
        } else {
            // Force portrait first, then release to user preference after rotation completes
            act.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_PORTRAIT
            kotlinx.coroutines.delay(500)
            act.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_UNSPECIFIED
        }
    }

    // Reset orientation on back
    DisposableEffect(Unit) {
        onDispose {
            activity?.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_UNSPECIFIED
        }
    }

    fun toggleFullscreen() {
        isFullscreen = !isFullscreen
    }

    val supportsPiP = remember(activity) {
        Build.VERSION.SDK_INT >= Build.VERSION_CODES.O &&
            activity?.packageManager?.hasSystemFeature(PackageManager.FEATURE_PICTURE_IN_PICTURE) == true
    }

    // Seek amount in seconds (matching mobile)
    val SEEK_AMOUNT = 5

    BoxWithConstraints(
        modifier = modifier
            .fillMaxSize()
            .background(Color.Black)
            .pointerInput(Unit) {
                awaitPointerEventScope {
                    while (true) {
                        val event = awaitPointerEvent(androidx.compose.ui.input.pointer.PointerEventPass.Initial)
                        val pressed = event.changes.any { it.pressed }
                        if (isUserTouching != pressed) {
                            isUserTouching = pressed
                            if (!pressed) {
                                // when user lifts finger, reset interaction timer
                                lastInteractionTime = System.currentTimeMillis()
                            }
                        }
                    }
                }
            }
            .pointerInput(uiState.showControls) {
                detectTapGestures(
                    onTap = { offset ->
                        if (uiState.showSeekFeedback && seekAccumulator != 0) {
                            val screenWidth = size.width.toFloat()
                            val zoneWidth = screenWidth / 3
                            val isLeftZone = offset.x < zoneWidth
                            val isRightZone = offset.x > screenWidth - zoneWidth

                            val currentSide = uiState.seekFeedbackSide
                            if ((isLeftZone && currentSide == "left") || (isRightZone && currentSide == "right")) {
                                val direction = if (isRightZone) SEEK_AMOUNT else -SEEK_AMOUNT
                                seekAccumulator += direction
                                actions.showSeekFeedback(
                                    side = currentSide,
                                    amount = kotlin.math.abs(seekAccumulator),
                                )

                                resetSeekJob?.cancel()
                                resetSeekJob = coroutineScope.launch {
                                    delay(800)
                                    if (seekAccumulator < 0) {
                                        actions.seekBackward(kotlin.math.abs(seekAccumulator))
                                    } else {
                                        actions.seekForward(kotlin.math.abs(seekAccumulator))
                                    }
                                    seekAccumulator = 0
                                    actions.hideSeekFeedback()
                                }
                                return@detectTapGestures
                            }
                        }

                        if (uiState.showControls) {
                            actions.toggleControls()
                        } else {
                            actions.toggleControls()

                            if (uiState.isPlaying) {
                                // Calculate viewport bounds
                                val screenWidth = size.width.toFloat()
                                val screenHeight = size.height.toFloat()
                                val videoWidth = uiState.videoWidth.takeIf { it > 0 }?.toFloat() ?: 16f
                                val videoHeight = uiState.videoHeight.takeIf { it > 0 }?.toFloat() ?: 9f

                                val videoAspectRatio = videoWidth / videoHeight
                                val screenAspectRatio = screenWidth / screenHeight

                                var viewportWidth = screenWidth
                                var viewportHeight = screenHeight

                                if (uiState.aspectRatio == AspectRatioUi.Contain) {
                                    if (videoAspectRatio > screenAspectRatio) {
                                        viewportHeight = screenWidth / videoAspectRatio
                                    } else {
                                        viewportWidth = screenHeight * videoAspectRatio
                                    }
                                }

                                val letterboxPx = (screenHeight - viewportHeight) / 2f
                                val pillarboxPx = (screenWidth - viewportWidth) / 2f

                                val isInsideViewport = offset.x >= pillarboxPx && offset.x <= screenWidth - pillarboxPx &&
                                                       offset.y >= letterboxPx && offset.y <= screenHeight - letterboxPx

                                if (isInsideViewport) {
                                    actions.togglePlayPause()
                                }
                            }
                        }
                    },
                    onDoubleTap = { offset ->
                        if (uiState.isLocked) return@detectTapGestures

                        val screenWidth = size.width.toFloat()
                        val screenHeight = size.height.toFloat()
                        val zoneWidth = screenWidth / 3

                        // Ignore double taps in controls area (top 80px or bottom 160px)
                        if (uiState.showControls) {
                            if (offset.y < 80 || offset.y > screenHeight - 160) return@detectTapGestures
                        }

                        val tapX = offset.x
                        val isLeftZone = tapX < zoneWidth
                        val isRightZone = tapX > screenWidth - zoneWidth

                        if (isLeftZone || isRightZone) {
                            val direction = if (isRightZone) SEEK_AMOUNT else -SEEK_AMOUNT
                            if (seekAccumulator == 0) {
                                seekAccumulator = direction
                            } else {
                                seekAccumulator += direction
                            }

                            actions.showSeekFeedback(
                                side = if (isRightZone) "right" else "left",
                                amount = kotlin.math.abs(seekAccumulator),
                            )

                            // Reset after delay and perform seek
                            resetSeekJob?.cancel()
                            resetSeekJob = coroutineScope.launch {
                                delay(800)
                                if (seekAccumulator < 0) {
                                    actions.seekBackward(kotlin.math.abs(seekAccumulator))
                                } else {
                                    actions.seekForward(kotlin.math.abs(seekAccumulator))
                                }
                                seekAccumulator = 0
                                actions.hideSeekFeedback()
                            }
                        } else {
                            // Double tap in the center zone -> treat as single tap (toggle controls/pause)
                            if (!uiState.showControls) {
                                actions.toggleControls()
                                if (uiState.isPlaying) {
                                    actions.togglePlayPause()
                                }
                            }
                        }
                    },
                )
            }
            .pointerInput(uiState.showControls) {
                detectDragGestures(
                    onDragStart = { offset ->
                        if (uiState.isLocked || uiState.showControls) return@detectDragGestures
                    },
                    onDrag = { change, dragAmount ->
                        if (uiState.isLocked || uiState.showControls) return@detectDragGestures

                        val screenWidth = size.width.toFloat()
                        val absoluteX = change.position.x

                        // Vertical drag = brightness (left side) or volume (right side)
                        if (kotlin.math.abs(dragAmount.y) > kotlin.math.abs(dragAmount.x)) {
                            val delta = -dragAmount.y / 200f

                            if (absoluteX < screenWidth / 2) {
                                // Left side = brightness (visual feedback only)
                                panGestureBrightness = (panGestureBrightness + delta).coerceIn(-1f, 1f)
                                showBrightnessFeedback = true
                                showVolumeFeedback = false
                            } else {
                                // Right side = volume
                                val newVolume = (uiState.volume + delta).coerceIn(0f, 1f)
                                panGestureVolume = newVolume
                                showVolumeFeedback = true
                                showBrightnessFeedback = false
                                actions.setVolume(newVolume)
                            }
                        }
                    },
                    onDragEnd = {
                        showVolumeFeedback = false
                        showBrightnessFeedback = false
                        panGestureVolume = 0f
                        panGestureBrightness = 0f
                    },
                )
            },
    ) {
        // ExoPlayer — apply aspect ratio via resizeMode
        val currentResizeMode = when (uiState.aspectRatio) {
            AspectRatioUi.Contain -> AspectRatioFrameLayout.RESIZE_MODE_FIT
            AspectRatioUi.Cover -> AspectRatioFrameLayout.RESIZE_MODE_ZOOM
            AspectRatioUi.Fill -> AspectRatioFrameLayout.RESIZE_MODE_FILL
        }
        AndroidView(
            factory = { ctx ->
                PlayerView(ctx).apply {
                    layoutParams = ViewGroup.LayoutParams(
                        ViewGroup.LayoutParams.MATCH_PARENT,
                        ViewGroup.LayoutParams.MATCH_PARENT,
                    )
                    useController = false
                    player = actions.player
                    resizeMode = currentResizeMode
                }
            },
            update = { playerView ->
                playerView.resizeMode = currentResizeMode
                playerView.keepScreenOn = uiState.isPlaying || uiState.isLocked
            },
            modifier = Modifier.fillMaxSize(),
        )

        // Cinema Overlay (before playback)
        if (uiState.showCinemaOverlay) {
            CinemaOverlay(
                items = uiState.cinemaItems.map { item ->
                    CinemaItem(
                        type = item.type,
                        title = item.title,
                        url = item.url,
                        skippable = item.skippable,
                    )
                },
                onComplete = { actions.skipAllCinema() },
                onItemEnded = { actions.skipCinemaItem() },
                modifier = Modifier.fillMaxSize(),
            )
        }

        // Seek Feedback Overlay
        SeekFeedbackOverlayContainer(
            visible = uiState.showSeekFeedback,
            side = uiState.seekFeedbackSide,
            amount = uiState.seekFeedbackAmount,
            modifier = Modifier.fillMaxSize(),
        )

        // Gesture Feedback Overlay (volume/brightness)
        GestureFeedbackOverlay(
            volume = if (showVolumeFeedback) panGestureVolume else null,
            brightness = if (showBrightnessFeedback) panGestureBrightness else null,
            modifier = Modifier.fillMaxSize(),
        )

        // Subtitle Search Modal
        if (uiState.showSubtitleSearch) {
            SubtitleSearchModal(
                mediaId = actions.mediaId,
                onDismiss = { actions.hideSubtitleSearch() },
                onDownloaded = { actions.hideSubtitleSearch() },
                searchSubtitles = { id, lang -> actions.searchSubtitles(lang) },
                downloadSubtitle = { id, externalId, lang ->
                    actions.downloadSubtitle("opensubtitles", externalId, lang)
                },
                isSearching = uiState.isSearchingSubtitles,
                isDownloading = uiState.isDownloadingSubtitle,
                searchResults = uiState.subtitleSearchResults,
                selectedLang = uiState.subtitleSearchLang,
                onLangSelected = { lang -> actions.searchSubtitles(lang) },
                modifier = Modifier.fillMaxSize(),
            )
        }

        // Calculate letterbox bounds
        val letterboxPx = calculateLetterboxHeight(
            videoWidth = uiState.videoWidth,
            videoHeight = uiState.videoHeight,
            containerWidth = constraints.maxWidth,
            containerHeight = constraints.maxHeight
        )
        val density = androidx.compose.ui.platform.LocalDensity.current
        val letterboxDp = with(density) { letterboxPx.toDp() }

        // Dual Subtitle Overlay — pass appearance from settings
        val subtitleAppearance = SubtitleAppearance(
            size = when (uiState.subtitleSize) {
                SubtitleSizeUi.Small -> "small"
                SubtitleSizeUi.Medium -> "medium"
                SubtitleSizeUi.Large -> "large"
            },
            color = uiState.subtitleColor,
            background = when (uiState.subtitleBackground) {
                SubtitleBackgroundUi.None -> "none"
                SubtitleBackgroundUi.Semi -> "semi"
                SubtitleBackgroundUi.Solid -> "solid"
            },
        )
        DualSubtitleOverlay(
            primaryCues = uiState.primarySubtitleCues.map { SubtitleCue(it.start, it.end, it.text) },
            secondaryCues = uiState.secondarySubtitleCues.map { SubtitleCue(it.start, it.end, it.text) },
            currentPosition = uiState.currentPosition,
            visible = uiState.secondarySubtitleEnabled || uiState.primarySubtitleCues.isNotEmpty(),
            letterboxHeightDp = letterboxDp,
            isVideoReady = uiState.videoWidth > 0 && uiState.videoHeight > 0,
            appearance = subtitleAppearance,
            offsetSeconds = uiState.subtitleDelay,
            modifier = Modifier.fillMaxSize(),
        )

        // Skip Segments floating above controls - independent of control visibility
        // For series with next episode: credits segment triggers Up Next instead of Skip Credits
        val hasNextEpisode = uiState.nextEpisodeId != null
        if (uiState.skipSegments.isNotEmpty()) {
            val segment = uiState.skipSegments.find {
                uiState.currentPosition >= it.start && uiState.currentPosition < it.end
            }
            // Show Skip button only if: not credits for series with next episode
            if (segment != null && !(segment.type == "credits" && hasNextEpisode)) {
                Box(
                    modifier = Modifier
                        .align(Alignment.BottomEnd)
                        .padding(bottom = maxOf(180.dp, letterboxDp + 16.dp), end = 32.dp)
                ) {
                    TextButton(
                        onClick = { actions.skipSegment(segment) },
                        colors = ButtonDefaults.textButtonColors(contentColor = NetflixWhite),
                        shape = RoundedCornerShape(4.dp),
                        modifier = Modifier.background(Color.Black.copy(alpha = 0.8f), RoundedCornerShape(4.dp)).border(1.dp, Color.White.copy(alpha = 0.3f), RoundedCornerShape(4.dp))
                    ) {
                        Text("Skip ${segment.type.replaceFirstChar { it.uppercase() }}")
                        Spacer(modifier = Modifier.width(4.dp))
                        Icon(LucideIcons.SkipNext, contentDescription = null, modifier = Modifier.size(16.dp))
                    }
                }
            }
        }



        // Controls overlay - HIDE if error exists
        if (uiState.error == null) {
            PlayerControlsOverlay(
                uiState = uiState,
                letterboxDp = letterboxDp,
                onBackClick = {
                    actions.saveProgress()
                    onBackClick()
                },
                onPlayPauseClick = actions::togglePlayPause,
                onSeekForward = {
                    actions.seekForward(5)
                    actions.showSeekFeedback("right", 5)
                },
                onSeekBackward = {
                    actions.seekBackward(5)
                    actions.showSeekFeedback("left", 5)
                },
                onSeek = actions::seekTo,
                onSkipSegment = actions::skipSegment,
                onToggleControls = actions::toggleControls,
                onSetPlaybackSpeed = actions::setPlaybackSpeed,
                onToggleLock = actions::toggleLock,
                onSetVolume = actions::setVolume,
                onEnterPiP = {
                    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                        activity?.let {
                            if (it.packageManager.hasSystemFeature(PackageManager.FEATURE_PICTURE_IN_PICTURE)) {
                                val params = PictureInPictureParams.Builder()
                                    .setAspectRatio(Rational(16, 9))
                                    .build()
                                it.enterPictureInPictureMode(params)
                            }
                        }
                    }
                },
                onSetAspectRatio = actions::setAspectRatio,
                onSetQuality = actions::setMaxQuality,
                onSetSubtitleDelay = actions::setSubtitleDelay,
                onAdjustSubtitleDelay = actions::adjustSubtitleDelay,
                onSetSubtitleSize = actions::setSubtitleSize,
                onSetSubtitleColor = actions::setSubtitleColor,
                onSetSubtitleBackground = actions::setSubtitleBackground,
                onSetRepeatMode = actions::setRepeatMode,
                onTogglePlaybackStats = actions::togglePlaybackStats,
                onPlayNextEpisode = {
                    if (uiState.nextEpisodeWatched) {
                        actions.showWatchedWarning()
                    } else {
                        uiState.nextEpisodeId?.let { nextId ->
                            actions.saveProgress()
                            onNavigateToMedia(nextId)
                        }
                    }
                },
                onDismissUpNext = actions::dismissUpNext,
                supportsPiP = supportsPiP,
                isFullscreen = isFullscreen,
                onToggleFullscreen = ::toggleFullscreen,
                actions = actions,
                castState = castState,
                onCastClick = {
                    castManager.getCastSession()?.let { session ->
                        val device = session.castDevice
                        if (device != null) {
                            val streamUrl = uiState.videoUrl
                            val title = uiState.playbackInfo?.method ?: "Video"
                            if (streamUrl != null) {
                                castManager.loadMediaAndPlay(
                                    streamUrl = streamUrl,
                                    title = title,
                                    subtitle = null,
                                    positionMs = uiState.currentPosition,
                                    apiKey = null // API key will be appended by backend if needed
                                )
                            }
                        }
                    }
                },
                onMenuOpenChange = { isAnyMenuOpen = it },
                onToggleDetailPanel = actions::toggleDetailPanel,
                onSelectSeasonPanelSeason = actions::selectSeasonPanelSeason,
                onSelectEpisode = { episodeMediaId, isCurrentEpisode ->
                    if (isCurrentEpisode) {
                        actions.closeDetailPanel()
                    } else {
                        actions.saveProgress()
                        onNavigateToMedia(episodeMediaId)
                    }
                },
            )
        }

        // Center Overlays (Buffering, Paused)
        if (uiState.isBuffering) {
            androidx.compose.material3.CircularProgressIndicator(
                color = NetflixWhite,
                modifier = Modifier.align(Alignment.Center)
            )
        } else {
            androidx.compose.animation.AnimatedVisibility(
                visible = !uiState.isPlaying && uiState.error == null && uiState.showControls,
                enter = androidx.compose.animation.scaleIn(androidx.compose.animation.core.tween(200, easing = androidx.compose.animation.core.LinearOutSlowInEasing)) + androidx.compose.animation.fadeIn(),
                exit = androidx.compose.animation.scaleOut(androidx.compose.animation.core.tween(200, easing = androidx.compose.animation.core.FastOutLinearInEasing)) + androidx.compose.animation.fadeOut(),
                modifier = Modifier.align(Alignment.Center)
            ) {
                Box(
                    modifier = Modifier
                        .size(72.dp)
                        .background(Color.Black.copy(alpha = 0.5f), androidx.compose.foundation.shape.CircleShape)
                        .clickable(
                            interactionSource = remember { androidx.compose.foundation.interaction.MutableInteractionSource() },
                            indication = null,
                            onClick = { actions.togglePlayPause() }
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        LucideIcons.PlayArrow,
                        contentDescription = "Play",
                        tint = Color.White,
                        modifier = Modifier.size(48.dp)
                    )
                }
            }
        }

        // Playback Stats Overlay
        PlaybackStatsOverlay(
            visible = uiState.showPlaybackStats,
            playbackInfo = uiState.playbackInfo,
            onClose = actions::togglePlaybackStats,
            modifier = Modifier.fillMaxSize(),
        )

        // Up Next Card
        val upNextTitle = uiState.upNextTitle
        if (uiState.showUpNext && upNextTitle != null) {
            UpNextCard(
                title = upNextTitle,
                countdown = uiState.upNextCountdown,
                onPlayNext = {
                    if (uiState.nextEpisodeWatched) {
                        actions.showWatchedWarning()
                    } else {
                        uiState.upNextId?.let { nextId ->
                            actions.saveProgress()
                            onNavigateToMedia(nextId)
                        }
                    }
                },
                onDismiss = actions::dismissUpNext,
                bottomPadding = maxOf(214.dp, letterboxDp + 50.dp),
                modifier = Modifier.fillMaxSize(),
            )
        }

        // Watched Warning Dialog
        if (uiState.showWatchedWarning) {
            AlertDialog(
                onDismissRequest = actions::hideWatchedWarning,
                title = { Text(stringResource(R.string.player_watched_title), color = Color.White) },
                text = { Text(stringResource(R.string.player_watched_desc), color = Color.Gray) },
                confirmButton = {
                    TextButton(onClick = {
                        actions.hideWatchedWarning()
                        uiState.nextEpisodeId?.let { nextId ->
                            actions.resetProgressAndNavigate(nextId, onNavigateToMedia)
                        }
                    }) {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Icon(LucideIcons.Replay, contentDescription = null, modifier = Modifier.size(16.dp))
                            Spacer(Modifier.width(4.dp))
                            Text(stringResource(R.string.player_watch_again), color = MaterialTheme.colorScheme.primary)
                        }
                    }
                },
                dismissButton = {
                    uiState.nextNextEpisodeId?.let { nextNextId ->
                        TextButton(onClick = {
                            actions.hideWatchedWarning()
                            actions.saveProgress()
                            onNavigateToMedia(nextNextId)
                        }) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Icon(LucideIcons.SkipNext, contentDescription = null, modifier = Modifier.size(16.dp))
                                Spacer(Modifier.width(4.dp))
                                Text(stringResource(R.string.player_next_episode), color = Color.White)
                            }
                        }
                    }
                },
                containerColor = Color(0xFF1E1E1E),
                titleContentColor = Color.White,
                textContentColor = Color.Gray
            )
        }

        // Loading indicator
        if (uiState.isLoading) {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center,
            ) {
                CircularProgressIndicator(color = NetflixRed)
            }
        }

        // Fallback back button — only when controls are hidden during loading/buffering/error
        if (!uiState.showControls && (uiState.isLoading || uiState.isBuffering || uiState.error != null)) {
            IconButton(
                onClick = {
                    actions.saveProgress()
                    onBackClick()
                },
                modifier = Modifier
                    .align(Alignment.TopStart)
                    .statusBarsPadding()
                    .padding(start = 12.dp, top = 12.dp)
                    .size(40.dp)
                    .background(Color.Black.copy(alpha = 0.5f), CircleShape),
            ) {
                Icon(LucideIcons.ChevronLeft, contentDescription = "Back", tint = NetflixWhite, modifier = Modifier.size(24.dp))
            }
        }

        // Error state
        uiState.error?.let { error ->
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center,
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Icon(
                        LucideIcons.Error,
                        contentDescription = null,
                        tint = NetflixRed,
                        modifier = Modifier.size(64.dp),
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    Text(
                        text = error,
                        color = NetflixWhite,
                        fontSize = 16.sp,
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    Button(
                        onClick = { actions.loadPlaybackInfo() },
                        colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                    ) {
                        Text(stringResource(R.string.action_retry))
                    }
                }
            }
        }
    }
}
