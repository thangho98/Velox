package com.velox.app.presentation.ui.components

import android.app.Activity
import com.velox.app.presentation.ui.components.LucideIcons
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
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.gestures.detectDragGestures
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material.icons.outlined.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.media3.common.util.UnstableApi
import androidx.media3.ui.AspectRatioFrameLayout
import androidx.media3.ui.PlayerView
import com.velox.app.presentation.viewmodel.AspectRatioUi
import com.velox.app.presentation.viewmodel.QualityOptionUi
import com.velox.app.presentation.viewmodel.RepeatModeUi
import com.velox.app.presentation.viewmodel.PlayerUiState
import com.velox.app.presentation.viewmodel.SubtitleTrackUi
import com.velox.app.presentation.viewmodel.PlayerViewModel
import com.velox.app.presentation.viewmodel.SubtitleSizeUi
import com.velox.app.presentation.viewmodel.SubtitleBackgroundUi
import com.velox.app.presentation.viewmodel.SkipSegmentUi
import com.velox.app.presentation.cast.CastManager
import com.velox.app.presentation.cast.CastUiState
import com.velox.app.presentation.cast.rememberCastManager
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme
import com.velox.app.presentation.ui.components.SubtitleSearchModal
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

@androidx.annotation.OptIn(UnstableApi::class)
@Composable
fun VideoPlayer(
    onBackClick: () -> Unit,
    onNavigateToMedia: (Int) -> Unit = {},
    viewModel: PlayerViewModel,
    modifier: Modifier = Modifier,
) {
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current
    val activity = context as? Activity

    // Cast Manager
    val castManager = rememberCastManager()
    val castState by castManager.uiState.collectAsState()

    var isAnyMenuOpen by remember { mutableStateOf(false) }
    
    var isUserTouching by remember { mutableStateOf(false) }
    var lastInteractionTime by remember { mutableLongStateOf(0L) }

    // Auto-hide controls
    LaunchedEffect(uiState.showControls, uiState.isPlaying, uiState.isLocked, isAnyMenuOpen, isUserTouching, lastInteractionTime) {
        if (uiState.showControls && uiState.isPlaying && !uiState.isLocked && !isAnyMenuOpen && !isUserTouching) {
            delay(3000)
            viewModel.toggleControls()
        }
    }

    // Update position periodically during playback
    LaunchedEffect(uiState.isPlaying) {
        while (uiState.isPlaying) {
            viewModel.updatePosition()
            delay(500)
        }
    }

    BackHandler {
        viewModel.saveProgress()
        onBackClick()
    }

    // Handle locked state - keep screen on
    LaunchedEffect(uiState.isLocked) {
        activity?.let {
            if (uiState.isLocked) {
                it.window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
            } else {
                it.window.clearFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
            }
        }
    }

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
        activity?.let {
            if (isFullscreen) {
                it.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_LANDSCAPE
            } else {
                it.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_PORTRAIT
            }
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
            .pointerInput(Unit) {
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
                                viewModel.showSeekFeedback(
                                    side = currentSide,
                                    amount = kotlin.math.abs(seekAccumulator),
                                )
                                
                                resetSeekJob?.cancel()
                                resetSeekJob = coroutineScope.launch {
                                    delay(800)
                                    if (seekAccumulator < 0) {
                                        viewModel.seekBackward(kotlin.math.abs(seekAccumulator))
                                    } else {
                                        viewModel.seekForward(kotlin.math.abs(seekAccumulator))
                                    }
                                    seekAccumulator = 0
                                    viewModel.hideSeekFeedback()
                                }
                                return@detectTapGestures
                            }
                        }
                        
                        viewModel.togglePlayPause() 
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

                            viewModel.showSeekFeedback(
                                side = if (isRightZone) "right" else "left",
                                amount = kotlin.math.abs(seekAccumulator),
                            )

                            // Reset after delay and perform seek
                            resetSeekJob?.cancel()
                            resetSeekJob = coroutineScope.launch {
                                delay(800)
                                if (seekAccumulator < 0) {
                                    viewModel.seekBackward(kotlin.math.abs(seekAccumulator))
                                } else {
                                    viewModel.seekForward(kotlin.math.abs(seekAccumulator))
                                }
                                seekAccumulator = 0
                                viewModel.hideSeekFeedback()
                            }
                        }
                    },
                )
            }
            .pointerInput(Unit) {
                detectDragGestures(
                    onDragStart = { offset ->
                        if (uiState.isLocked) return@detectDragGestures
                        // Reset state on drag start
                    },
                    onDrag = { change, dragAmount ->
                        if (uiState.isLocked) return@detectDragGestures

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
                                viewModel.setVolume(newVolume)
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
                    player = viewModel.player
                    resizeMode = currentResizeMode
                }
            },
            update = { playerView ->
                playerView.resizeMode = currentResizeMode
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
                onComplete = { viewModel.skipAllCinema() },
                onItemEnded = { viewModel.skipCinemaItem() },
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
                mediaId = viewModel.mediaId,
                onDismiss = { viewModel.hideSubtitleSearch() },
                onDownloaded = { viewModel.hideSubtitleSearch() },
                searchSubtitles = { id, lang -> viewModel.searchSubtitles(lang) },
                downloadSubtitle = { id, externalId, lang ->
                    viewModel.downloadSubtitle("opensubtitles", externalId, lang)
                },
                isSearching = uiState.isSearchingSubtitles,
                isDownloading = uiState.isDownloadingSubtitle,
                searchResults = uiState.subtitleSearchResults,
                selectedLang = uiState.subtitleSearchLang,
                onLangSelected = { lang -> viewModel.searchSubtitles(lang) },
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
                        onClick = { viewModel.skipSegment(segment) },
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

        // Center Overlays (Buffering, Paused)
        if (uiState.isBuffering) {
            androidx.compose.material3.CircularProgressIndicator(
                color = NetflixWhite,
                modifier = Modifier.align(Alignment.Center)
            )
        } else {
            androidx.compose.animation.AnimatedVisibility(
                visible = !uiState.isPlaying && uiState.error == null,
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
                            onClick = { viewModel.togglePlayPause() }
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

        // Controls overlay - HIDE if error exists
        if (uiState.error == null) {
            PlayerControlsOverlay(
                uiState = uiState,
                letterboxDp = letterboxDp,
                onBackClick = {
                    viewModel.saveProgress()
                    onBackClick()
                },
                onPlayPauseClick = viewModel::togglePlayPause,
                onSeekForward = viewModel::seekForward,
                onSeekBackward = viewModel::seekBackward,
                onSeek = viewModel::seekTo,
                onSkipSegment = viewModel::skipSegment,
                onToggleControls = viewModel::toggleControls,
                onSetPlaybackSpeed = viewModel::setPlaybackSpeed,
                onToggleLock = viewModel::toggleLock,
                onSetVolume = viewModel::setVolume,
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
                onSetAspectRatio = viewModel::setAspectRatio,
                onSetQuality = viewModel::setMaxQuality,
                onSetSubtitleDelay = viewModel::setSubtitleDelay,
                onAdjustSubtitleDelay = viewModel::adjustSubtitleDelay,
                onSetSubtitleSize = viewModel::setSubtitleSize,
                onSetSubtitleColor = viewModel::setSubtitleColor,
                onSetSubtitleBackground = viewModel::setSubtitleBackground,
                onSetRepeatMode = viewModel::setRepeatMode,
                onTogglePlaybackStats = viewModel::togglePlaybackStats,
                onPlayNextEpisode = {
                    uiState.nextEpisodeId?.let { nextId ->
                        viewModel.saveProgress()
                        onNavigateToMedia(nextId)
                    }
                },
                onDismissUpNext = viewModel::dismissUpNext,
                isFullscreen = isFullscreen,
                onToggleFullscreen = ::toggleFullscreen,
                viewModel = viewModel,
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
                onMenuOpenChange = { isAnyMenuOpen = it }
            )
        }

        // Playback Stats Overlay
        PlaybackStatsOverlay(
            visible = uiState.showPlaybackStats,
            playbackInfo = uiState.playbackInfo,
            onClose = viewModel::togglePlaybackStats,
            modifier = Modifier.fillMaxSize(),
        )

        // Up Next Card
        val upNextTitle = uiState.upNextTitle
        if (uiState.showUpNext && upNextTitle != null) {
            UpNextCard(
                title = upNextTitle,
                countdown = uiState.upNextCountdown,
                onPlayNext = {
                    uiState.upNextId?.let { nextId ->
                        viewModel.saveProgress()
                        onNavigateToMedia(nextId)
                    }
                },
                onDismiss = viewModel::dismissUpNext,
                bottomPadding = maxOf(214.dp, letterboxDp + 50.dp),
                modifier = Modifier.fillMaxSize(),
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
                        onClick = { viewModel.loadPlaybackInfo() },
                        colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                    ) {
                        Text("Retry")
                    }
                }
            }
        }
    }
}

