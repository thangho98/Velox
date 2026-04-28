package com.velox.app.presentation.ui.screens.livetv

import android.app.Activity
import android.app.PictureInPictureParams
import android.content.pm.ActivityInfo
import android.content.pm.PackageManager
import android.net.Uri
import android.os.Build
import android.util.Rational
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectDragGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.systemBarsPadding
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.media3.common.MediaItem
import androidx.media3.common.MimeTypes
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.DefaultLoadControl
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.hls.HlsMediaSource
import androidx.media3.ui.AspectRatioFrameLayout
import androidx.media3.ui.PlayerView
import com.velox.app.domain.model.livetv.LiveProgram
import com.velox.app.presentation.ui.components.GestureFeedbackOverlay
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.ui.components.livetv.LivePulseDot
import com.velox.app.presentation.viewmodel.livetv.LiveTvPlayerViewModel
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import kotlinx.coroutines.delay
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter

private const val IDLE_HIDE_MS = 3000L
private val timeFormatter = DateTimeFormatter.ofPattern("HH:mm").withZone(ZoneId.systemDefault())

@Composable
fun LiveTvPlayerScreen(
    channelId: Int,
    viewModel: LiveTvPlayerViewModel = hiltViewModel(),
    onBackClick: () -> Unit,
) {
    val context = LocalContext.current
    val baseUrl = viewModel.baseUrl
    val channel by viewModel.channel.collectAsStateWithLifecycle()
    val epg by viewModel.epg.collectAsStateWithLifecycle()

    LaunchedEffect(channelId) { viewModel.load(channelId) }

    val currentProgram = remember(epg) { findCurrentProgram(epg) }

    BackHandler(onBack = onBackClick)

    if (baseUrl.isEmpty()) {
        Box(modifier = Modifier.fillMaxSize().background(Color.Black), contentAlignment = Alignment.Center) {
            CircularProgressIndicator(color = NetflixWhite)
        }
        return
    }

    val streamUrl = remember(channelId) { viewModel.streamUrl(channelId) }

    val exoPlayer = remember(streamUrl) {
        val liveConfig = MediaItem.LiveConfiguration.Builder()
            .setTargetOffsetMs(12_000)
            .setMinPlaybackSpeed(0.95f)
            .setMaxPlaybackSpeed(1.05f)
            .build()
        val mediaItem = MediaItem.Builder()
            .setUri(Uri.parse(streamUrl))
            .setMimeType(MimeTypes.APPLICATION_M3U8)
            .setLiveConfiguration(liveConfig)
            .build()
        val httpFactory = DefaultHttpDataSource.Factory()
            .setConnectTimeoutMs(15_000)
            .setReadTimeoutMs(15_000)
            .setAllowCrossProtocolRedirects(true)
        val hlsSource = HlsMediaSource.Factory(httpFactory).createMediaSource(mediaItem)
        val loadControl = DefaultLoadControl.Builder()
            .setBufferDurationsMs(15_000, 60_000, 2_500, 5_000)
            .setPrioritizeTimeOverSizeThresholds(true)
            .build()
        ExoPlayer.Builder(context)
            .setLoadControl(loadControl)
            .build()
            .apply {
                setMediaSource(hlsSource)
                prepare()
                playWhenReady = true
            }
    }

    var isBuffering by remember(channelId) { mutableStateOf(true) }
    var isPaused by remember(channelId) { mutableStateOf(false) }
    var volume by remember { mutableFloatStateOf(1f) }
    var volumeBeforeMute by remember { mutableFloatStateOf(1f) }
    val isMuted = volume == 0f
    var errorMessage by remember(channelId) { mutableStateOf<String?>(null) }
    var showControls by remember { mutableStateOf(true) }
    var interactionKey by remember { mutableStateOf(0) }
    var isFullscreen by remember { mutableStateOf(false) }

    var showVolumeFeedback by remember { mutableStateOf(false) }
    var showBrightnessFeedback by remember { mutableStateOf(false) }
    var panBrightness by remember { mutableFloatStateOf(0f) }

    val activity = remember(context) {
        var ctx = context
        while (ctx is android.content.ContextWrapper && ctx !is Activity) ctx = ctx.baseContext
        ctx as? Activity
    }

    val supportsPiP = remember(activity) {
        Build.VERSION.SDK_INT >= Build.VERSION_CODES.O &&
            activity?.packageManager?.hasSystemFeature(PackageManager.FEATURE_PICTURE_IN_PICTURE) == true
    }

    LaunchedEffect(isFullscreen) {
        val act = activity ?: return@LaunchedEffect
        if (isFullscreen) {
            act.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_SENSOR_LANDSCAPE
        } else {
            act.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_PORTRAIT
            delay(500)
            act.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_UNSPECIFIED
        }
    }

    DisposableEffect(Unit) {
        onDispose { activity?.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_UNSPECIFIED }
    }

    DisposableEffect(exoPlayer) {
        val listener = object : Player.Listener {
            override fun onPlaybackStateChanged(state: Int) {
                isBuffering = state == Player.STATE_BUFFERING
                if (state == Player.STATE_READY) errorMessage = null
            }

            override fun onIsPlayingChanged(playing: Boolean) {
                isPaused = !playing
            }

            override fun onPlayerError(error: PlaybackException) {
                errorMessage = "Stream unavailable"
                isBuffering = false
            }
        }
        exoPlayer.addListener(listener)
        onDispose {
            exoPlayer.removeListener(listener)
            exoPlayer.release()
        }
    }

    LaunchedEffect(interactionKey, showControls) {
        if (showControls) {
            delay(IDLE_HIDE_MS)
            showControls = false
        }
    }

    fun bumpInteraction() {
        showControls = true
        interactionKey++
    }

    fun togglePlay() {
        if (exoPlayer.isPlaying) exoPlayer.pause() else exoPlayer.play()
        bumpInteraction()
    }

    fun toggleMute() {
        volume = if (isMuted) {
            volumeBeforeMute.takeIf { it > 0f } ?: 1f
        } else {
            volumeBeforeMute = volume
            0f
        }
        exoPlayer.volume = volume
        bumpInteraction()
    }

    fun toggleFullscreen() {
        isFullscreen = !isFullscreen
        bumpInteraction()
    }

    fun enterPip() {
        if (!supportsPiP || Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return
        activity?.let {
            val params = PictureInPictureParams.Builder()
                .setAspectRatio(Rational(16, 9))
                .build()
            runCatching { it.enterPictureInPictureMode(params) }
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black)
            .clickable(
                interactionSource = remember { androidx.compose.foundation.interaction.MutableInteractionSource() },
                indication = null,
            ) {
                showControls = !showControls
                if (showControls) interactionKey++
            }
            .pointerInput(Unit) {
                detectDragGestures(
                    onDrag = { change, dragAmount ->
                        if (kotlin.math.abs(dragAmount.y) <= kotlin.math.abs(dragAmount.x)) return@detectDragGestures
                        val delta = -dragAmount.y / 200f
                        val screenWidth = size.width.toFloat()
                        if (change.position.x < screenWidth / 2) {
                            panBrightness = (panBrightness + delta).coerceIn(-1f, 1f)
                            showBrightnessFeedback = true
                            showVolumeFeedback = false
                        } else {
                            val newVolume = (volume + delta).coerceIn(0f, 1f)
                            volume = newVolume
                            exoPlayer.volume = newVolume
                            if (newVolume > 0f) volumeBeforeMute = newVolume
                            showVolumeFeedback = true
                            showBrightnessFeedback = false
                        }
                    },
                    onDragEnd = {
                        showVolumeFeedback = false
                        showBrightnessFeedback = false
                        panBrightness = 0f
                    },
                )
            },
    ) {
        AndroidView(
            factory = { ctx ->
                PlayerView(ctx).apply {
                    useController = false
                    resizeMode = AspectRatioFrameLayout.RESIZE_MODE_FIT
                    isClickable = false
                    isFocusable = false
                    isFocusableInTouchMode = false
                }
            },
            update = { view ->
                view.player = exoPlayer
                view.keepScreenOn = !isPaused
            },
            modifier = Modifier.fillMaxSize(),
        )

        if (showControls) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(
                        Brush.verticalGradient(
                            0f to Color.Black.copy(alpha = 0.55f),
                            0.22f to Color.Transparent,
                            0.75f to Color.Transparent,
                            1f to Color.Black.copy(alpha = 0.85f),
                        ),
                    ),
            )
        }

        if (errorMessage != null) {
            ErrorOverlay(message = errorMessage!!, onBack = onBackClick)
        } else if (isBuffering) {
            CircularProgressIndicator(
                color = NetflixWhite,
                strokeWidth = 3.dp,
                modifier = Modifier.align(Alignment.Center).size(56.dp),
            )
        } else if (isPaused) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(Color.Black.copy(alpha = 0.3f))
                    .clickable { togglePlay() },
                contentAlignment = Alignment.Center,
            ) {
                Box(
                    modifier = Modifier
                        .size(80.dp)
                        .clip(CircleShape)
                        .background(NetflixWhite),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(
                        LucideIcons.PlayArrow,
                        contentDescription = "Play",
                        tint = Color.Black,
                        modifier = Modifier.size(40.dp),
                    )
                }
            }
        }

        GestureFeedbackOverlay(
            volume = if (showVolumeFeedback) volume else null,
            brightness = if (showBrightnessFeedback) panBrightness else null,
        )

        if (showControls && errorMessage == null) {
            TopBar(
                onBack = {
                    bumpInteraction()
                    onBackClick()
                },
                modifier = Modifier.align(Alignment.TopCenter),
            )
            BottomBar(
                channelName = channel?.name ?: "Live channel",
                metaText = channel?.let { ch ->
                    listOfNotNull(
                        ch.country.takeIf { it.isNotBlank() },
                        ch.groupTitle.takeIf { it.isNotBlank() },
                    ).joinToString(" · ").uppercase().ifEmpty { null }
                },
                currentProgram = currentProgram,
                isPaused = isPaused,
                isMuted = isMuted,
                isFullscreen = isFullscreen,
                showPip = supportsPiP,
                onTogglePlay = ::togglePlay,
                onToggleMute = ::toggleMute,
                onToggleFullscreen = ::toggleFullscreen,
                onEnterPip = ::enterPip,
                modifier = Modifier.align(Alignment.BottomCenter),
            )
        }
    }
}

