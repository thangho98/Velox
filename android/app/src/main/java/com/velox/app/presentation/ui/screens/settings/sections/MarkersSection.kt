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

@Composable
fun MarkersSectionRoute(
    viewModel: MediaSettingsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    
    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.loadMarkerStats()
        viewModel.loadLibraries()
    }
    
    MarkersSectionContent(
        viewModel = viewModel,
        uiState = uiState
    )
}

@Composable
internal fun MarkersSectionContent(viewModel: MediaSettingsViewModel, uiState: MediaSettingsUiState) {
    var selectedLibraryId by remember { mutableStateOf(0) }

    // Auto-select first library (or fix stale selection)
    LaunchedEffect(uiState.libraries) {
        if (uiState.libraries.isNotEmpty() && (selectedLibraryId == 0 || uiState.libraries.none { it.id == selectedLibraryId })) {
            selectedLibraryId = uiState.libraries.first().id
        }
    }

    val libraryOptions = uiState.libraries.map { it.id.toString() to it.name }

    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        // Section header (matches web SectionHeader)
        Column {
            Text(stringResource(R.string.settings_title_markers), color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text(stringResource(R.string.markers_desc), color = NetflixLightGray, fontSize = 14.sp)
        }

        // ── Stats Overview Card ──
        val stats = uiState.markerStats
        MarkerCard {
            Text(stringResource(R.string.markers_overview), color = NetflixWhite, fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
            Spacer(modifier = Modifier.height(16.dp))

            if (uiState.isLoading) {
                Text("Loading...", color = NetflixLightGray, fontSize = 14.sp)
            } else if (stats != null && stats.totalMarkers > 0) {
                // 4-column stat grid
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    MarkerStatBlock(
                        value = stats.totalMarkers.toString(),
                        label = "Total Markers",
                        valueColor = NetflixWhite,
                        modifier = Modifier.weight(1f)
                    )
                    MarkerStatBlock(
                        value = stats.introMarkers.toString(),
                        label = "Intro Markers",
                        valueColor = Color(0xFF60A5FA), // blue-400
                        modifier = Modifier.weight(1f)
                    )
                    MarkerStatBlock(
                        value = stats.creditsMarkers.toString(),
                        label = "Credits Markers",
                        valueColor = Color(0xFFC084FC), // purple-400
                        modifier = Modifier.weight(1f)
                    )
                    MarkerStatBlock(
                        value = stats.totalFiles.toString(),
                        label = "Total Files",
                        valueColor = Color(0xFFD1D5DB), // gray-300
                        modifier = Modifier.weight(1f)
                    )
                }

                Spacer(modifier = Modifier.height(20.dp))

                // Coverage bars
                val introPercent = if (stats.totalFiles > 0) ((stats.filesWithIntro.toFloat() / stats.totalFiles) * 100).toInt() else 0
                val creditsPercent = if (stats.totalFiles > 0) ((stats.filesWithCredits.toFloat() / stats.totalFiles) * 100).toInt() else 0

                MarkerCoverageBar(
                    label = "Files with Intro (${stats.filesWithIntro}/${stats.totalFiles})",
                    percent = introPercent,
                    color = Color(0xFF3B82F6) // blue-500
                )

                Spacer(modifier = Modifier.height(12.dp))

                MarkerCoverageBar(
                    label = "Files with Credits (${stats.filesWithCredits}/${stats.totalFiles})",
                    percent = creditsPercent,
                    color = Color(0xFFA855F7) // purple-500
                )

                // Source breakdown
                Spacer(modifier = Modifier.height(20.dp))
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                Spacer(modifier = Modifier.height(16.dp))

                Text(stringResource(R.string.markers_source), color = Color(0xFFD1D5DB), fontSize = 14.sp, fontWeight = FontWeight.Medium)
                Spacer(modifier = Modifier.height(8.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(24.dp)) {
                    if (stats.chapterSource > 0) {
                        Text("Chapter: ${stats.chapterSource}", color = Color(0xFF4ADE80), fontSize = 14.sp)
                    }
                    if (stats.fingerprintSource > 0) {
                        Text("Fingerprint: ${stats.fingerprintSource}", color = Color(0xFF60A5FA), fontSize = 14.sp)
                    }
                    if (stats.manualSource > 0) {
                        Text("Manual: ${stats.manualSource}", color = Color(0xFFFACC15), fontSize = 14.sp)
                    }
                }
            } else {
                // Empty state — matches web's "No markers detected yet"
                Text(
                    "No markers detected yet. Run detection below to scan your library.",
                    color = Color(0xFF6B7280), // gray-500
                    fontSize = 14.sp
                )
            }
        }

        // ── Run Detection Card ──
        MarkerCard {
            Text(stringResource(R.string.markers_detection), color = NetflixWhite, fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                "Run audio fingerprint analysis to detect intro/credits segments across your library. Works best for TV series with recurring intros.",
                color = NetflixLightGray,
                fontSize = 14.sp
            )

            Spacer(modifier = Modifier.height(20.dp))

            // Reuse SettingsDropdown component (ModalBottomSheet picker)
            SettingsDropdown(
                label = "Select Library",
                value = selectedLibraryId.toString(),
                options = libraryOptions,
                onValueChange = { selectedLibraryId = it.toIntOrNull() ?: 0 },
            )

            Spacer(modifier = Modifier.height(16.dp))

            Button(
                onClick = { viewModel.runMarkerDetection(selectedLibraryId) },
                colors = ButtonDefaults.buttonColors(
                    containerColor = NetflixRed,
                    contentColor = NetflixWhite,
                    disabledContainerColor = NetflixRed.copy(alpha = 0.5f)
                ),
                shape = RoundedCornerShape(8.dp),
                enabled = !uiState.isRunningDetection && selectedLibraryId != 0,
                modifier = Modifier.fillMaxWidth().height(48.dp)
            ) {
                if (uiState.isRunningDetection) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(18.dp),
                        color = NetflixWhite,
                        strokeWidth = 2.dp
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Running...", fontWeight = FontWeight.SemiBold)
                } else {
                    Text(stringResource(R.string.action_run_detection), fontWeight = FontWeight.SemiBold)
                }
            }

            // Running indicator — matches web's real-time progress area
            if (uiState.isRunningDetection) {
                Spacer(modifier = Modifier.height(16.dp))
                Surface(
                    modifier = Modifier.fillMaxWidth(),
                    color = NetflixGray,
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Row(
                        modifier = Modifier.padding(16.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(20.dp),
                            color = NetflixRed,
                            strokeWidth = 2.dp
                        )
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text("Analyzing library...", color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium)
                            Text("Audio fingerprint detection in progress", color = NetflixLightGray, fontSize = 12.sp)
                        }
                    }
                }
            }

            // Success result — matches web's complete result panel
            uiState.successMessage?.let { message ->
                if (!uiState.isRunningDetection) {
                    Spacer(modifier = Modifier.height(16.dp))
                    Surface(
                        modifier = Modifier.fillMaxWidth(),
                        color = NetflixGray,
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Row(
                            modifier = Modifier.padding(16.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(
                                Icons.Default.CheckCircle,
                                contentDescription = null,
                                tint = Color(0xFF4ADE80),
                                modifier = Modifier.size(20.dp)
                            )
                            Spacer(modifier = Modifier.width(12.dp))
                            Text(message, color = Color(0xFF4ADE80), fontSize = 14.sp)
                        }
                    }
                }
            }

            // Error result — matches web's error panel
            uiState.error?.let { errorMsg ->
                if (!uiState.isRunningDetection) {
                    Spacer(modifier = Modifier.height(16.dp))
                    Surface(
                        modifier = Modifier.fillMaxWidth(),
                        color = Color(0xFF7F1D1D).copy(alpha = 0.3f), // red-900/30
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Row(
                            modifier = Modifier.padding(16.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(
                                Icons.Default.Warning,
                                contentDescription = null,
                                tint = Color(0xFFF87171), // red-400
                                modifier = Modifier.size(20.dp)
                            )
                            Spacer(modifier = Modifier.width(12.dp))
                            Text(errorMsg, color = Color(0xFFF87171), fontSize = 14.sp)
                        }
                    }
                }
            }
        }

        // ── How it works Card ──
        MarkerCard {
            Text("How it works", color = NetflixWhite, fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
            Spacer(modifier = Modifier.height(16.dp))

            MarkerBulletItem(Color(0xFF4ADE80), "Chapter — Extracted from video chapter metadata during scan")
            Spacer(modifier = Modifier.height(12.dp))
            MarkerBulletItem(Color(0xFF60A5FA), "Fingerprint — Audio analysis compares episodes to find recurring segments")
            Spacer(modifier = Modifier.height(12.dp))
            MarkerBulletItem(Color(0xFFFACC15), "Manual — User-adjusted skip boundaries manually set in player")
        }
    }
}

/** Card wrapper matching web's `rounded-lg bg-netflix-black p-6 ring-1 ring-white/10` */
@Composable
internal fun MarkerCard(content: @Composable ColumnScope.() -> Unit) {
    Surface(
        modifier = Modifier
            .fillMaxWidth(),
        color = NetflixBlack,
        shape = RoundedCornerShape(12.dp),
        border = BorderStroke(1.dp, Color.White.copy(alpha = 0.1f))
    ) {
        Column(modifier = Modifier.padding(20.dp)) {
            content()
        }
    }
}

@Composable
internal fun MarkerStatBlock(value: String, label: String, valueColor: Color, modifier: Modifier = Modifier) {
    Column(
        modifier = modifier
            .background(NetflixGray, RoundedCornerShape(8.dp))
            .padding(12.dp)
    ) {
        Text(value, color = valueColor, fontSize = 24.sp, fontWeight = FontWeight.Bold)
        Spacer(modifier = Modifier.height(4.dp))
        Text(label, color = NetflixLightGray, fontSize = 11.sp, lineHeight = 14.sp)
    }
}

@Composable
internal fun MarkerCoverageBar(label: String, percent: Int, color: Color) {
    Column {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Text(label, color = NetflixLightGray, fontSize = 14.sp)
            Text("$percent%", color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium)
        }
        Spacer(modifier = Modifier.height(4.dp))
        LinearProgressIndicator(
            progress = { percent / 100f },
            modifier = Modifier
                .fillMaxWidth()
                .height(8.dp)
                .clip(RoundedCornerShape(4.dp)),
            color = color,
            trackColor = Color(0xFF374151), // gray-700
            drawStopIndicator = {}
        )
    }
}

@Composable
internal fun MarkerBulletItem(color: Color, text: String) {
    Row(verticalAlignment = Alignment.Top) {
        Box(
            modifier = Modifier
                .padding(top = 6.dp, end = 8.dp)
                .size(8.dp)
                .background(color, androidx.compose.foundation.shape.CircleShape)
        )
        Text(text, color = NetflixLightGray, fontSize = 14.sp, lineHeight = 20.sp)
    }
}