@OptIn(androidx.compose.material3.ExperimentalMaterial3Api::class)
@Composable
private fun PlayerControlsOverlay(
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
    isFullscreen: Boolean,
    onToggleFullscreen: () -> Unit,
    viewModel: PlayerViewModel,
    castState: CastUiState,
    onCastClick: () -> Unit,
    onMenuOpenChange: (Boolean) -> Unit = {}
) {
    var showSubtitleMenu by remember { mutableStateOf(false) }
    var showAudioMenu by remember { mutableStateOf(false) }
    var showSpeedMenu by remember { mutableStateOf(false) }
    var showSettingsMenu by remember { mutableStateOf(false) }

    LaunchedEffect(showSubtitleMenu, showAudioMenu, showSpeedMenu, showSettingsMenu) {
        onMenuOpenChange(showSubtitleMenu || showAudioMenu || showSpeedMenu || showSettingsMenu)
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
            modifier = Modifier
                .fillMaxSize()
                .clickable(
                    indication = null,
                    interactionSource = remember { MutableInteractionSource() },
                    onClick = onToggleControls,
                ),
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

            // Top Bar
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp)
                    .align(Alignment.TopCenter),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    IconButton(onClick = onBackClick) {
                        Icon(LucideIcons.ChevronLeft, contentDescription = "Back", tint = NetflixWhite)
                    }
                    Text(text = "Back", color = NetflixWhite, fontSize = 16.sp)
                }

                // Volume slider top right
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier.width(160.dp).padding(horizontal = 8.dp)
                ) {
                    Icon(
                        if (uiState.volume == 0f) LucideIcons.VolumeOff else LucideIcons.VolumeUp,
                        contentDescription = "Volume",
                        tint = NetflixWhite,
                        modifier = Modifier.size(20.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Slider(
                        value = uiState.volume,
                        onValueChange = onSetVolume,
                        thumb = {
                            Box(
                                modifier = Modifier
                                    .size(12.dp)
                                    .background(NetflixWhite, CircleShape)
                            )
                        },
                        track = { sliderState ->
                            SliderDefaults.Track(
                                colors = SliderDefaults.colors(
                                    activeTrackColor = NetflixWhite,
                                    inactiveTrackColor = NetflixWhite.copy(alpha = 0.3f),
                                ),
                                sliderState = sliderState,
                                modifier = Modifier.height(4.dp)
                            )
                        },
                        modifier = Modifier.weight(1f)
                    )
                }
            }



            // Bottom controls
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .align(Alignment.BottomCenter)
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
                                    onSelectPrimary = { viewModel.selectSubtitleTrack(it) },
                                    onSelectSecondary = { viewModel.selectSecondarySubtitleTrack(it) },
                                    onSearchSubtitles = { viewModel.showSubtitleSearch() },
                                    onTranslateSubtitles = { /* no-op for now */ }
                                )
                            }
                        }

                        // Audio
                        if (uiState.audioTracks.isNotEmpty()) {
                            MetadataIconButton(LucideIcons.MusicTrack, { showAudioMenu = true })
                            DropdownMenu(expanded = showAudioMenu, onDismissRequest = { showAudioMenu = false }) {
                                uiState.audioTracks.forEachIndexed { index, track ->
                                    DropdownMenuItem(
                                        text = { Text(track.label) },
                                        onClick = { viewModel.selectAudioTrack(index); showAudioMenu = false },
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
                    duration = uiState.duration,
                    onSeek = onSeek,
                )

                Spacer(modifier = Modifier.height(2.dp))

                // Bottom row with precise aligned play controls and timing details
                Row(
                    modifier = Modifier.fillMaxWidth().height(48.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(text = formatTime(uiState.currentPosition), color = NetflixWhite, fontSize = 14.sp)
                    Spacer(modifier = Modifier.width(32.dp))
                    
                    // Rewind
                    IconButton(onClick = onSeekBackward) {
                        Icon(LucideIcons.Replay10, contentDescription = "Rewind", tint = NetflixWhite, modifier = Modifier.size(24.dp))
                    }
                    Spacer(modifier = Modifier.width(20.dp))

                    // Play/Pause
                    IconButton(onClick = onPlayPauseClick, modifier = Modifier.size(48.dp)) {
                        Icon(
                            if (uiState.isPlaying) LucideIcons.Pause else LucideIcons.PlayArrow,
                            contentDescription = if (uiState.isPlaying) "Pause" else "Play",
                            tint = NetflixWhite,
                            modifier = Modifier.size(36.dp)
                        )
                    }
                    
                    Spacer(modifier = Modifier.width(20.dp))
                    // Forward
                    IconButton(onClick = onSeekForward) {
                        Icon(LucideIcons.Forward10, contentDescription = "Forward", tint = NetflixWhite, modifier = Modifier.size(24.dp))
                    }

                    Spacer(modifier = Modifier.weight(1f))

                    Text(text = "-${formatTime(uiState.duration - uiState.currentPosition)}", color = NetflixWhite, fontSize = 14.sp)
                }
            }
        }
    }
}

