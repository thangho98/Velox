package com.velox.app.presentation.ui.components

import androidx.compose.foundation.background
import com.velox.app.presentation.ui.components.LucideIcons
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Movie
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.SkipNext
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

/**
 * CinemaOverlay — Pre-roll cinema experience before main video playback
 *
 * Shows intro/trailer cards that can be skipped.
 * User can skip individual items or skip to main movie at any time.
 */

data class CinemaItem(
    val type: String, // "intro" or "trailer"
    val title: String,
    val url: String,
    val skippable: Boolean = true,
)

@Composable
fun CinemaOverlay(
    items: List<CinemaItem>,
    onComplete: () -> Unit,
    onItemEnded: () -> Unit,
    modifier: Modifier = Modifier,
) {
    if (items.isEmpty()) {
        onComplete()
        return
    }

    var currentIndex by remember { mutableStateOf(0) }
    val currentItem = items.getOrNull(currentIndex)

    if (currentItem == null) {
        onComplete()
        return
    }

    val isIntro = currentItem.type == "intro"
    val trailerIndex = items.slice(0..currentIndex).count { it.type == "trailer" }
    val totalTrailers = items.count { it.type == "trailer" }

    Box(
        modifier = modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        // Media content
        when (currentItem.type) {
            "intro" -> {
                // Intro video placeholder - would integrate with video player
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator(color = NetflixRed)
                }
            }
            "trailer" -> {
                // Trailer info card
                TrailerCard(
                    item = currentItem,
                    onWatchTrailer = {
                        // Would open YouTube or play trailer
                        onItemEnded()
                    },
                )
            }
        }

        // Top bar with gradient
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .align(Alignment.TopCenter)
                .background(
                    Brush.verticalGradient(
                        colors = listOf(
                            Color.Black.copy(alpha = 0.7f),
                            Color.Transparent,
                        ),
                    ),
                )
                .padding(top = 48.dp, start = 20.dp, end = 20.dp, bottom = 20.dp),
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Icon(
                    LucideIcons.Movie,
                    contentDescription = null,
                    tint = Color(0xFFFFC107),
                    modifier = Modifier.size(18.dp),
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = if (isIntro) "Cinema Intro" else "Trailer $trailerIndex of $totalTrailers",
                    color = NetflixWhite.copy(alpha = 0.85f),
                    fontSize = 14.sp,
                    fontWeight = FontWeight.SemiBold,
                )
                Spacer(modifier = Modifier.weight(1f))
                Text(
                    text = "${currentIndex + 1} / ${items.size}",
                    color = NetflixWhite.copy(alpha = 0.4f),
                    fontSize = 12.sp,
                )
            }
        }

        // Bottom bar with skip controls
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .align(Alignment.BottomCenter)
                .background(
                    Brush.verticalGradient(
                        colors = listOf(
                            Color.Transparent,
                            Color.Black.copy(alpha = 0.7f),
                        ),
                    ),
                )
                .padding(horizontal = 20.dp, vertical = 32.dp),
        ) {
            Column {
                Text(
                    text = currentItem.title,
                    color = NetflixWhite.copy(alpha = 0.6f),
                    fontSize = 13.sp,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.padding(bottom = 12.dp),
                )

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.End,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    // Skip button
                    if (currentItem.skippable) {
                        Button(
                            onClick = {
                                val nextIndex = currentIndex + 1
                                if (nextIndex >= items.size) {
                                    onComplete()
                                } else {
                                    currentIndex = nextIndex
                                }
                            },
                            colors = ButtonDefaults.buttonColors(
                                containerColor = NetflixWhite.copy(alpha = 0.2f),
                            ),
                            shape = RoundedCornerShape(8.dp),
                            contentPadding = PaddingValues(horizontal = 16.dp, vertical = 10.dp),
                        ) {
                            Icon(
                                LucideIcons.SkipNext,
                                contentDescription = null,
                                modifier = Modifier.size(16.dp),
                            )
                            Spacer(modifier = Modifier.width(6.dp))
                            Text("Skip", fontSize = 14.sp)
                        }

                        Spacer(modifier = Modifier.width(12.dp))
                    }

                    // Skip all button
                    OutlinedButton(
                        onClick = onComplete,
                        colors = ButtonDefaults.outlinedButtonColors(
                            contentColor = NetflixWhite.copy(alpha = 0.7f),
                        ),
                        border = ButtonDefaults.outlinedButtonBorder.copy(
                            brush = Brush.linearGradient(
                                colors = listOf(
                                    NetflixWhite.copy(alpha = 0.3f),
                                    NetflixWhite.copy(alpha = 0.3f),
                                ),
                            ),
                        ),
                        shape = RoundedCornerShape(8.dp),
                        contentPadding = PaddingValues(horizontal = 16.dp, vertical = 10.dp),
                    ) {
                        Text("Skip to Movie", fontSize = 14.sp)
                    }
                }
            }
        }
    }
}

@Composable
private fun TrailerCard(
    item: CinemaItem,
    onWatchTrailer: () -> Unit,
) {
    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.Center,
    ) {
        Card(
            modifier = Modifier
                .padding(horizontal = 40.dp),
            colors = CardDefaults.cardColors(
                containerColor = NetflixBlack,
            ),
            shape = RoundedCornerShape(16.dp),
        ) {
            Column(
                modifier = Modifier.padding(32.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Icon(
                    LucideIcons.Movie,
                    contentDescription = null,
                    tint = Color(0xFFFFC107),
                    modifier = Modifier.size(40.dp),
                )

                Spacer(modifier = Modifier.height(16.dp))

                Text(
                    text = item.title,
                    color = NetflixWhite,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Bold,
                    textAlign = TextAlign.Center,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )

                Text(
                    text = "YouTube Trailer",
                    color = NetflixWhite.copy(alpha = 0.5f),
                    fontSize = 13.sp,
                    modifier = Modifier.padding(top = 6.dp, bottom = 20.dp),
                )

                Button(
                    onClick = onWatchTrailer,
                    colors = ButtonDefaults.buttonColors(
                        containerColor = NetflixRed,
                    ),
                    shape = RoundedCornerShape(8.dp),
                    contentPadding = PaddingValues(horizontal = 24.dp, vertical = 12.dp),
                ) {
                    Icon(
                        LucideIcons.PlayArrow,
                        contentDescription = null,
                        modifier = Modifier.size(20.dp),
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "Watch in YouTube",
                        fontSize = 15.sp,
                        fontWeight = FontWeight.SemiBold,
                    )
                }
            }
        }
    }
}

private val SampleCinemaItems = listOf(
    CinemaItem(type = "intro", title = "Velox Cinema Intro", url = "", skippable = true),
    CinemaItem(type = "trailer", title = "The Dark Knight - Official Trailer", url = "https://youtube.com/embed/EXeTwQW6GQg", skippable = true),
    CinemaItem(type = "trailer", title = "Inception - Final Trailer", url = "https://youtube.com/embed/YoHD9XEInc0", skippable = true),
)

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun CinemaOverlayPreview() {
    VeloxTheme {
        CinemaOverlay(
            items = SampleCinemaItems,
            onComplete = {},
            onItemEnded = {},
        )
    }
}
