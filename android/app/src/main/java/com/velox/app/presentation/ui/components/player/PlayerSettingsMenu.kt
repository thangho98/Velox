package com.velox.app.presentation.ui.components.player

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.R
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.viewmodel.AspectRatioUi
import com.velox.app.presentation.viewmodel.QualityOptionUi
import com.velox.app.presentation.viewmodel.RepeatModeUi
import com.velox.app.presentation.viewmodel.SubtitleBackgroundUi
import com.velox.app.presentation.viewmodel.SubtitleSizeUi

@Composable
internal fun SettingsSectionTitle(
    title: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = Modifier.padding(bottom = 8.dp),
    ) {
        Icon(
            icon,
            contentDescription = null,
            modifier = Modifier.size(14.dp),
            tint = Color.White.copy(alpha = 0.4f),
        )
        Spacer(modifier = Modifier.width(5.dp))
        Text(
            text = title,
            color = Color.White.copy(alpha = 0.4f),
            fontSize = 10.sp,
            fontWeight = FontWeight.SemiBold,
            letterSpacing = 1.sp,
        )
    }
}

@Composable
internal fun SettingsMenu(
    expanded: Boolean,
    onDismiss: () -> Unit,
    aspectRatio: AspectRatioUi,
    onSetAspectRatio: (AspectRatioUi) -> Unit,
    maxQuality: Int,
    qualityOptions: List<QualityOptionUi>,
    onSetQuality: (Int) -> Unit,
    subtitleDelay: Float,
    onAdjustSubtitleDelay: (Float) -> Unit,
    subtitleSize: SubtitleSizeUi,
    onSetSubtitleSize: (SubtitleSizeUi) -> Unit,
    subtitleColor: String,
    onSetSubtitleColor: (String) -> Unit,
    subtitleBackground: SubtitleBackgroundUi,
    onSetSubtitleBackground: (SubtitleBackgroundUi) -> Unit,
    repeatMode: RepeatModeUi,
    onSetRepeatMode: (RepeatModeUi) -> Unit,
    onTogglePlaybackStats: () -> Unit,
) {
    if (!expanded) return
    var settingsView by remember { mutableStateOf("main") }

    androidx.compose.ui.window.Popup(
        alignment = Alignment.BottomEnd,
        offset = androidx.compose.ui.unit.IntOffset(-40, -180),
        onDismissRequest = onDismiss,
        properties = androidx.compose.ui.window.PopupProperties(focusable = true),
    ) {
        Box(
            modifier = Modifier
                .width(240.dp)
                .background(Color(0xFF1E1E1E), RoundedCornerShape(12.dp))
                .border(1.dp, Color.White.copy(alpha = 0.1f), RoundedCornerShape(12.dp)),
        ) {
            if (settingsView == "quality") {
                QualitySubmenu(
                    qualityOptions = qualityOptions,
                    maxQuality = maxQuality,
                    onSetQuality = { h ->
                        onSetQuality(h)
                        settingsView = "main"
                    },
                    onBack = { settingsView = "main" },
                )
            } else {
                MainSettingsView(
                    aspectRatio = aspectRatio,
                    onSetAspectRatio = onSetAspectRatio,
                    subtitleDelay = subtitleDelay,
                    onAdjustSubtitleDelay = onAdjustSubtitleDelay,
                    subtitleSize = subtitleSize,
                    onSetSubtitleSize = onSetSubtitleSize,
                    subtitleColor = subtitleColor,
                    onSetSubtitleColor = onSetSubtitleColor,
                    subtitleBackground = subtitleBackground,
                    onSetSubtitleBackground = onSetSubtitleBackground,
                    repeatMode = repeatMode,
                    onSetRepeatMode = onSetRepeatMode,
                    onTogglePlaybackStats = onTogglePlaybackStats,
                    maxQuality = maxQuality,
                    onOpenQuality = { settingsView = "quality" },
                )
            }
        }
    }
}

@Composable
private fun QualitySubmenu(
    qualityOptions: List<QualityOptionUi>,
    maxQuality: Int,
    onSetQuality: (Int) -> Unit,
    onBack: () -> Unit,
) {
    Column {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .clickable { onBack() }
                .padding(horizontal = 16.dp, vertical = 10.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Icon(
                LucideIcons.ChevronLeft,
                contentDescription = null,
                modifier = Modifier.size(14.dp),
                tint = Color.White.copy(0.7f),
            )
            Text(
                stringResource(R.string.player_quality),
                color = Color.White.copy(0.7f),
                fontSize = 12.sp,
                fontWeight = FontWeight.SemiBold,
            )
        }
        HorizontalDivider(
            color = Color.White.copy(alpha = 0.1f),
            thickness = 1.dp,
        )
        Column(
            modifier = Modifier
                .heightIn(max = 300.dp)
                .verticalScroll(rememberScrollState())
                .padding(vertical = 4.dp),
        ) {
            qualityOptions.forEach { option ->
                QualityOptionRow(
                    option = option,
                    isSelected = maxQuality == option.height,
                    onSelect = { onSetQuality(option.height) },
                )
            }
            HorizontalDivider(
                color = Color.White.copy(alpha = 0.1f),
                thickness = 1.dp,
            )
            QualityOptionRow(
                option = null,
                isSelected = maxQuality == 0,
                onSelect = { onSetQuality(0) },
            )
        }
    }
}

