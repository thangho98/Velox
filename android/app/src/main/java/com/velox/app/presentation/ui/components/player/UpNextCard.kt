package com.velox.app.presentation.ui.components.player

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite

@Suppress("UNUSED_PARAMETER") // countdown not yet displayed but part of public API
@Composable
internal fun UpNextCard(
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

