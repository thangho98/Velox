package com.velox.app.presentation.ui.components

import androidx.compose.foundation.background
import com.velox.app.presentation.ui.components.LucideIcons
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.ClosedCaption
import androidx.compose.material.icons.filled.Subtitles
import androidx.compose.material.icons.filled.Translate
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Popup
import androidx.compose.ui.window.PopupProperties
import com.velox.app.presentation.viewmodel.PlayerUiState
import com.velox.app.presentation.viewmodel.SubtitleTrackUi
import com.velox.app.ui.theme.NetflixWhite

@Composable
fun SubtitleMenuPopup(
    uiState: PlayerUiState,
    onDismiss: () -> Unit,
    onSelectPrimary: (Int) -> Unit,
    onSelectSecondary: (Int) -> Unit,
    onSearchSubtitles: () -> Unit,
    onTranslateSubtitles: () -> Unit,
) {
    Popup(
        alignment = Alignment.BottomEnd,
        onDismissRequest = onDismiss,
        properties = PopupProperties(focusable = true, dismissOnClickOutside = true)
    ) {
        Box(
            modifier = Modifier
                .width(288.dp) // sm:w-72 (288px)
                .heightIn(max = 400.dp) // limit height to avoid clipping
                .padding(bottom = 48.dp) // float above the anchor
                .clip(RoundedCornerShape(16.dp)) // rounded-xl
                .background(Color(0xFF242424)) // exact web bg
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .verticalScroll(rememberScrollState())
            ) {
                // Title
                Text(
                    text = "Subtitles",
                    color = NetflixWhite,
                    fontSize = 14.sp,
                    fontWeight = FontWeight.SemiBold,
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 12.dp),
                    textAlign = TextAlign.Center
                )

                HorizontalDivider(color = Color.White.copy(alpha = 0.1f), thickness = 1.dp)

                // Primary Subtitles
                uiState.subtitleTracks.forEach { track ->
                    SubtitleMenuItem(
                        track = track,
                        isSelected = track.index == uiState.selectedSubtitleIndex,
                        onClick = {
                            onSelectPrimary(track.index)
                            onDismiss()
                        }
                    )
                }

                HorizontalDivider(color = Color.White.copy(alpha = 0.1f), thickness = 1.dp)
                
                // Secondary Title
                Text(
                    text = "SECONDARY SUBTITLE",
                    color = Color.White.copy(alpha = 0.4f),
                    fontSize = 10.sp,
                    fontWeight = FontWeight.SemiBold,
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(start = 16.dp, top = 12.dp, bottom = 8.dp)
                )

                // Secondary Subtitles
                uiState.subtitleTracks.forEach { track ->
                    SubtitleMenuItem(
                        track = track,
                        isSelected = track.index == uiState.selectedSecondarySubtitleIndex,
                        onClick = {
                            onSelectSecondary(track.index)
                            onDismiss()
                        }
                    )
                }

                HorizontalDivider(color = Color.White.copy(alpha = 0.1f), thickness = 1.dp)

                // Actions
                ActionMenuItem(
                    icon = LucideIcons.Translate,
                    text = "Translate Subtitle",
                    onClick = {
                        onTranslateSubtitles()
                        onDismiss()
                    }
                )

                ActionMenuItem(
                    icon = LucideIcons.Search,
                    text = "Search for Subtitles",
                    onClick = {
                        onSearchSubtitles()
                        onDismiss()
                    }
                )
            }
        }
    }
}

@Composable
private fun SubtitleMenuItem(
    track: SubtitleTrackUi,
    isSelected: Boolean,
    onClick: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(horizontal = 16.dp, vertical = 10.dp), // tight padding like web py-2.5
        verticalAlignment = Alignment.CenterVertically
    ) {
        Icon(
            imageVector = LucideIcons.Subtitles,
            contentDescription = null,
            tint = if (isSelected) NetflixWhite else Color.White.copy(alpha = 0.4f),
            modifier = Modifier.size(18.dp)
        )
        
        Spacer(modifier = Modifier.width(12.dp))
        
        Column(modifier = Modifier.weight(1f)) {
            val displayLabel = track.label.takeIf { it.isNotBlank() } ?: track.language.takeIf { it.isNotBlank() } ?: "Track ${track.index}"
            Text(
                text = if (track.index == -1) "Off" else displayLabel,
                color = if (isSelected) NetflixWhite else Color.White.copy(alpha = 0.7f),
                fontSize = 14.sp,
                fontWeight = FontWeight.Medium,
                lineHeight = 16.sp
            )
            if (!track.format.isNullOrEmpty() && track.index != -1) {
                Text(
                    text = track.format,
                    color = Color.White.copy(alpha = 0.4f),
                    fontSize = 12.sp,
                    lineHeight = 14.sp
                )
            }
        }
        
        // Simple text checkmark like web `✓`
        Text(
            text = if (isSelected) "✓" else "",
            color = NetflixWhite,
            fontSize = 14.sp,
            modifier = Modifier.defaultMinSize(minWidth = 16.dp),
            textAlign = TextAlign.Center
        )
    }
}

@Composable
private fun ActionMenuItem(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    text: String,
    onClick: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Icon(
            imageVector = icon,
            contentDescription = null,
            tint = Color.White.copy(alpha = 0.5f),
            modifier = Modifier.size(18.dp)
        )
        Spacer(modifier = Modifier.width(12.dp))
        Text(
            text = text,
            color = Color.White.copy(alpha = 0.5f),
            fontSize = 14.sp,
            fontWeight = FontWeight.Medium
        )
    }
}