@Composable
private fun QualityOptionRow(
    option: QualityOptionUi?,
    isSelected: Boolean,
    onSelect: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onSelect() }
            .then(
                if (isSelected) {
                    Modifier.background(Color.White.copy(alpha = 0.1f))
                } else {
                    Modifier
                },
            )
            .padding(horizontal = 16.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(6.dp),
        ) {
            val label = option?.label ?: stringResource(R.string.player_auto)
            Text(
                text = label,
                color = if (isSelected) Color.White else Color.White.copy(alpha = 0.7f),
                fontSize = 12.sp,
            )
            if (option?.instant == true) {
                Icon(
                    LucideIcons.FlashOn,
                    contentDescription = null,
                    modifier = Modifier.size(11.dp),
                    tint = Color(0xFFFACC15),
                )
            }
        }
        if (isSelected) {
            Icon(
                LucideIcons.Check,
                contentDescription = null,
                modifier = Modifier.size(14.dp),
                tint = Color.White,
            )
        }
    }
}

@Composable
private fun MainSettingsView(
    aspectRatio: AspectRatioUi,
    onSetAspectRatio: (AspectRatioUi) -> Unit,
    subtitleDelay: Float,
    onAdjustSubtitleDelay: (Float) -> Unit,
    subtitleSize: SubtitleSizeUi,
    onSetSubtitleSize: (SubtitleSizeUi) -> Unit,
    subtitleColor: String,
    onSetSubtitleColor: (String) -> Unit,
    subtitleBackground: SubtitleBackgroundUi,
    onSetSubtitleBackground: (SubtitleBackgroundUi) -> Unit,
    repeatMode: RepeatModeUi,
    onSetRepeatMode: (RepeatModeUi) -> Unit,
    onTogglePlaybackStats: () -> Unit,
    maxQuality: Int,
    onOpenQuality: () -> Unit,
) {
    Column(
        modifier = Modifier
            .heightIn(max = 480.dp)
            .verticalScroll(rememberScrollState())
            .padding(12.dp),
    ) {
        // ASPECT RATIO
        SettingsSectionTitle("ASPECT RATIO", LucideIcons.Fullscreen)
        AspectRatioRow(aspectRatio = aspectRatio, onSetAspectRatio = onSetAspectRatio)

        HorizontalDivider(
            color = Color.White.copy(alpha = 0.1f),
            thickness = 1.dp,
            modifier = Modifier.padding(vertical = 10.dp),
        )

        // SUBTITLES
        SettingsSectionTitle("SUBTITLES", LucideIcons.Subtitles)
        SubtitleSizeRow(subtitleSize = subtitleSize, onSetSubtitleSize = onSetSubtitleSize)
        Spacer(modifier = Modifier.height(4.dp))
        SubtitleBgRow(
            subtitleBackground = subtitleBackground,
            onSetSubtitleBackground = onSetSubtitleBackground,
        )
        Spacer(modifier = Modifier.height(8.dp))
        SubtitleColorRow(
            subtitleColor = subtitleColor,
            onSetSubtitleColor = onSetSubtitleColor,
        )
        Spacer(modifier = Modifier.height(10.dp))
        DelayRow(subtitleDelay = subtitleDelay, onAdjustSubtitleDelay = onAdjustSubtitleDelay)
        Spacer(modifier = Modifier.height(6.dp))
        DelayButtons(subtitleDelay = subtitleDelay, onAdjustSubtitleDelay = onAdjustSubtitleDelay)

        HorizontalDivider(
            color = Color.White.copy(alpha = 0.1f),
            thickness = 1.dp,
            modifier = Modifier.padding(vertical = 10.dp),
        )

        // QUALITY
        QualityRow(maxQuality = maxQuality, onOpenQuality = onOpenQuality)

        HorizontalDivider(
            color = Color.White.copy(alpha = 0.1f),
            thickness = 1.dp,
            modifier = Modifier.padding(vertical = 10.dp),
        )

        // REPEAT
        SettingsSectionTitle("REPEAT", LucideIcons.Repeat)
        RepeatRow(repeatMode = repeatMode, onSetRepeatMode = onSetRepeatMode)

        HorizontalDivider(
            color = Color.White.copy(alpha = 0.1f),
            thickness = 1.dp,
            modifier = Modifier.padding(vertical = 10.dp),
        )

        // PLAYBACK INFO
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(8.dp))
                .clickable { onTogglePlaybackStats() }
                .padding(horizontal = 12.dp, vertical = 6.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(
                LucideIcons.ShowChart,
                contentDescription = null,
                modifier = Modifier.size(14.dp),
                tint = Color.White.copy(0.5f),
            )
            Spacer(modifier = Modifier.width(5.dp))
            Text(
                stringResource(R.string.player_playback_info),
                color = Color.White.copy(0.7f),
                fontSize = 12.sp,
            )
        }
    }
}

