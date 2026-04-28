package com.velox.app.presentation.ui.components.livetv

import android.net.Uri
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
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
import com.velox.app.domain.model.livetv.LiveChannel
import com.velox.app.domain.model.livetv.LiveProgram
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter

private val timeFormatter = DateTimeFormatter.ofPattern("HH:mm").withZone(ZoneId.systemDefault())

private fun formatTime(iso: String): String = runCatching {
    timeFormatter.format(Instant.parse(iso))
}.getOrDefault("")

@androidx.annotation.OptIn(androidx.media3.common.util.UnstableApi::class)
@Composable
fun LiveTvMiniPlayer(
    channel: LiveChannel?,
    currentProgram: LiveProgram?,
    streamUrl: String,
    onWatch: (LiveChannel, Boolean) -> Unit,
    modifier: Modifier = Modifier,
) {
    if (channel == null || streamUrl.isEmpty()) {
        Box(
            modifier = modifier.background(NetflixDark),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                text = "Select a channel to watch",
                color = NetflixLightGray,
                fontSize = 14.sp,
                fontWeight = FontWeight.Medium,
            )
        }
        return
    }

    val context = LocalContext.current
    val lifecycleOwner = LocalLifecycleOwner.current

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
                volume = 0f
            }
    }

    var isBuffering by remember(channel.id) { mutableStateOf(true) }
    var errorMessage by remember(channel.id) { mutableStateOf<String?>(null) }

    DisposableEffect(exoPlayer) {
        val listener = object : Player.Listener {
            override fun onPlaybackStateChanged(state: Int) {
                isBuffering = state == Player.STATE_BUFFERING
                if (state == Player.STATE_READY) {
                    errorMessage = null
                }
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

    DisposableEffect(lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> exoPlayer.pause()
                Lifecycle.Event.ON_RESUME -> exoPlayer.play()
                else -> Unit
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose { lifecycleOwner.lifecycle.removeObserver(observer) }
    }

    val metaParts = listOfNotNull(
        channel.country.takeIf { it.isNotBlank() },
        channel.groupTitle.takeIf { it.isNotBlank() },
    )

    Box(modifier = modifier.background(Color.Black)) {
        AndroidView(
            factory = { ctx ->
                PlayerView(ctx).apply {
                    useController = false
                    resizeMode = AspectRatioFrameLayout.RESIZE_MODE_ZOOM
                    isClickable = false
                    isFocusable = false
                    isFocusableInTouchMode = false
                }
            },
            update = { view -> view.player = exoPlayer },
            modifier = Modifier.fillMaxSize(),
        )

        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    Brush.verticalGradient(
                        0f to Color.Black.copy(alpha = 0.55f),
                        0.35f to Color.Transparent,
                        0.7f to Color.Transparent,
                        1f to Color.Black.copy(alpha = 0.9f),
                    ),
                ),
        )

        Row(
            modifier = Modifier
                .align(Alignment.TopStart)
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Row(
                modifier = Modifier
                    .clip(RoundedCornerShape(2.dp))
                    .background(NetflixRed)
                    .padding(horizontal = 8.dp, vertical = 3.dp),
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

        Column(
            modifier = Modifier
                .align(Alignment.BottomStart)
                .fillMaxWidth()
                .padding(horizontal = 20.dp, vertical = 20.dp),
        ) {
            Text(
                text = "NOW WATCHING",
                color = NetflixRed,
                fontSize = 10.sp,
                fontWeight = FontWeight.Bold,
                letterSpacing = 3.sp,
            )
            Spacer(modifier = Modifier.height(6.dp))
            Text(
                text = channel.name,
                color = NetflixWhite,
                fontSize = 30.sp,
                fontWeight = FontWeight.ExtraBold,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
            if (metaParts.isNotEmpty()) {
                Spacer(modifier = Modifier.height(6.dp))
                Text(
                    text = metaParts.joinToString(" · ").uppercase(),
                    color = NetflixLightGray,
                    fontSize = 11.sp,
                    fontWeight = FontWeight.SemiBold,
                    letterSpacing = 2.sp,
                )
            }
            if (currentProgram != null) {
                Spacer(modifier = Modifier.height(8.dp))
                val range = "${formatTime(currentProgram.startTimeIso)} – ${formatTime(currentProgram.endTimeIso)}"
                Text(
                    text = "${currentProgram.title}  ·  $range",
                    color = NetflixLightGray,
                    fontSize = 13.sp,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }

            Spacer(modifier = Modifier.height(14.dp))
            Row(verticalAlignment = Alignment.CenterVertically) {
                Button(
                    onClick = { onWatch(channel, false) },
                    shape = RoundedCornerShape(2.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = NetflixWhite,
                        contentColor = NetflixBlack,
                    ),
                ) {
                    Icon(
                        LucideIcons.PlayArrow,
                        contentDescription = null,
                        modifier = Modifier.size(18.dp),
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "WATCH",
                        fontSize = 13.sp,
                        fontWeight = FontWeight.Bold,
                        letterSpacing = 1.sp,
                    )
                }
                Spacer(modifier = Modifier.width(10.dp))
                Box(
                    modifier = Modifier
                        .size(44.dp)
                        .clip(RoundedCornerShape(2.dp))
                        .background(NetflixWhite.copy(alpha = 0.15f))
                        .clickable { onWatch(channel, true) },
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(
                        LucideIcons.Fullscreen,
                        contentDescription = "Fullscreen",
                        tint = NetflixWhite,
                        modifier = Modifier.size(18.dp),
                    )
                }
            }
        }

        if (isBuffering && errorMessage == null) {
            CircularProgressIndicator(
                color = NetflixWhite,
                strokeWidth = 3.dp,
                modifier = Modifier
                    .align(Alignment.Center)
                    .size(44.dp),
            )
        }

        if (errorMessage != null) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(Color.Black.copy(alpha = 0.75f)),
                contentAlignment = Alignment.Center,
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(
                        text = "Stream Unavailable",
                        color = NetflixWhite,
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Bold,
                    )
                    Spacer(modifier = Modifier.height(6.dp))
                    Text(
                        text = errorMessage!!,
                        color = NetflixLightGray,
                        fontSize = 13.sp,
                    )
                }
            }
        }
    }
}
