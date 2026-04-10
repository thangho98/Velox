package com.velox.app.presentation.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Shadow
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

/**
 * DualSubtitleOverlay — renders two subtitle tracks over the video.
 *
 * Primary subtitle: bottom of video, larger text (white by default).
 * Secondary subtitle: above primary, smaller text, yellow (#FFD700).
 *
 * Positioned inside video area (not letterbox) using bottom letterbox calculation.
 */

data class SubtitleCue(
    val start: Long,  // milliseconds
    val end: Long,    // milliseconds
    val text: String,
)

data class SubtitleAppearance(
    val size: String = "large", // "small", "medium", "large"
    val color: String = "#ffffff",
    val background: String = "none", // "solid", "semi", "none"
)

// Font sizes in sp
private val PRIMARY_SIZE_MAP = mapOf(
    "small" to 18.sp,
    "medium" to 24.sp,
    "large" to 32.sp,
)

private val SECONDARY_SIZE_MAP = mapOf(
    "small" to 13.sp,
    "medium" to 16.sp,
    "large" to 20.sp,
)

private val DEFAULT_APPEARANCE = SubtitleAppearance()

@Composable
fun DualSubtitleOverlay(
    primaryCues: List<SubtitleCue>,
    secondaryCues: List<SubtitleCue>,
    currentPosition: Long,
    visible: Boolean,
    letterboxHeightDp: androidx.compose.ui.unit.Dp = 0.dp,
    isVideoReady: Boolean = true,
    appearance: SubtitleAppearance = DEFAULT_APPEARANCE,
    offsetSeconds: Float = 0f,
    modifier: Modifier = Modifier,
) {
    if (!visible) return

    // Apply subtitle delay: positive = delay (show later), negative = show earlier
    val adjustedPosition = currentPosition - (offsetSeconds * 1000).toLong()

    // Find active cues
    val primaryText = primaryCues.find { cue ->
        adjustedPosition >= cue.start && adjustedPosition <= cue.end
    }?.text

    val secondaryText = secondaryCues.find { cue ->
        adjustedPosition >= cue.start && adjustedPosition <= cue.end
    }?.text

    if (primaryText == null && secondaryText == null) return

    // Dynamic font sizing based on breakpoints (similar to Web Tailwind: md=768, lg=1024)
    val configuration = androidx.compose.ui.platform.LocalConfiguration.current
    val screenWidth = configuration.screenWidthDp
    
    val baseFontSizeDp = when {
        screenWidth >= 1024 -> 32.dp // Desktop / Large Tablet Landscape
        screenWidth >= 768 -> 26.dp  // Tablet Portrait
        screenWidth >= 600 -> 22.dp  // Small Tablet / Phablet
        else -> 18.dp                // Phone
    }
    
    val density = androidx.compose.ui.platform.LocalDensity.current
    // Note: To avoid users' system font scaling (Accessibility) from making subs absurdly large, 
    // we use `dp` mapped to `sp` dynamically, or we can just divide by fontScale.
    val fontScale = density.fontScale
    val baseFontSizeSp = with(density) { (baseFontSizeDp / fontScale).toSp() }
    
    val primaryFontSize = when (appearance.size) {
        "small" -> baseFontSizeSp * 0.8f
        "large" -> baseFontSizeSp * 1.3f
        else -> baseFontSizeSp // "medium"
    }
    
    val secondaryFontSize = primaryFontSize * 0.65f

    val stripTags: (String) -> String = { text ->
        text.replace(Regex("<[^>]+>"), "")
    }

    // Strip HTML tags from VTT text
    val CONTROLS_HEIGHT = 120.dp

    // Match web-app logic: if letterbox is large, render inside the black bar.
    // If small, render just above it.
    val bottomPadding = if (!isVideoReady) {
        80.dp // Safe fallback zone before video dimensions are parsed
    } else if (letterboxHeightDp > CONTROLS_HEIGHT) {
        // Large letterbox: position inside the black bar (50% up from the bottom of screen)
        androidx.compose.ui.unit.max(80.dp, letterboxHeightDp * 0.5f)
    } else {
        // Landscape or small letterbox: position closely above video edge
        letterboxHeightDp + 16.dp
    }

    Box(
        modifier = modifier
            .fillMaxWidth()
            .padding(bottom = bottomPadding),
        contentAlignment = Alignment.BottomCenter
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            // Secondary subtitle (above primary, yellow, smaller)
            secondaryText?.let { text ->
                SubtitleText(
                    text = stripTags(text),
                    fontSize = secondaryFontSize,
                    color = Color(0xFFFDE047), // Yellow (matches webapp #fde047)
                    fontWeight = FontWeight.Medium,
                    background = appearance.background,
                )
            }

            // Primary subtitle (bottom, larger, user-selected color)
            primaryText?.let { text ->
                val textColor = try {
                    Color(android.graphics.Color.parseColor(appearance.color))
                } catch (_: Exception) {
                    Color.White
                }
                SubtitleText(
                    text = stripTags(text),
                    fontSize = primaryFontSize,
                    color = textColor,
                    fontWeight = FontWeight.Bold,
                    background = appearance.background,
                )
            }
        }
    }
}