@Composable
private fun TopBar(
    onBack: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .systemBarsPadding()
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Box(
            modifier = Modifier
                .size(40.dp)
                .clip(CircleShape)
                .clickable(onClick = onBack),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                LucideIcons.ArrowBack,
                contentDescription = "Back",
                tint = NetflixWhite,
                modifier = Modifier.size(28.dp),
            )
        }
        Row(
            modifier = Modifier
                .clip(RoundedCornerShape(2.dp))
                .background(NetflixRed)
                .padding(horizontal = 8.dp, vertical = 4.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            LivePulseDot(color = NetflixWhite)
            Spacer(modifier = Modifier.width(6.dp))
            Text(
                text = "LIVE",
                color = NetflixWhite,
                fontSize = 10.sp,
                fontWeight = FontWeight.Bold,
                letterSpacing = 2.sp,
            )
        }
    }
}

@Composable
private fun BottomBar(
    channelName: String,
    metaText: String?,
    currentProgram: LiveProgram?,
    isPaused: Boolean,
    isMuted: Boolean,
    isFullscreen: Boolean,
    showPip: Boolean,
    onTogglePlay: () -> Unit,
    onToggleMute: () -> Unit,
    onToggleFullscreen: () -> Unit,
    onEnterPip: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .systemBarsPadding()
            .padding(horizontal = 16.dp, vertical = 16.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            modifier = Modifier
                .size(56.dp)
                .clip(CircleShape)
                .background(NetflixWhite)
                .clickable(onClick = onTogglePlay),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                if (isPaused) LucideIcons.PlayArrow else LucideIcons.Pause,
                contentDescription = if (isPaused) "Play" else "Pause",
                tint = Color.Black,
                modifier = Modifier.size(26.dp),
            )
        }
        Spacer(modifier = Modifier.width(14.dp))
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = channelName,
                color = NetflixWhite,
                fontSize = 18.sp,
                fontWeight = FontWeight.Bold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            val subtitleParts = buildList {
                currentProgram?.let {
                    val range = "${formatTime(it.startTimeIso)} – ${formatTime(it.endTimeIso)}"
                    add("${it.title} · $range")
                }
                metaText?.let { add(it) }
            }
            if (subtitleParts.isNotEmpty()) {
                Spacer(modifier = Modifier.height(2.dp))
                Text(
                    text = subtitleParts.joinToString(" • "),
                    color = NetflixLightGray,
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Medium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
        Spacer(modifier = Modifier.width(10.dp))
        IconSquare(
            icon = if (isMuted) LucideIcons.VolumeOff else LucideIcons.VolumeUp,
            description = if (isMuted) "Unmute" else "Mute",
            onClick = onToggleMute,
        )
        if (showPip) {
            Spacer(modifier = Modifier.width(8.dp))
            IconSquare(
                icon = LucideIcons.PictureInPicture,
                description = "Picture in picture",
                onClick = onEnterPip,
            )
        }
        Spacer(modifier = Modifier.width(8.dp))
        IconSquare(
            icon = if (isFullscreen) LucideIcons.FullscreenExit else LucideIcons.Fullscreen,
            description = if (isFullscreen) "Exit fullscreen" else "Fullscreen",
            onClick = onToggleFullscreen,
        )
    }
}