@Composable
fun MetadataIconButton(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    onClick: () -> Unit,
    isActive: Boolean = false
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
            contentDescription = null,
            tint = NetflixWhite,
            modifier = Modifier.size(18.dp)
        )
    }
}

@OptIn(androidx.compose.material3.ExperimentalMaterial3Api::class)
@Composable
private fun PlayerSeekBar(
    currentPosition: Long,
    duration: Long,
    onSeek: (Long) -> Unit,
) {
    var sliderPosition by remember { mutableFloatStateOf(0f) }
    var isDragging by remember { mutableStateOf(false) }

    LaunchedEffect(currentPosition) {
        if (!isDragging && duration > 0) {
            sliderPosition = (currentPosition.toFloat() / duration.toFloat()).coerceIn(0f, 1f)
        }
    }

    Slider(
        value = sliderPosition,
        onValueChange = { value ->
            isDragging = true
            sliderPosition = value
        },
        onValueChangeFinished = {
            isDragging = false
            onSeek((sliderPosition * duration).toLong())
        },
        thumb = {
            Box(
                modifier = Modifier
                    .size(14.dp)
                    .background(NetflixWhite, CircleShape)
            )
        },
        track = { sliderState ->
            SliderDefaults.Track(
                colors = SliderDefaults.colors(
                    activeTrackColor = NetflixWhite,
                    inactiveTrackColor = NetflixWhite.copy(alpha = 0.3f),
                ),
                sliderState = sliderState,
                modifier = Modifier.height(4.dp)
            )
        },
        modifier = Modifier.fillMaxWidth(),
    )
}

private fun formatTime(ms: Long): String {
    if (ms <= 0) return "0:00"
    val totalSeconds = ms / 1000
    val hours = totalSeconds / 3600
    val minutes = (totalSeconds % 3600) / 60
    val seconds = totalSeconds % 60
    return if (hours > 0) {
        String.format("%d:%02d:%02d", hours, minutes, seconds)
    } else {
        String.format("%d:%02d", minutes, seconds)
    }
}

@Composable
private fun SettingsSectionTitle(title: String, icon: androidx.compose.ui.graphics.vector.ImageVector) {
    Row(verticalAlignment = Alignment.CenterVertically, modifier = Modifier.padding(bottom = 8.dp)) {
        Icon(icon, contentDescription = null, modifier = Modifier.size(14.dp), tint = Color.White.copy(alpha = 0.4f))
        Spacer(modifier = Modifier.width(5.dp))
        Text(
            text = title,
            color = Color.White.copy(alpha = 0.4f),
            fontSize = 10.sp,
            fontWeight = FontWeight.SemiBold,
            letterSpacing = 1.sp
        )
    }
}

