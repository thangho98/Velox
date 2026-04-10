package com.velox.app.presentation.ui.components

import androidx.compose.foundation.background
import com.velox.app.presentation.ui.components.LucideIcons
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.VolumeOff
import androidx.compose.material.icons.automirrored.filled.VolumeUp
import androidx.compose.material.icons.filled.BrightnessHigh
import androidx.compose.material.icons.filled.BrightnessLow
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.ui.theme.VeloxTheme

/**
 * GestureFeedbackOverlay — shows volume/brightness feedback during pan gestures.
 *
 * Left side pan: brightness feedback (visual only - no actual control)
 * Right side pan: volume feedback (actually adjusts volume)
 */
@Composable
fun GestureFeedbackOverlay(
    volume: Float?,      // null = no feedback, 0-1 = volume level
    brightness: Float?, // null = no feedback, -1 to 1 = brightness delta
    modifier: Modifier = Modifier,
) {
    Box(modifier = modifier.fillMaxSize()) {
        // Volume feedback (right side)
        volume?.let { vol ->
            Box(
                modifier = Modifier
                    .align(Alignment.CenterEnd)
                    .padding(end = 80.dp)
                    .background(
                        color = Color.Black.copy(alpha = 0.7f),
                        shape = RoundedCornerShape(8.dp),
                    )
                    .padding(horizontal = 20.dp, vertical = 12.dp),
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    Icon(
                        imageVector = if (vol > 0.5f) LucideIcons.VolumeUp else LucideIcons.VolumeOff,
                        contentDescription = null,
                        tint = Color.White,
                        modifier = Modifier.size(20.dp),
                    )
                    Text(
                        text = "${(vol * 100).toInt()}%",
                        color = Color.White,
                        fontSize = 20.sp,
                        fontWeight = FontWeight.SemiBold,
                    )
                }
            }
        }

        // Brightness feedback (left side)
        brightness?.let { bright ->
            Box(
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .padding(start = 80.dp)
                    .background(
                        color = Color.Black.copy(alpha = 0.7f),
                        shape = RoundedCornerShape(8.dp),
                    )
                    .padding(horizontal = 20.dp, vertical = 12.dp),
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    Icon(
                        imageVector = if (bright >= 0) LucideIcons.BrightnessHigh else LucideIcons.BrightnessLow,
                        contentDescription = null,
                        tint = Color.White,
                        modifier = Modifier.size(20.dp),
                    )
                    Text(
                        text = "${if (bright >= 0) "+" else ""}${(bright * 100).toInt()}%",
                        color = Color.White,
                        fontSize = 20.sp,
                        fontWeight = FontWeight.SemiBold,
                    )
                }
            }
        }
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun GestureFeedbackVolumePreview() {
    VeloxTheme {
        GestureFeedbackOverlay(
            volume = 0.75f,
            brightness = null,
        )
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun GestureFeedbackBrightnessPreview() {
    VeloxTheme {
        GestureFeedbackOverlay(
            volume = null,
            brightness = 0.3f,
        )
    }
}