@Composable
private fun IconSquare(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    description: String,
    onClick: () -> Unit,
) {
    Box(
        modifier = Modifier
            .size(40.dp)
            .clip(RoundedCornerShape(4.dp))
            .background(NetflixWhite.copy(alpha = 0.12f))
            .clickable(onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        Icon(
            icon,
            contentDescription = description,
            tint = NetflixWhite,
            modifier = Modifier.size(20.dp),
        )
    }
}

@Composable
private fun ErrorOverlay(message: String, onBack: () -> Unit) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black.copy(alpha = 0.82f)),
        contentAlignment = Alignment.Center,
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text(
                text = "Stream Unavailable",
                color = NetflixWhite,
                fontSize = 20.sp,
                fontWeight = FontWeight.Bold,
            )
            Spacer(modifier = Modifier.height(6.dp))
            Text(
                text = message,
                color = NetflixLightGray,
                fontSize = 14.sp,
            )
            Spacer(modifier = Modifier.height(18.dp))
            Row(
                modifier = Modifier
                    .clip(RoundedCornerShape(2.dp))
                    .background(NetflixWhite)
                    .clickable(onClick = onBack)
                    .padding(horizontal = 16.dp, vertical = 10.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Icon(
                    LucideIcons.ArrowBack,
                    contentDescription = null,
                    tint = Color.Black,
                    modifier = Modifier.size(16.dp),
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "BACK TO CHANNELS",
                    color = Color.Black,
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Bold,
                    letterSpacing = 1.sp,
                )
            }
        }
    }
}

private fun formatTime(iso: String): String = runCatching {
    timeFormatter.format(Instant.parse(iso))
}.getOrDefault("")

private fun findCurrentProgram(programs: List<LiveProgram>): LiveProgram? {
    if (programs.isEmpty()) return null
    val now = System.currentTimeMillis()
    return programs.firstOrNull { p ->
        val s = runCatching { Instant.parse(p.startTimeIso).toEpochMilli() }.getOrDefault(0L)
        val e = runCatching { Instant.parse(p.endTimeIso).toEpochMilli() }.getOrDefault(0L)
        s <= now && now < e
    }
}
