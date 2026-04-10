package com.velox.app.presentation.ui.components

import androidx.compose.foundation.layout.*
import com.velox.app.presentation.ui.components.LucideIcons
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

@Composable
fun QuickActionsMenu(
    title: String,
    visible: Boolean,
    onPlay: () -> Unit,
    onViewDetails: () -> Unit,
    onDismiss: () -> Unit,
) {
    if (visible) {
        AlertDialog(
            onDismissRequest = onDismiss,
            containerColor = NetflixDark,
            title = {
                Text(
                    text = title,
                    color = NetflixWhite,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 1,
                )
            },
            text = {
                Column(
                    modifier = Modifier.fillMaxWidth(),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    // Play button
                    Surface(
                        modifier = Modifier.fillMaxWidth(),
                        color = NetflixRed,
                        shape = RoundedCornerShape(8.dp),
                        onClick = {
                            onDismiss()
                            onPlay()
                        },
                    ) {
                        Row(
                            modifier = Modifier.padding(14.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.Center,
                        ) {
                            Icon(
                                LucideIcons.PlayArrow,
                                contentDescription = null,
                                tint = NetflixWhite,
                                modifier = Modifier.size(24.dp),
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(
                                text = "Play",
                                color = NetflixWhite,
                                fontSize = 16.sp,
                                fontWeight = FontWeight.SemiBold,
                            )
                        }
                    }

                    // View Details button
                    Surface(
                        modifier = Modifier.fillMaxWidth(),
                        color = NetflixBlack,
                        shape = RoundedCornerShape(8.dp),
                        onClick = {
                            onDismiss()
                            onViewDetails()
                        },
                    ) {
                        Row(
                            modifier = Modifier.padding(14.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.Center,
                        ) {
                            Icon(
                                LucideIcons.Info,
                                contentDescription = null,
                                tint = NetflixWhite,
                                modifier = Modifier.size(24.dp),
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(
                                text = "View Details",
                                color = NetflixWhite,
                                fontSize = 16.sp,
                                fontWeight = FontWeight.Normal,
                            )
                        }
                    }
                }
            },
            confirmButton = {},
            dismissButton = {},
        )
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun QuickActionsMenuPreview() {
    VeloxTheme {
        QuickActionsMenu(
            title = "The Dark Knight Rises",
            visible = true,
            onPlay = {},
            onViewDetails = {},
            onDismiss = {},
        )
    }
}