@Composable
private fun SettingsMenu(
    expanded: Boolean,
    onDismiss: () -> Unit,
    aspectRatio: AspectRatioUi,
    onSetAspectRatio: (AspectRatioUi) -> Unit,
    maxQuality: Int,
    qualityOptions: List<QualityOptionUi>,
    onSetQuality: (Int) -> Unit,
    subtitleDelay: Float,
    onAdjustSubtitleDelay: (Float) -> Unit,
    subtitleSize: SubtitleSizeUi,
    onSetSubtitleSize: (SubtitleSizeUi) -> Unit,
    subtitleColor: String,
    onSetSubtitleColor: (String) -> Unit,
    subtitleBackground: SubtitleBackgroundUi,
    onSetSubtitleBackground: (SubtitleBackgroundUi) -> Unit,
    repeatMode: RepeatModeUi,
    onSetRepeatMode: (RepeatModeUi) -> Unit,
    onTogglePlaybackStats: () -> Unit,
) {
    if (!expanded) return
    // Two-view navigation like webapp: 'main' or 'quality'
    var settingsView by remember { mutableStateOf("main") }

    androidx.compose.ui.window.Popup(
        alignment = Alignment.BottomEnd,
        offset = androidx.compose.ui.unit.IntOffset(-40, -180),
        onDismissRequest = onDismiss,
        properties = androidx.compose.ui.window.PopupProperties(focusable = true)
    ) {
        Box(
            modifier = Modifier
                .width(240.dp)
                .background(Color(0xFF1E1E1E), RoundedCornerShape(12.dp))
                .border(1.dp, Color.White.copy(alpha = 0.1f), RoundedCornerShape(12.dp))
        ) {
            if (settingsView == "quality") {
                // Quality submenu (separate view with back button — matches webapp)
                Column {
                    // Header with back button
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable { settingsView = "main" }
                            .padding(horizontal = 16.dp, vertical = 10.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        Icon(LucideIcons.ChevronLeft, contentDescription = null, modifier = Modifier.size(14.dp), tint = Color.White.copy(0.7f))
                        Text("Quality", color = Color.White.copy(0.7f), fontSize = 12.sp, fontWeight = FontWeight.SemiBold)
                    }
                    Divider(color = Color.White.copy(alpha = 0.1f), thickness = 1.dp)

                    // Quality options list
                    Column(
                        modifier = Modifier
                            .heightIn(max = 300.dp)
                            .verticalScroll(rememberScrollState())
                            .padding(vertical = 4.dp)
                    ) {
                        qualityOptions.forEach { option ->
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .clickable { onSetQuality(option.height); settingsView = "main" }
                                    .then(if (maxQuality == option.height) Modifier.background(Color.White.copy(alpha = 0.1f)) else Modifier)
                                    .padding(horizontal = 16.dp, vertical = 8.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.SpaceBetween,
                            ) {
                                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                                    Text(
                                        text = option.label,
                                        color = if (maxQuality == option.height) Color.White else Color.White.copy(alpha = 0.7f),
                                        fontSize = 12.sp,
                                    )
                                    if (option.instant) {
                                        Icon(LucideIcons.FlashOn, contentDescription = null, modifier = Modifier.size(11.dp), tint = Color(0xFFFACC15))
                                    }
                                }
                                if (maxQuality == option.height) {
                                    Icon(LucideIcons.Check, contentDescription = null, modifier = Modifier.size(14.dp), tint = Color.White)
                                }
                            }
                        }
                        // Auto option at bottom with separator
                        Divider(color = Color.White.copy(alpha = 0.1f), thickness = 1.dp)
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .clickable { onSetQuality(0); settingsView = "main" }
                                .then(if (maxQuality == 0) Modifier.background(Color.White.copy(alpha = 0.1f)) else Modifier)
                                .padding(horizontal = 16.dp, vertical = 8.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.SpaceBetween,
                        ) {
                            Text("Auto", color = if (maxQuality == 0) Color.White else Color.White.copy(alpha = 0.7f), fontSize = 12.sp)
                            if (maxQuality == 0) {
                                Icon(LucideIcons.Check, contentDescription = null, modifier = Modifier.size(14.dp), tint = Color.White)
                            }
                        }
                    }
                }
            } else {
                // Main settings view
                Column(
                    modifier = Modifier
                        .heightIn(max = 480.dp)
                        .verticalScroll(rememberScrollState())
                        .padding(12.dp)
                ) {
                    // ASPECT RATIO
                    SettingsSectionTitle("ASPECT RATIO", LucideIcons.Fullscreen)
                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                        val options = listOf("Auto", "Cover", "Fill")
                        val selectedIndex = when (aspectRatio) { AspectRatioUi.Contain -> 0; AspectRatioUi.Cover -> 1; AspectRatioUi.Fill -> 2 }
                        options.forEachIndexed { i, text ->
                            val isSelected = selectedIndex == i
                            Box(
                                modifier = Modifier.weight(1f).height(32.dp)
                                    .background(if (isSelected) Color.White.copy(alpha = 0.2f) else Color.White.copy(alpha = 0.05f), RoundedCornerShape(8.dp))
                                    .clickable { onSetAspectRatio(AspectRatioUi.values()[i]) },
                                contentAlignment = Alignment.Center
                            ) {
                                Text(text, color = if (isSelected) Color.White else Color.White.copy(0.7f), fontSize = 12.sp, fontWeight = FontWeight.Medium)
                            }
                        }
                    }

                    Divider(color = Color.White.copy(alpha = 0.1f), thickness = 1.dp, modifier = Modifier.padding(vertical = 10.dp))

                    // SUBTITLES
                    SettingsSectionTitle("SUBTITLES", LucideIcons.Subtitles)
                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                        val sizeOptions = listOf("S" to SubtitleSizeUi.Small, "M" to SubtitleSizeUi.Medium, "L" to SubtitleSizeUi.Large)
                        sizeOptions.forEach { (text, size) ->
                            val isSelected = subtitleSize == size
                            Box(modifier = Modifier.weight(1f).height(32.dp)
                                .background(if (isSelected) Color.White.copy(alpha = 0.2f) else Color.White.copy(alpha = 0.05f), RoundedCornerShape(8.dp))
                                .clickable { onSetSubtitleSize(size) },
                                contentAlignment = Alignment.Center) {
                                Text(text, color = if (isSelected) Color.White else Color.White.copy(0.7f), fontSize = 12.sp, fontWeight = FontWeight.Medium)
                            }
                        }
                    }
                    Spacer(Modifier.height(4.dp))
                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                        val bgOptions = listOf("None" to SubtitleBackgroundUi.None, "Semi" to SubtitleBackgroundUi.Semi, "Solid" to SubtitleBackgroundUi.Solid)
                        bgOptions.forEach { (text, bg) ->
                            val isSelected = subtitleBackground == bg
                            Box(modifier = Modifier.weight(1f).height(32.dp)
                                .background(if (isSelected) Color.White.copy(alpha = 0.2f) else Color.White.copy(alpha = 0.05f), RoundedCornerShape(8.dp))
                                .clickable { onSetSubtitleBackground(bg) },
                                contentAlignment = Alignment.Center) {
                                Text(text, color = if (isSelected) Color.White else Color.White.copy(0.7f), fontSize = 12.sp, fontWeight = FontWeight.Medium)
                            }
                        }
                    }
                    Spacer(Modifier.height(8.dp))
                    Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                        val colorOptions = listOf("#ffffff" to Color.White, "#fde047" to Color(0xFFFDE047), "#4ade80" to Color(0xFF4ADE80), "#60a5fa" to Color(0xFF60A5FA))
                        colorOptions.forEach { (hex, color) ->
                            val isSelected = subtitleColor == hex
                            Box(modifier = Modifier.size(20.dp)
                                .background(color, CircleShape)
                                .border(if (isSelected) 2.dp else 1.dp, if (isSelected) Color.White else Color.White.copy(alpha = 0.2f), CircleShape)
                                .clickable { onSetSubtitleColor(hex) })
                        }
                    }
                    Spacer(Modifier.height(10.dp))
                    // Delay container with subtle background (matches webapp)
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(Color.White.copy(alpha = 0.04f), RoundedCornerShape(8.dp))
                            .padding(horizontal = 12.dp, vertical = 8.dp)
                    ) {
                        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                            Text("Delay", color = Color.White.copy(0.7f), fontSize = 12.sp)
                            Text(if (subtitleDelay > 0) "+${String.format("%.2f", subtitleDelay)}s" else "${String.format("%.2f", subtitleDelay)}s", color = Color.White, fontSize = 12.sp, fontWeight = FontWeight.SemiBold)
                        }
                    }
                    Spacer(Modifier.height(6.dp))
                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                        Box(modifier = Modifier.weight(1f).height(32.dp)
                            .background(Color.White.copy(alpha = 0.05f), RoundedCornerShape(8.dp))
                            .clickable { onAdjustSubtitleDelay(-0.25f) }, contentAlignment = Alignment.Center) {
                            Text("-0.25s", color = Color.White.copy(0.7f), fontSize = 12.sp, fontWeight = FontWeight.Medium)
                        }
                        Box(modifier = Modifier.weight(1f).height(32.dp)
                            .background(Color.White.copy(alpha = 0.05f), RoundedCornerShape(8.dp))
                            .clickable { onAdjustSubtitleDelay(-subtitleDelay) }, contentAlignment = Alignment.Center) {
                            Text("Reset", color = Color.White.copy(0.7f), fontSize = 12.sp, fontWeight = FontWeight.Medium)
                        }
                        Box(modifier = Modifier.weight(1f).height(32.dp)
                            .background(Color.White.copy(alpha = 0.05f), RoundedCornerShape(8.dp))
                            .clickable { onAdjustSubtitleDelay(0.25f) }, contentAlignment = Alignment.Center) {
                            Text("+0.25s", color = Color.White.copy(0.7f), fontSize = 12.sp, fontWeight = FontWeight.Medium)
                        }
                    }

                    Divider(color = Color.White.copy(alpha = 0.1f), thickness = 1.dp, modifier = Modifier.padding(vertical = 10.dp))

                    // QUALITY — clickable row that navigates to quality submenu
                    Row(
                        modifier = Modifier.fillMaxWidth()
                            .clip(RoundedCornerShape(8.dp))
                            .clickable { settingsView = "quality" }
                            .padding(horizontal = 12.dp, vertical = 6.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Icon(LucideIcons.FlashOn, contentDescription = null, modifier = Modifier.size(14.dp), tint = Color.White.copy(0.5f))
                            Spacer(Modifier.width(5.dp))
                            Text("Quality", color = Color.White.copy(0.7f), fontSize = 12.sp)
                        }
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Text(if (maxQuality == 0) "Auto" else "${maxQuality}p", color = Color.White.copy(0.5f), fontSize = 12.sp)
                            Icon(LucideIcons.ChevronRight, contentDescription = null, tint = Color.White.copy(0.5f), modifier = Modifier.size(14.dp))
                        }
                    }

                    Divider(color = Color.White.copy(alpha = 0.1f), thickness = 1.dp, modifier = Modifier.padding(vertical = 10.dp))

                    // REPEAT
                    SettingsSectionTitle("REPEAT", LucideIcons.Repeat)
                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                        val repeatOptions = listOf("None" to RepeatModeUi.None, "One" to RepeatModeUi.One, "All" to RepeatModeUi.All)
                        repeatOptions.forEach { (text, mode) ->
                            val isSelected = repeatMode == mode
                            Box(
                                modifier = Modifier.weight(1f).height(32.dp)
                                    .background(if (isSelected) Color.White.copy(alpha = 0.2f) else Color.White.copy(alpha = 0.05f), RoundedCornerShape(8.dp))
                                    .clickable { onSetRepeatMode(mode) },
                                contentAlignment = Alignment.Center
                            ) {
                                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(3.dp)) {
                                    if (isSelected) Icon(if (mode == RepeatModeUi.One) LucideIcons.RepeatOne else LucideIcons.Repeat, contentDescription = null, modifier = Modifier.size(14.dp), tint = Color.White)
                                    Text(text, color = if (isSelected) Color.White else Color.White.copy(0.7f), fontSize = 12.sp, fontWeight = FontWeight.Medium)
                                }
                            }
                        }
                    }

                    Divider(color = Color.White.copy(alpha = 0.1f), thickness = 1.dp, modifier = Modifier.padding(vertical = 10.dp))

                    // PLAYBACK INFO
                    Row(
                        modifier = Modifier.fillMaxWidth()
                            .clip(RoundedCornerShape(8.dp))
                            .clickable { onTogglePlaybackStats(); onDismiss() }
                            .padding(horizontal = 12.dp, vertical = 6.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(LucideIcons.ShowChart, contentDescription = null, modifier = Modifier.size(14.dp), tint = Color.White.copy(0.5f))
                        Spacer(Modifier.width(5.dp))
                        Text("Playback Info", color = Color.White.copy(0.7f), fontSize = 12.sp)
                    }
                }
            }
        }
    }
}