@Composable
private fun AspectRatioRow(
    aspectRatio: AspectRatioUi,
    onSetAspectRatio: (AspectRatioUi) -> Unit,
) {
    val options = listOf("Auto", "Cover", "Fill")
    val selectedIndex = when (aspectRatio) {
        AspectRatioUi.Contain -> 0
        AspectRatioUi.Cover -> 1
        AspectRatioUi.Fill -> 2
    }
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        options.forEachIndexed { i, text ->
            val isSelected = selectedIndex == i
            Box(
                modifier = Modifier
                    .weight(1f)
                    .height(32.dp)
                    .background(
                        if (isSelected) {
                            Color.White.copy(alpha = 0.2f)
                        } else {
                            Color.White.copy(alpha = 0.05f)
                        },
                        RoundedCornerShape(8.dp),
                    )
                    .clickable { onSetAspectRatio(AspectRatioUi.values()[i]) },
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text,
                    color = if (isSelected) Color.White else Color.White.copy(0.7f),
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Medium,
                )
            }
        }
    }
}

@Composable
private fun SubtitleSizeRow(
    subtitleSize: SubtitleSizeUi,
    onSetSubtitleSize: (SubtitleSizeUi) -> Unit,
) {
    val sizeOptions = listOf(
        "S" to SubtitleSizeUi.Small,
        "M" to SubtitleSizeUi.Medium,
        "L" to SubtitleSizeUi.Large,
    )
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        sizeOptions.forEach { (text, size) ->
            val isSelected = subtitleSize == size
            Box(
                modifier = Modifier
                    .weight(1f)
                    .height(32.dp)
                    .background(
                        if (isSelected) {
                            Color.White.copy(alpha = 0.2f)
                        } else {
                            Color.White.copy(alpha = 0.05f)
                        },
                        RoundedCornerShape(8.dp),
                    )
                    .clickable { onSetSubtitleSize(size) },
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text,
                    color = if (isSelected) Color.White else Color.White.copy(0.7f),
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Medium,
                )
            }
        }
    }
}

@Composable
private fun SubtitleBgRow(
    subtitleBackground: SubtitleBackgroundUi,
    onSetSubtitleBackground: (SubtitleBackgroundUi) -> Unit,
) {
    val bgOptions = listOf(
        "None" to SubtitleBackgroundUi.None,
        "Semi" to SubtitleBackgroundUi.Semi,
        "Solid" to SubtitleBackgroundUi.Solid,
    )
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        bgOptions.forEach { (text, bg) ->
            val isSelected = subtitleBackground == bg
            Box(
                modifier = Modifier
                    .weight(1f)
                    .height(32.dp)
                    .background(
                        if (isSelected) {
                            Color.White.copy(alpha = 0.2f)
                        } else {
                            Color.White.copy(alpha = 0.05f)
                        },
                        RoundedCornerShape(8.dp),
                    )
                    .clickable { onSetSubtitleBackground(bg) },
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text,
                    color = if (isSelected) Color.White else Color.White.copy(0.7f),
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Medium,
                )
            }
        }
    }
}

@Composable
private fun SubtitleColorRow(
    subtitleColor: String,
    onSetSubtitleColor: (String) -> Unit,
) {
    val colorOptions = listOf(
        "#ffffff" to Color.White,
        "#fde047" to Color(0xFFFDE047),
        "#4ade80" to Color(0xFF4ADE80),
        "#60a5fa" to Color(0xFF60A5FA),
    )
    Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
        colorOptions.forEach { (hex, color) ->
            val isSelected = subtitleColor == hex
            Box(
                modifier = Modifier
                    .size(20.dp)
                    .background(color, CircleShape)
                    .border(
                        if (isSelected) 2.dp else 1.dp,
                        if (isSelected) Color.White else Color.White.copy(alpha = 0.2f),
                        CircleShape,
                    )
                    .clickable { onSetSubtitleColor(hex) },
            )
        }
    }
}

