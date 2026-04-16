package com.velox.app.presentation.ui.screens.settings.sections

import android.content.Intent
import android.net.Uri
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.velox.app.R
import com.velox.app.domain.model.Library
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.ui.screens.settings.components.*
import com.velox.app.presentation.viewmodel.*
import com.velox.app.presentation.viewmodel.MediaSettingsUiState
import com.velox.app.presentation.viewmodel.MediaSettingsViewModel
import com.velox.app.ui.theme.*
import com.velox.app.ui.theme.TextMuted
import java.net.HttpURLConnection
import java.net.URL
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.json.JSONObject

internal fun formatBytes(bytes: Long): String {
    if (bytes <= 0) return "0 B"
    val units = arrayOf("B", "KB", "MB", "GB", "TB")
    val digitGroups = (Math.log10(bytes.toDouble()) / Math.log10(1024.0)).toInt()
    return String.format("%.1f %s", bytes / Math.pow(1024.0, digitGroups.toDouble()), units[digitGroups])
}

@Composable
fun PretranscodeSectionRoute(
    viewModel: MediaSettingsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    
    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.loadPretranscode()
        viewModel.loadLibraries()
    }
    
    androidx.compose.runtime.DisposableEffect(Unit) {
        viewModel.startPretranscodePolling()
        onDispose { viewModel.stopPretranscodePolling() }
    }
    
    PretranscodeSectionContent(
        viewModel = viewModel,
        uiState = uiState
    )
}