@Composable
private fun SubtitleText(
    text: String,
    fontSize: androidx.compose.ui.unit.TextUnit,
    color: Color,
    fontWeight: FontWeight,
    background: String = "none",
) {
    when (background) {
        "solid" -> {
            // Solid black background (rgba(0,0,0,0.8))
            Box(
                modifier = Modifier
                    .background(Color.Black.copy(alpha = 0.8f), RoundedCornerShape(4.dp))
                    .padding(horizontal = 12.dp, vertical = 4.dp),
            ) {
                Text(
                    text = text,
                    color = color,
                    fontSize = fontSize,
                    lineHeight = fontSize * 1.25f,
                    fontWeight = fontWeight,
                    textAlign = TextAlign.Center,
                    style = TextStyle(
                        shadow = Shadow(color = Color.Black.copy(alpha = 0.6f), offset = Offset(0f, 1f), blurRadius = 2f),
                    ),
                )
            }
        }
        "semi" -> {
            // Semi-transparent background (rgba(0,0,0,0.5))
            Box(
                modifier = Modifier
                    .background(Color.Black.copy(alpha = 0.5f), RoundedCornerShape(4.dp))
                    .padding(horizontal = 12.dp, vertical = 4.dp),
            ) {
                Text(
                    text = text,
                    color = color,
                    fontSize = fontSize,
                    lineHeight = fontSize * 1.25f,
                    fontWeight = fontWeight,
                    textAlign = TextAlign.Center,
                    style = TextStyle(
                        shadow = Shadow(color = Color.Black.copy(alpha = 0.8f), offset = Offset(0f, 1f), blurRadius = 3f),
                    ),
                )
            }
        }
        else -> {
            // "none" — Netflix/Emby style: no background, text stroke + shadow
            // Draw black stroke behind, then colored fill on top
            Box(modifier = Modifier.padding(horizontal = 12.dp, vertical = 4.dp)) {
                // Stroke layer (behind)
                Text(
                    text = text,
                    color = Color.Black,
                    fontSize = fontSize,
                    lineHeight = fontSize * 1.25f,
                    fontWeight = fontWeight,
                    textAlign = TextAlign.Center,
                    style = TextStyle(
                        drawStyle = Stroke(width = 4f),
                        shadow = Shadow(color = Color.Black.copy(alpha = 0.9f), offset = Offset(0f, 0f), blurRadius = 8f),
                    ),
                )
                // Fill layer (on top)
                Text(
                    text = text,
                    color = color,
                    fontSize = fontSize,
                    lineHeight = fontSize * 1.25f,
                    fontWeight = fontWeight,
                    textAlign = TextAlign.Center,
                    style = TextStyle(
                        shadow = Shadow(color = Color.Black.copy(alpha = 0.9f), offset = Offset(1f, 1f), blurRadius = 4f),
                    ),
                )
            }
        }
    }
}

/**
 * Calculate bottom letterbox height from video dimensions.
 * Used to position subtitles inside the video area.
 */
fun calculateLetterboxHeight(
    videoWidth: Int,
    videoHeight: Int,
    containerWidth: Int,
    containerHeight: Int,
): Int {
    if (containerHeight <= 0 || videoHeight <= 0 || videoWidth <= 0 || containerWidth <= 0) return 0
    val videoAspect = videoWidth.toFloat() / videoHeight.toFloat()
    val containerAspect = containerWidth.toFloat() / containerHeight.toFloat()
    
    // If video is wider than container, it will be letterboxed (black bars top/bottom)
    if (videoAspect > containerAspect) {
        val displayHeight = containerWidth / videoAspect
        return ((containerHeight - displayHeight) / 2f).toInt()
    }
    return 0
}

private val SamplePrimaryCues = listOf(
    SubtitleCue(start = 0L, end = 5000L, text = "This is the primary subtitle text."),
    SubtitleCue(start = 5000L, end = 10000L, text = "This scene is very dramatic."),
)

private val SampleSecondaryCues = listOf(
    SubtitleCue(start = 0L, end = 5000L, text = "Đây là dòng phụ đề phụ."),
    SubtitleCue(start = 5000L, end = 10000L, text = "Cảnh này rất kịch tính."),
)

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun DualSubtitleOverlayPreview() {
    VeloxTheme {
        DualSubtitleOverlay(
            primaryCues = SamplePrimaryCues,
            secondaryCues = SampleSecondaryCues,
            currentPosition = 2500L,
            visible = true,
        )
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun DualSubtitlePrimaryOnlyPreview() {
    VeloxTheme {
        DualSubtitleOverlay(
            primaryCues = SamplePrimaryCues,
            secondaryCues = emptyList(),
            currentPosition = 2500L,
            visible = true,
        )
    }
}
