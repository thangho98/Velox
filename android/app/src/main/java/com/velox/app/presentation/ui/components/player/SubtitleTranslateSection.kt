package com.velox.app.presentation.ui.components.player

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.R
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.viewmodel.SubtitleTrackUi
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite

@Composable
fun SubtitleTranslateSection(
    subtitles: List<SubtitleTrackUi>,
    onTranslate: (subtitleId: Int, targetLanguage: String) -> Unit,
    onDismiss: () -> Unit,
    isTranslating: Boolean,
    modifier: Modifier = Modifier,
) {
    val translateLanguages = listOf(
        "vi" to "Vietnamese",
        "en" to "English",
        "fr" to "French",
        "de" to "German",
        "es" to "Spanish",
        "ja" to "Japanese",
        "ko" to "Korean",
        "zh" to "Chinese",
    )

    var selectedSubtitleId by remember(subtitles) { mutableIntStateOf(subtitles.firstOrNull()?.id ?: -1) }
    var selectedTargetLang by remember { mutableStateOf<String?>(null) }
    var submitted by remember { mutableStateOf(false) }

    LaunchedEffect(isTranslating) {
        if (submitted && !isTranslating) {
            onDismiss()
        }
    }

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        containerColor = NetflixBlack,
        modifier = modifier,
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 20.dp)
                .padding(bottom = 32.dp),
        ) {
            Text(
                text = "Translate Subtitle",
                color = NetflixWhite,
                fontSize = 18.sp,
                fontWeight = FontWeight.Bold,
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Source subtitle selector (if multiple)
            if (subtitles.size > 1) {
                Text(
                    text = "Source Subtitle",
                    color = NetflixWhite.copy(alpha = 0.7f),
                    fontSize = 14.sp,
                )
                Spacer(modifier = Modifier.height(8.dp))
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    subtitles.filter { it.id > 0 }.forEach { subtitle ->
                        FilterChip(
                            selected = selectedSubtitleId == subtitle.id,
                            onClick = { selectedSubtitleId = subtitle.id },
                            label = { Text(subtitle.label) },
                            colors = FilterChipDefaults.filterChipColors(
                                containerColor = Color(0xFF2A2A2A),
                                labelColor = NetflixWhite,
                                selectedContainerColor = NetflixRed,
                                selectedLabelColor = NetflixWhite,
                            ),
                        )
                    }
                }
                Spacer(modifier = Modifier.height(16.dp))
            }

            // Target language selector
            Text(
                text = "Translate To",
                color = NetflixWhite.copy(alpha = 0.7f),
                fontSize = 14.sp,
            )
            Spacer(modifier = Modifier.height(8.dp))
            Row(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                translateLanguages.forEach { lang ->
                    FilterChip(
                        selected = selectedTargetLang == lang.first,
                        onClick = { selectedTargetLang = lang.first },
                        label = { Text(lang.second) },
                        colors = FilterChipDefaults.filterChipColors(
                            containerColor = Color(0xFF2A2A2A),
                            labelColor = NetflixWhite,
                            selectedContainerColor = NetflixRed,
                            selectedLabelColor = NetflixWhite,
                        ),
                    )
                }
            }

            Spacer(modifier = Modifier.height(24.dp))

            // Translate button
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Button(
                    onClick = {
                        if (selectedSubtitleId >= 0 && selectedTargetLang != null) {
                            submitted = true
                            onTranslate(selectedSubtitleId, selectedTargetLang!!)
                        }
                    },
                    enabled = selectedSubtitleId >= 0 && selectedTargetLang != null && !isTranslating,
                    colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF2563EB)),
                    modifier = Modifier.weight(1f),
                    shape = RoundedCornerShape(8.dp),
                ) {
                    if (isTranslating) {
                        CircularProgressIndicator(
                            color = NetflixWhite,
                            modifier = Modifier.size(20.dp),
                            strokeWidth = 2.dp,
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                    }
                    Text(stringResource(R.string.player_translate))
                }

                TextButton(
                    onClick = onDismiss,
                    colors = ButtonDefaults.textButtonColors(contentColor = NetflixWhite.copy(alpha = 0.7f)),
                ) {
                    Text(stringResource(R.string.action_cancel))
                }
            }
        }
    }
}
