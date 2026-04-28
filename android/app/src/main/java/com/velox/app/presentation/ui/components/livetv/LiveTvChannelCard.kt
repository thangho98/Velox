package com.velox.app.presentation.ui.components.livetv

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil.compose.AsyncImage
import coil.request.ImageRequest
import com.velox.app.domain.model.livetv.LiveChannel
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite

private val cardColors = listOf(
    Color(0xFF12234D),
    Color(0xFF3D141A),
    Color(0xFF2C1C08),
    Color(0xFF083023),
    Color(0xFF112933),
    Color(0xFF1F1540),
    Color(0xFF33152A),
    Color(0xFF1A1A1A),
)

private fun seedColor(channel: LiveChannel): Color {
    val seed = "${channel.name}:${channel.groupTitle}"
    var hash = 0
    for (ch in seed) {
        hash = (hash * 31 + ch.code) % 997
    }
    return cardColors[kotlin.math.abs(hash) % cardColors.size]
}

private val decorationRegex = Regex("\\s*[(\\[][^)\\]]*[)\\]]")

private fun stripDecorations(name: String): String =
    name.replace(decorationRegex, "").trim().ifEmpty { name }

@Composable
fun LiveTvChannelCard(
    channel: LiveChannel,
    isActive: Boolean,
    onClick: () -> Unit,
    onToggleHidden: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val bg = remember(channel) { seedColor(channel) }
    val cleanName = remember(channel.name) { stripDecorations(channel.name) }
    var imageFailed by remember(channel.logo) { mutableStateOf(false) }
    var menuOpen by remember { mutableStateOf(false) }

    val activeBorder = if (isActive) {
        Modifier.border(BorderStroke(2.dp, NetflixRed), RoundedCornerShape(4.dp))
    } else {
        Modifier
    }

    Column(
        modifier = modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(4.dp))
            .then(activeBorder)
            .clickable(onClick = onClick),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .aspectRatio(4f / 3f)
                .background(bg),
            contentAlignment = Alignment.Center,
        ) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(
                        Brush.verticalGradient(
                            0f to Color.White.copy(alpha = 0.06f),
                            1f to Color.Black.copy(alpha = 0.28f),
                        ),
                    ),
            )

            if (!channel.logo.isNullOrBlank() && !imageFailed) {
                AsyncImage(
                    model = ImageRequest.Builder(androidx.compose.ui.platform.LocalContext.current)
                        .data(channel.logo)
                        .crossfade(true)
                        .build(),
                    contentDescription = channel.name,
                    contentScale = ContentScale.Fit,
                    onError = { imageFailed = true },
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(horizontal = 22.dp, vertical = 18.dp),
                )
            } else {
                Text(
                    text = cleanName,
                    color = NetflixWhite.copy(alpha = 0.92f),
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Black,
                    textAlign = TextAlign.Center,
                    maxLines = 3,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.padding(horizontal = 8.dp),
                )
            }

            Row(
                modifier = Modifier
                    .align(Alignment.TopStart)
                    .padding(8.dp)
                    .clip(RoundedCornerShape(2.dp))
                    .background(NetflixBlack.copy(alpha = 0.7f))
                    .padding(horizontal = 6.dp, vertical = 3.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                LivePulseDot(color = NetflixRed, size = 5.dp)
                Spacer(modifier = Modifier.width(4.dp))
                Text(
                    text = "LIVE",
                    color = NetflixWhite,
                    fontSize = 9.sp,
                    fontWeight = FontWeight.Bold,
                    letterSpacing = 2.sp,
                )
            }

            Box(modifier = Modifier.align(Alignment.TopEnd)) {
                IconButton(
                    onClick = { menuOpen = true },
                    modifier = Modifier.size(32.dp),
                ) {
                    Box(
                        modifier = Modifier
                            .size(26.dp)
                            .clip(CircleShape)
                            .background(NetflixBlack.copy(alpha = 0.55f)),
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(
                            LucideIcons.MoreVert,
                            contentDescription = "Options",
                            tint = NetflixWhite,
                            modifier = Modifier.size(16.dp),
                        )
                    }
                }
                DropdownMenu(
                    expanded = menuOpen,
                    onDismissRequest = { menuOpen = false },
                ) {
                    DropdownMenuItem(
                        text = {
                            Text(if (channel.isHidden) "Unhide Channel" else "Hide Channel")
                        },
                        onClick = {
                            menuOpen = false
                            onToggleHidden()
                        },
                    )
                }
            }

            if (isActive) {
                Row(
                    modifier = Modifier
                        .align(Alignment.BottomEnd)
                        .padding(8.dp)
                        .clip(RoundedCornerShape(2.dp))
                        .background(NetflixRed)
                        .padding(horizontal = 6.dp, vertical = 2.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        text = "PLAYING",
                        color = NetflixWhite,
                        fontSize = 9.sp,
                        fontWeight = FontWeight.Bold,
                        letterSpacing = 2.sp,
                    )
                }
            }
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color(0xFF181818))
                .padding(horizontal = 10.dp, vertical = 8.dp),
        ) {
            Text(
                text = channel.name,
                color = NetflixWhite,
                fontSize = 13.sp,
                fontWeight = FontWeight.Bold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                text = channel.groupTitle.ifBlank { "Live channel" },
                color = NetflixLightGray,
                fontSize = 11.sp,
                fontWeight = FontWeight.Medium,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}