@Composable
internal fun PretranscodeSectionContent(viewModel: MediaSettingsViewModel, uiState: MediaSettingsUiState) {
    val settings = uiState.pretranscodeSettings
    val status = uiState.pretranscodeStatus
    val profiles = uiState.pretranscodeProfiles
    val estimate = uiState.pretranscodeEstimate
    val libraries = uiState.libraries

    var selectedLibraryId by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf(0) }
    var showCleanupConfirm by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf(false) }

    // Auto-select first library
    androidx.compose.runtime.LaunchedEffect(libraries, selectedLibraryId) {
        if (selectedLibraryId == 0 && libraries.isNotEmpty()) {
            selectedLibraryId = libraries.first().id
            viewModel.loadPretranscodeEstimate(selectedLibraryId)
        }
    }

    val isEncoding = status != null && (status.encoding > 0 || status.queued > 0) && !status.paused
    val progressPercent = if (status != null && status.total > 0) ((status.done.toFloat() / status.total.toFloat()) * 100).toInt() else 0

    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        Column {
            Text("Pre-transcode", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text("Encode media in advance for instant playback — no buffering, no waiting.", color = NetflixLightGray, fontSize = 14.sp)
        }

        // Enable Toggle
        Surface(modifier = Modifier.fillMaxWidth(), color = NetflixDark, shape = RoundedCornerShape(12.dp)) {
            Column(modifier = Modifier.padding(16.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text("Offline Encoding", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                        Spacer(modifier = Modifier.height(4.dp))
                        Text("Pre-encode your library into browser-compatible H.264+AAC MP4 files. Like Netflix — instant playback, zero transcoding delay.", color = NetflixLightGray, fontSize = 12.sp)
                    }
                    Switch(
                        checked = settings?.enabled == true,
                        onCheckedChange = { viewModel.updatePretranscodeSettings(enabled = it) },
                        colors = SwitchDefaults.colors(checkedThumbColor = NetflixWhite, checkedTrackColor = NetflixRed)
                    )
                }
            }
        }

        if (settings?.enabled == true) {
            // Quality Profiles
            Surface(modifier = Modifier.fillMaxWidth(), color = NetflixDark, shape = RoundedCornerShape(12.dp)) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Quality Profiles", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                    Spacer(modifier = Modifier.height(16.dp))

                    profiles.forEach { p ->
                        Surface(
                            color = if (p.enabled) NetflixRed.copy(alpha = 0.1f) else NetflixGray.copy(alpha = 0.2f),
                            shape = RoundedCornerShape(8.dp),
                            modifier = Modifier.fillMaxWidth().padding(bottom = 8.dp).clickable { viewModel.togglePretranscodeProfile(p.id, !p.enabled) }
                        ) {
                            Row(modifier = Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
                                Checkbox(
                                    checked = p.enabled,
                                    onCheckedChange = { viewModel.togglePretranscodeProfile(p.id, it) },
                                    colors = androidx.compose.material3.CheckboxDefaults.colors(checkedColor = NetflixRed)
                                )
                                Column(modifier = Modifier.weight(1f)) {
                                    Text(p.name, color = NetflixWhite, fontWeight = FontWeight.Medium, fontSize = 14.sp)
                                    Text("${p.height}p, ${p.videoBitrate / 1000}Mbps video + ${p.audioBitrate}kbps audio", color = NetflixLightGray, fontSize = 12.sp)
                                }
                                val gbPerFilm = (((p.videoBitrate + p.audioBitrate) * 5400) / 8 / 1024 / 1024.0).toFloat()
                                Text(String.format("~%.1f GB/film", gbPerFilm), color = NetflixLightGray, fontSize = 12.sp)
                            }
                        }
                    }
                }
            }

            // Schedule & Concurrency
            Surface(modifier = Modifier.fillMaxWidth(), color = NetflixDark, shape = RoundedCornerShape(12.dp)) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Schedule", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                    Spacer(modifier = Modifier.height(16.dp))

                    SettingsDropdown(
                        label = "Encode time",
                        value = settings.schedule.takeIf { it.isNotEmpty() } ?: "always",
                        options = listOf("always" to "Always (fastest)", "night" to "Night only (00:00–06:00)", "idle" to "When idle (no one watching)"),
                        onValueChange = { viewModel.updatePretranscodeSettings(schedule = it) }
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    SettingsDropdown(
                        label = "Concurrent jobs",
                        value = settings.concurrency.takeIf { it.isNotEmpty() } ?: "1",
                        options = listOf("1" to "1 (NAS-friendly)", "2" to "2", "3" to "3", "4" to "4 (powerful server)"),
                        onValueChange = { viewModel.updatePretranscodeSettings(concurrency = it) }
                    )
                }
            }

            // Storage Estimation
            Surface(modifier = Modifier.fillMaxWidth(), color = NetflixDark, shape = RoundedCornerShape(12.dp)) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Storage Estimation", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                    Spacer(modifier = Modifier.height(16.dp))

                    val libraryOptions = libraries.map { it.id.toString() to it.name }
                    SettingsDropdown(
                        label = "Library",
                        value = selectedLibraryId.toString(),
                        options = libraryOptions.ifEmpty { listOf("0" to "None") },
                        onValueChange = {
                            val newId = it.toIntOrNull() ?: 0
                            selectedLibraryId = newId
                            viewModel.loadPretranscodeEstimate(newId)
                        }
                    )

                    if (estimate != null) {
                        Spacer(modifier = Modifier.height(16.dp))
                        Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            estimate.profiles?.forEach { ep ->
                                Row(horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                                    Text("${ep.profileName} (${ep.fileCount} files)", color = NetflixLightGray, fontSize = 14.sp)
                                    Text(String.format("%.1f GB", ep.estimatedGb), color = NetflixWhite, fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
                                }
                            }
                            HorizontalDivider(color = NetflixGray, modifier = Modifier.padding(vertical = 4.dp))
                            Row(horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                                Text("Total", color = NetflixWhite, fontWeight = FontWeight.Bold, fontSize = 14.sp)
                                Text(formatBytes(estimate.totalBytes), color = NetflixWhite, fontWeight = FontWeight.Bold, fontSize = 14.sp)
                            }
                            Row(horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                                Text("Disk free", color = NetflixLightGray, fontSize = 14.sp)
                                val isEnough = estimate.diskFreeBytes > estimate.totalBytes
                                val color = if (isEnough) androidx.compose.ui.graphics.Color(0xFF22C55E) else androidx.compose.ui.graphics.Color(0xFFEF4444)
                                Text("${formatBytes(estimate.diskFreeBytes)} " + if (isEnough) "✓ Enough" else "✗ Not enough", color = color, fontSize = 14.sp)
                            }
                        }
                    }
                }
            }

            // Progress Dashboard
            if (status != null && status.total > 0) {
                Surface(modifier = Modifier.fillMaxWidth(), color = NetflixDark, shape = RoundedCornerShape(12.dp)) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text("Progress", color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                        Spacer(modifier = Modifier.height(16.dp))

                        Row(horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                            Text("${status.done}/${status.total} files", color = NetflixLightGray, fontSize = 12.sp)
                            Text("$progressPercent%", color = NetflixLightGray, fontSize = 12.sp)
                        }
                        Spacer(modifier = Modifier.height(4.dp))
                        LinearProgressIndicator(
                            progress = { progressPercent.toFloat() / 100f },
                            modifier = Modifier.fillMaxWidth().height(8.dp).clip(RoundedCornerShape(4.dp)),
                            color = NetflixRed,
                            trackColor = NetflixGray,
                            drawStopIndicator = {}
                        )

                        if (status.currentFile != null) {
                            Spacer(modifier = Modifier.height(12.dp))
                            Row {
                                Text("Encoding: ", color = NetflixLightGray, fontSize = 12.sp)
                                Text(status.currentFile, color = NetflixWhite, fontSize = 12.sp, maxLines = 1, modifier = Modifier.weight(1f))
                                if (status.speed != null) {
                                    Text("(${status.speed})", color = NetflixLightGray, fontSize = 12.sp)
                                }
                            }
                        }

                        Spacer(modifier = Modifier.height(16.dp))
                        Row(horizontalArrangement = Arrangement.SpaceEvenly, modifier = Modifier.fillMaxWidth()) {
                            Text("✓ Done: ${status.done}", color = androidx.compose.ui.graphics.Color(0xFF22C55E), fontSize = 12.sp)
                            if (status.failed > 0) Text("✗ Failed: ${status.failed}", color = androidx.compose.ui.graphics.Color(0xFFEF4444), fontSize = 12.sp)
                            Text("Queued: ${status.queued}", color = NetflixLightGray, fontSize = 12.sp)
                            Text("Disk: ${formatBytes(status.diskUsed)}", color = NetflixLightGray, fontSize = 12.sp)
                        }

                        Spacer(modifier = Modifier.height(16.dp))
                        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                            if (status.paused) {
                                Button(
                                    onClick = { viewModel.executePretranscodeAction("resume") },
                                    colors = ButtonDefaults.buttonColors(containerColor = androidx.compose.ui.graphics.Color(0xFF22C55E)),
                                    shape = RoundedCornerShape(8.dp)
                                ) {
                                    Icon(LucideIcons.PlayArrow, contentDescription = null, modifier = Modifier.size(16.dp))
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text("Resume")
                                }
                            } else if (isEncoding) {
                                Button(
                                    onClick = { viewModel.executePretranscodeAction("stop") /* Note: Stop acts as pause/stop currently in Velox server */ },
                                    colors = ButtonDefaults.buttonColors(containerColor = androidx.compose.ui.graphics.Color(0xFFEAB308)),
                                    shape = RoundedCornerShape(8.dp)
                                ) {
                                    Icon(LucideIcons.Pause, contentDescription = null, modifier = Modifier.size(16.dp))
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text("Pause")
                                }
                            }

                            Button(
                                onClick = { viewModel.executePretranscodeAction("stop") },
                                colors = ButtonDefaults.buttonColors(containerColor = NetflixGray),
                                shape = RoundedCornerShape(8.dp)
                            ) {
                                Icon(androidx.compose.material.icons.Icons.Filled.Close, contentDescription = null, modifier = Modifier.size(16.dp))
                                Spacer(modifier = Modifier.width(8.dp))
                                Text("Stop")
                            }
                        }
                    }
                }
            }

            // Actions
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                if (!isEncoding) {
                    Button(
                        onClick = { viewModel.executePretranscodeAction("start") },
                        colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Text("Start Encoding", fontWeight = FontWeight.SemiBold)
                    }
                }

                Button(
                    onClick = { showCleanupConfirm = true },
                    colors = ButtonDefaults.buttonColors(containerColor = NetflixGray),
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Icon(LucideIcons.Delete, contentDescription = null, modifier = Modifier.size(16.dp))
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Delete All Files")
                }
            }
        }
    }

    if (showCleanupConfirm) {
        AlertDialog(
            onDismissRequest = { showCleanupConfirm = false },
            title = { Text("Delete All Pre-transcode Files", color = NetflixWhite) },
            text = { Text("This will permanently delete all pre-encoded files and disable pre-transcode. Your original media files are NOT affected.", color = NetflixLightGray) },
            containerColor = NetflixDark,
            confirmButton = {
                TextButton(
                    onClick = {
                        viewModel.executePretranscodeAction("cleanup")
                        showCleanupConfirm = false
                    }
                ) {
                    Text("Delete All", color = androidx.compose.ui.graphics.Color(0xFFEF4444))
                }
            },
            dismissButton = {
                TextButton(onClick = { showCleanupConfirm = false }) {
                    Text(stringResource(R.string.action_cancel), color = NetflixWhite)
                }
            }
        )
    }
}