@Composable
private fun PlaybackStatsOverlay(
    visible: Boolean,
    playbackInfo: com.velox.app.domain.model.PlaybackInfo?,
    onClose: () -> Unit,
    modifier: Modifier = Modifier,
) {
    if (visible && playbackInfo != null) {
        val isTranscoding = playbackInfo.method == "FullTranscode" || playbackInfo.method == "TranscodeAudio"
        val isPreTranscode = playbackInfo.method == "PreTranscode"

        // Use pretranscode details when available; fallback to original if "copy" or 0
        val ptCodecValid = playbackInfo.ptVideoCodec != null && playbackInfo.ptVideoCodec != "copy"
        val ptHeightValid = playbackInfo.ptHeight != null && playbackInfo.ptHeight > 0
        val videoCodec = if (isPreTranscode && ptCodecValid) playbackInfo.ptVideoCodec else playbackInfo.videoCodec
        val videoHeight = if (isPreTranscode && ptHeightValid) playbackInfo.ptHeight!! else playbackInfo.height ?: 0
        val videoWidth = if (isPreTranscode && ptHeightValid && playbackInfo.width != null && playbackInfo.height != null && playbackInfo.height > 0) {
            (playbackInfo.width.toFloat() / playbackInfo.height * playbackInfo.ptHeight!!).toInt()
        } else playbackInfo.width ?: 0
        val videoBitrate = if (isPreTranscode && playbackInfo.ptVideoBitrate != null && playbackInfo.ptVideoBitrate > 0) playbackInfo.ptVideoBitrate else playbackInfo.bitrate ?: 0

        val selectedAudio = playbackInfo.audioTracks.find { it.selected }
            ?: playbackInfo.audioTracks.find { it.isDefault }
            ?: playbackInfo.audioTracks.firstOrNull()

        val monoFamily = androidx.compose.ui.text.font.FontFamily.Monospace

        Box(
            modifier = modifier
                .fillMaxSize()
                .padding(start = 16.dp, top = 80.dp),
        ) {
            Surface(
                modifier = Modifier
                    .width(320.dp)
                    .align(Alignment.TopStart),
                shape = RoundedCornerShape(12.dp),
                color = Color.Black.copy(alpha = 0.7f),
            ) {
                Box {
                    Column {
                        // Playback Method
                        Column(modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                Text("Playback", color = Color.White, fontSize = 14.sp, fontWeight = FontWeight.Bold)
                                MethodBadge(playbackInfo.method)
                            }
                            if (playbackInfo.decisionReason != null) {
                                Spacer(Modifier.height(6.dp))
                                Text(
                                    text = playbackInfo.decisionReason,
                                    color = Color.White.copy(alpha = 0.6f),
                                    fontSize = 12.sp,
                                    fontFamily = monoFamily,
                                    lineHeight = 18.sp,
                                )
                            }
                        }

                        Divider(color = Color.White.copy(alpha = 0.1f), thickness = 1.dp)

                        // Video
                        Column(modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                Text("Video", color = Color.White, fontSize = 14.sp, fontWeight = FontWeight.Bold)
                                when {
                                    playbackInfo.method == "FullTranscode" -> StatusBadge("Transcoding", Color(0xFFF87171))
                                    isPreTranscode -> StatusBadge("PreTranscode", Color(0xFF22D3EE))
                                    else -> StatusBadge("Direct", Color(0xFF4ADE80))
                                }
                            }
                            Spacer(Modifier.height(6.dp))
                            val codecLine = buildString {
                                append(videoCodec?.uppercase() ?: "—")
                                if (videoHeight > 0) append("  ${videoWidth}×${videoHeight}")
                                if (!isPreTranscode && playbackInfo.videoProfile != null) append("  ${playbackInfo.videoProfile}")
                                if (!isPreTranscode && playbackInfo.videoLevel != null) append("  L${playbackInfo.videoLevel}")
                            }
                            Text(codecLine, color = Color.White.copy(alpha = 0.8f), fontSize = 12.sp, fontFamily = monoFamily, lineHeight = 18.sp)

                            val bitrateLineParts = mutableListOf<String>()
                            if (videoBitrate > 0) {
                                bitrateLineParts.add(if (videoBitrate >= 1000) "${"%.1f".format(videoBitrate / 1000f)} Mbps" else "$videoBitrate Kbps")
                            }
                            if (playbackInfo.videoFps != null && playbackInfo.videoFps > 0) {
                                val fpsStr = if (playbackInfo.videoFps == playbackInfo.videoFps.toInt().toFloat()) "${playbackInfo.videoFps.toInt()} fps" else "${"%.2f".format(playbackInfo.videoFps)} fps"
                                bitrateLineParts.add(fpsStr)
                            }
                            if (bitrateLineParts.isNotEmpty()) {
                                Text(bitrateLineParts.joinToString(" · "), color = Color.White.copy(alpha = 0.6f), fontSize = 12.sp, fontFamily = monoFamily, lineHeight = 18.sp)
                            }
                            Spacer(Modifier.height(4.dp))
                            Text("Dropped: 0", color = Color.White.copy(alpha = 0.6f), fontSize = 12.sp, fontFamily = monoFamily)
                        }

                        // Audio
                        if (selectedAudio != null) {
                            Divider(color = Color.White.copy(alpha = 0.1f), thickness = 1.dp)
                            Column(modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)) {
                                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                    Text("Audio", color = Color.White, fontSize = 14.sp, fontWeight = FontWeight.Bold)
                                    when {
                                        isTranscoding -> StatusBadge("Transcoding", Color(0xFFFACC15))
                                        isPreTranscode -> StatusBadge("PreTranscode", Color(0xFF22D3EE))
                                        else -> StatusBadge("Direct", Color(0xFF4ADE80))
                                    }
                                }
                                Spacer(Modifier.height(6.dp))
                                val channelLayout = formatChannelLayout(selectedAudio.channels)
                                val audioLine = buildString {
                                    append(selectedAudio.codec.uppercase())
                                    append(" $channelLayout")
                                    if (selectedAudio.language.isNotEmpty()) append(" · ${selectedAudio.language}")
                                    if (selectedAudio.isDefault) append(" (Default)")
                                }
                                Text(audioLine, color = Color.White.copy(alpha = 0.8f), fontSize = 12.sp, fontFamily = monoFamily, lineHeight = 18.sp)

                                val audioBitrateParts = mutableListOf<String>()
                                if (selectedAudio.bitrate != null && selectedAudio.bitrate > 0) {
                                    audioBitrateParts.add(if (selectedAudio.bitrate >= 1000) "${selectedAudio.bitrate / 1000} Kbps" else "${selectedAudio.bitrate} bps")
                                }
                                if (selectedAudio.sampleRate != null && selectedAudio.sampleRate > 0) {
                                    audioBitrateParts.add("${selectedAudio.sampleRate} Hz")
                                }
                                if (audioBitrateParts.isNotEmpty()) {
                                    Text(audioBitrateParts.joinToString(" · "), color = Color.White.copy(alpha = 0.6f), fontSize = 12.sp, fontFamily = monoFamily, lineHeight = 18.sp)
                                }
                            }
                        }

                        // Stream
                        Divider(color = Color.White.copy(alpha = 0.1f), thickness = 1.dp)
                        Column(modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)) {
                            Text("Stream", color = Color.White, fontSize = 14.sp, fontWeight = FontWeight.Bold)
                            Spacer(Modifier.height(6.dp))
                            val streamType = when {
                                playbackInfo.method == "DirectPlay" -> "HTTP Range"
                                isPreTranscode -> "HTTP Range (PreTranscode)"
                                else -> "HLS"
                            }
                            val containerStr = if (isPreTranscode) "MP4" else playbackInfo.container?.uppercase() ?: "—"
                            val streamLine = buildString {
                                append("$streamType · $containerStr")
                                if (!isPreTranscode && playbackInfo.fileSize != null && playbackInfo.fileSize > 0) {
                                    append(" · ${"%.1f".format(playbackInfo.fileSize / (1024.0 * 1024.0 * 1024.0))} GB")
                                }
                            }
                            Text(streamLine, color = Color.White.copy(alpha = 0.8f), fontSize = 12.sp, fontFamily = monoFamily, lineHeight = 18.sp)

                            if (playbackInfo.estimatedBitrate != null && playbackInfo.estimatedBitrate > 0 && (isTranscoding || isPreTranscode)) {
                                val estStr = if (playbackInfo.estimatedBitrate >= 1000) "${"%.1f".format(playbackInfo.estimatedBitrate / 1000f)} Mbps" else "${playbackInfo.estimatedBitrate} Kbps"
                                Text("Estimated: $estStr", color = Color.White.copy(alpha = 0.6f), fontSize = 12.sp, fontFamily = monoFamily, lineHeight = 18.sp)
                            }
                        }
                    }

                    // Close button
                    IconButton(
                        onClick = onClose,
                        modifier = Modifier.align(Alignment.TopEnd).size(36.dp),
                    ) {
                        Icon(LucideIcons.Close, contentDescription = "Close", tint = Color.White.copy(alpha = 0.4f), modifier = Modifier.size(16.dp))
                    }
                }
            }
        }
    }
}