@Composable
private fun DelayRow(
    subtitleDelay: Float,
    onAdjustSubtitleDelay: (Float) -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White.copy(alpha = 0.04f), RoundedCornerShape(8.dp))
            .padding(horizontal = 12.dp, vertical = 8.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                stringResource(R.string.player_delay),
                color = Color.White.copy(0.7f),
                fontSize = 12.sp,
            )
            val delayText = if (subtitleDelay > 0) {
                "+${String.format("%.2f", subtitleDelay)}s"
            } else {
                "${String.format("%.2f", subtitleDelay)}s"
            }
            Text(
                delayText,
                color = Color.White,
                fontSize = 12.sp,
                fontWeight = FontWeight.SemiBold,
            )
        }
    }
}

@Composable
private fun DelayButtons(
    subtitleDelay: Float,
    onAdjustSubtitleDelay: (Float) -> Unit,
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        Box(
            modifier = Modifier
                .weight(1f)
                .height(32.dp)
                .background(Color.White.copy(alpha = 0.05f), RoundedCornerShape(8.dp))
                .clickable { onAdjustSubtitleDelay(-0.25f) },
            contentAlignment = Alignment.Center,
        ) {
            Text(
                "-0.25s",
                color = Color.White.copy(0.7f),
                fontSize = 12.sp,
                fontWeight = FontWeight.Medium,
            )
        }
        Box(
            modifier = Modifier
                .weight(1f)
                .height(32.dp)
                .background(Color.White.copy(alpha = 0.05f), RoundedCornerShape(8.dp))
                .clickable { onAdjustSubtitleDelay(-subtitleDelay) },
            contentAlignment = Alignment.Center,
        ) {
            Text(
                stringResource(R.string.player_reset),
                color = Color.White.copy(0.7f),
                fontSize = 12.sp,
                fontWeight = FontWeight.Medium,
            )
        }
        Box(
            modifier = Modifier
                .weight(1f)
                .height(32.dp)
                .background(Color.White.copy(alpha = 0.05f), RoundedCornerShape(8.dp))
                .clickable { onAdjustSubtitleDelay(0.25f) },
            contentAlignment = Alignment.Center,
        ) {
            Text(
                "+0.25s",
                color = Color.White.copy(0.7f),
                fontSize = 12.sp,
                fontWeight = FontWeight.Medium,
            )
        }
    }
}

@Composable
private fun QualityRow(
    maxQuality: Int,
    onOpenQuality: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(8.dp))
            .clickable { onOpenQuality() }
            .padding(horizontal = 12.dp, vertical = 6.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Icon(
                LucideIcons.FlashOn,
                contentDescription = null,
                modifier = Modifier.size(14.dp),
                tint = Color.White.copy(0.5f),
            )
            Spacer(modifier = Modifier.width(5.dp))
            Text(
                stringResource(R.string.player_quality),
                color = Color.White.copy(0.7f),
                fontSize = 12.sp,
            )
        }
        Row(verticalAlignment = Alignment.CenterVertically) {
            val qualityText = if (maxQuality == 0) "Auto" else "${maxQuality}p"
            Text(
                qualityText,
                color = Color.White.copy(0.5f),
                fontSize = 12.sp,
            )
            Icon(
                LucideIcons.ChevronRight,
                contentDescription = null,
                tint = Color.White.copy(0.5f),
                modifier = Modifier.size(14.dp),
            )
        }
    }
}

@Composable
private fun RepeatRow(
    repeatMode: RepeatModeUi,
    onSetRepeatMode: (RepeatModeUi) -> Unit,
) {
    val repeatOptions = listOf(
        "None" to RepeatModeUi.None,
        "One" to RepeatModeUi.One,
        "All" to RepeatModeUi.All,
    )
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        repeatOptions.forEach { (text, mode) ->
            val isSelected = repeatMode == mode
            Box(
                modifier = Modifier
                    .weight(1f)
                    .height(32.dp)
                    .background(
                        if (isSelected) {
                            Color.White.copy(alpha = 0.2f)
                        } else {
                            Color.White.copy(alpha = 0.05f)
                        },
                        RoundedCornerShape(8.dp),
                    )
                    .clickable { onSetRepeatMode(mode) },
                contentAlignment = Alignment.Center,
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(3.dp),
                ) {
                    if (isSelected) {
                        val icon = if (mode == RepeatModeUi.One) {
                            LucideIcons.RepeatOne
                        } else {
                            LucideIcons.Repeat
                        }
                        Icon(
                            icon,
                            contentDescription = null,
                            modifier = Modifier.size(14.dp),
                            tint = Color.White,
                        )
                    }
                    Text(
                        text,
                        color = if (isSelected) Color.White else Color.White.copy(0.7f),
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Medium,
                    )
                }
            }
        }
    }
}
