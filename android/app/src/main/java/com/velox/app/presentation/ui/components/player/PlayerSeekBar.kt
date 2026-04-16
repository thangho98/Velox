package com.velox.app.presentation.ui.components.player

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.velox.app.ui.theme.NetflixWhite

@OptIn(androidx.compose.material3.ExperimentalMaterial3Api::class)
@Composable
internal fun PlayerSeekBar(
    currentPosition: Long,
    duration: Long,
    onSeek: (Long) -> Unit,
) {
    var sliderPosition by remember { mutableFloatStateOf(0f) }
    var isDragging by remember { mutableStateOf(false) }

    LaunchedEffect(currentPosition, duration) {
        if (!isDragging) {
            if (duration > 0) {
                sliderPosition = (currentPosition.toFloat() / duration.toFloat()).coerceIn(0f, 1f)
            } else {
                sliderPosition = 0f
            }
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
                modifier = Modifier.height(4.dp),
                drawStopIndicator = null,
            )
        },
        modifier = Modifier.fillMaxWidth(),
    )
}

internal fun formatTime(ms: Long): String {
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