@Composable
private fun MethodBadge(method: String) {
    val (bgColor, textColor, label) = when (method) {
        "DirectPlay" -> Triple(Color(0xFF4ADE80).copy(alpha = 0.2f), Color(0xFF4ADE80), "Direct Play")
        "DirectStream" -> Triple(Color(0xFF60A5FA).copy(alpha = 0.2f), Color(0xFF60A5FA), "Direct Stream")
        "PreTranscode" -> Triple(Color(0xFF22D3EE).copy(alpha = 0.2f), Color(0xFF22D3EE), "PreTranscode")
        "TranscodeAudio" -> Triple(Color(0xFFFACC15).copy(alpha = 0.2f), Color(0xFFFACC15), "Transcode Audio")
        "FullTranscode" -> Triple(Color(0xFFF87171).copy(alpha = 0.2f), Color(0xFFF87171), "Full Transcode")
        else -> Triple(Color.White.copy(alpha = 0.1f), Color.White.copy(alpha = 0.6f), method)
    }
    Surface(shape = RoundedCornerShape(4.dp), color = bgColor) {
        Text(label, color = textColor, fontSize = 10.sp, fontWeight = FontWeight.SemiBold, modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp))
    }
}

@Composable
private fun StatusBadge(label: String, color: Color) {
    Surface(shape = RoundedCornerShape(4.dp), color = color.copy(alpha = 0.2f)) {
        Text(label, color = color, fontSize = 10.sp, fontWeight = FontWeight.SemiBold, modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp))
    }
}

private fun formatChannelLayout(channels: Int): String {
    return when (channels) {
        1 -> "1.0"
        2 -> "2.0"
        6 -> "5.1"
        8 -> "7.1"
        else -> "$channels.0"
    }
}

@Composable
private fun UpNextCard(
    title: String,
    countdown: Int,
    onPlayNext: () -> Unit,
    onDismiss: () -> Unit,
    bottomPadding: androidx.compose.ui.unit.Dp = 180.dp,
    modifier: Modifier = Modifier,
) {
    Box(
        modifier = modifier,
        contentAlignment = Alignment.BottomEnd,
    ) {
        Surface(
            modifier = Modifier.padding(end = 16.dp, bottom = bottomPadding),
            shape = RoundedCornerShape(8.dp),
            color = Color.Black.copy(alpha = 0.9f),
        ) {
        Column(
            modifier = Modifier.padding(12.dp),
        ) {
            Text(
                text = "Up next",
                color = NetflixWhite.copy(alpha = 0.6f),
                fontSize = 10.sp,
                fontWeight = FontWeight.Bold,
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = title,
                color = NetflixWhite,
                fontSize = 14.sp,
                maxLines = 2,
            )
            Spacer(modifier = Modifier.height(8.dp))
            Row(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Button(
                    onClick = onPlayNext,
                    colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                    contentPadding = PaddingValues(horizontal = 12.dp, vertical = 4.dp),
                ) {
                    Icon(
                        LucideIcons.PlayArrow,
                        contentDescription = null,
                        modifier = Modifier.size(16.dp),
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = "Play Next",
                        fontSize = 12.sp,
                    )
                }
                TextButton(
                    onClick = onDismiss,
                    colors = ButtonDefaults.textButtonColors(contentColor = NetflixWhite.copy(alpha = 0.6f)),
                ) {
                    Text(
                        text = "Dismiss",
                        fontSize = 12.sp,
                    )
                }
            }
        }
    }
    }
}

@Composable
fun SubtitleTranslateSection(
    subtitles: List<SubtitleTrackUi>,
    onTranslate: (subtitleId: Int, targetLanguage: String) -> Unit,
    onDismiss: () -> Unit,
    isTranslating: Boolean,
    modifier: Modifier = Modifier,
) {
    val translateLanguages = listOf(
        "vi" to "Vietnamese",
        "en" to "English",
        "fr" to "French",
        "de" to "German",
        "es" to "Spanish",
        "ja" to "Japanese",
        "ko" to "Korean",
        "zh" to "Chinese",
    )

    var selectedSubtitleId by remember { mutableIntStateOf(subtitles.firstOrNull()?.index ?: -1) }
    var selectedTargetLang by remember { mutableStateOf<String?>(null) }

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        containerColor = NetflixBlack,
        modifier = modifier,
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 20.dp)
                .padding(bottom = 32.dp),
        ) {
            Text(
                text = "Translate Subtitle",
                color = NetflixWhite,
                fontSize = 18.sp,
                fontWeight = FontWeight.Bold,
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Source subtitle selector (if multiple)
            if (subtitles.size > 1) {
                Text(
                    text = "Source Subtitle",
                    color = NetflixWhite.copy(alpha = 0.7f),
                    fontSize = 14.sp,
                )
                Spacer(modifier = Modifier.height(8.dp))
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    subtitles.filter { it.index >= 0 }.forEach { subtitle ->
                        FilterChip(
                            selected = selectedSubtitleId == subtitle.index,
                            onClick = { selectedSubtitleId = subtitle.index },
                            label = { Text(subtitle.label) },
                            colors = FilterChipDefaults.filterChipColors(
                                containerColor = Color(0xFF2A2A2A),
                                labelColor = NetflixWhite,
                                selectedContainerColor = NetflixRed,
                                selectedLabelColor = NetflixWhite,
                            ),
                        )
                    }
                }
                Spacer(modifier = Modifier.height(16.dp))
            }

            // Target language selector
            Text(
                text = "Translate To",
                color = NetflixWhite.copy(alpha = 0.7f),
                fontSize = 14.sp,
            )
            Spacer(modifier = Modifier.height(8.dp))
            Row(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                translateLanguages.forEach { lang ->
                    FilterChip(
                        selected = selectedTargetLang == lang.first,
                        onClick = { selectedTargetLang = lang.first },
                        label = { Text(lang.second) },
                        colors = FilterChipDefaults.filterChipColors(
                            containerColor = Color(0xFF2A2A2A),
                            labelColor = NetflixWhite,
                            selectedContainerColor = NetflixRed,
                            selectedLabelColor = NetflixWhite,
                        ),
                    )
                }
            }

            Spacer(modifier = Modifier.height(24.dp))

            // Translate button
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Button(
                    onClick = {
                        if (selectedSubtitleId >= 0 && selectedTargetLang != null) {
                            onTranslate(selectedSubtitleId, selectedTargetLang!!)
                        }
                    },
                    enabled = selectedSubtitleId >= 0 && selectedTargetLang != null && !isTranslating,
                    colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF2563EB)),
                    modifier = Modifier.weight(1f),
                    shape = RoundedCornerShape(8.dp),
                ) {
                    if (isTranslating) {
                        CircularProgressIndicator(
                            color = NetflixWhite,
                            modifier = Modifier.size(20.dp),
                            strokeWidth = 2.dp,
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                    }
                    Text("Translate")
                }

                TextButton(
                    onClick = onDismiss,
                    colors = ButtonDefaults.textButtonColors(contentColor = NetflixWhite.copy(alpha = 0.7f)),
                ) {
                    Text("Cancel")
                }
            }
        }
    }
}

// Note: VideoPlayer requires ExoPlayer (platform API) and cannot be rendered in IDE preview.
// The following previews cover self-contained sub-components.

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun UpNextCardPreview() {
    VeloxTheme {
        UpNextCard(
            title = "S2E5 - The Battle of Winterfell",
            countdown = 8,
            onPlayNext = {},
            onDismiss = {},
        )
    }
}

private val SamplePlaybackInfo = com.velox.app.domain.model.PlaybackInfo(
    mediaId = 1,
    primaryFileId = 1,
    method = "DirectPlay",
    streamUrl = "http://server:8098/api/stream/1",
    directUrl = "http://server:8098/api/stream/1?pm=direct",
    abrUrl = null,
    pretranscodeUrl = "http://server:8098/api/stream/1/pretranscode",
    hlsUrl = "http://server:8098/api/stream/1/hls/master.m3u8",
    prefer = "direct",
    streamSessionId = "abc123",
    ptVideoCodec = "h264",
    ptAudioCodec = "aac",
    ptHeight = 1080,
    ptVideoBitrate = 8000,
    ptAudioBitrate = 128,
    position = 120f,
    duration = 7200f,
    width = 1920,
    height = 1080,
    audioTracks = emptyList(),
    subtitleTracks = emptyList(),
    skipSegments = emptyList(),
    availableQualities = emptyList(),
)

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun PlaybackStatsOverlayPreview() {
    VeloxTheme {
        PlaybackStatsOverlay(
            visible = true,
            playbackInfo = SamplePlaybackInfo,
            onClose = {},
        )
    }
}
