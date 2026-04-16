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
fun PlaybackSectionRoute(
    viewModel: MediaSettingsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    
    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.loadPlayback()
    }
    
    PlaybackSectionContent(
        viewModel = viewModel,
        uiState = uiState
    )
}

@Composable
internal fun PlaybackSectionContent(viewModel: MediaSettingsViewModel, uiState: MediaSettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        Column {
            Text(stringResource(R.string.player_playback), color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text(stringResource(R.string.settings_playback_desc), color = NetflixLightGray, fontSize = 14.sp)
        }

        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = NetflixDark,
            shape = RoundedCornerShape(12.dp),
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text(stringResource(R.string.playback_mode), color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.height(4.dp))
                Text(stringResource(R.string.playback_mode_desc), color = NetflixLightGray, fontSize = 12.sp)
                Spacer(modifier = Modifier.height(16.dp))

                val currentMode = uiState.adminPlaybackSettings?.playbackMode ?: "auto"

                // Auto Option
                Surface(
                    modifier = Modifier.fillMaxWidth().clickable { viewModel.updatePlaybackMode("auto") },
                    color = if (currentMode == "auto") androidx.compose.ui.graphics.Color(0xFFE50914).copy(alpha = 0.1f) else androidx.compose.ui.graphics.Color.Transparent,
                    shape = RoundedCornerShape(8.dp),
                    border = androidx.compose.foundation.BorderStroke(2.dp, if (currentMode == "auto") NetflixRed else NetflixGray.copy(alpha = 0.5f))
                ) {
                    Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.Top) {
                        Surface(
                            shape = androidx.compose.foundation.shape.CircleShape,
                            color = if (currentMode == "auto") NetflixRed else androidx.compose.ui.graphics.Color.Transparent,
                            border = androidx.compose.foundation.BorderStroke(2.dp, if (currentMode == "auto") NetflixRed else androidx.compose.ui.graphics.Color.Gray),
                            modifier = Modifier.size(20.dp).padding(top = 2.dp)
                        ) {
                            if (currentMode == "auto") {
                                Box(contentAlignment = Alignment.Center) {
                                    Surface(color = NetflixWhite, shape = androidx.compose.foundation.shape.CircleShape, modifier = Modifier.size(8.dp)) {}
                                }
                            }
                        }
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text(stringResource(R.string.playback_auto), color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium)
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(stringResource(R.string.playback_auto_desc), color = NetflixLightGray, fontSize = 12.sp)
                        }
                    }
                }

                Spacer(modifier = Modifier.height(12.dp))

                // Direct Play Option
                Surface(
                    modifier = Modifier.fillMaxWidth().clickable { viewModel.updatePlaybackMode("direct_play") },
                    color = if (currentMode == "direct_play") androidx.compose.ui.graphics.Color(0xFFE50914).copy(alpha = 0.1f) else androidx.compose.ui.graphics.Color.Transparent,
                    shape = RoundedCornerShape(8.dp),
                    border = androidx.compose.foundation.BorderStroke(2.dp, if (currentMode == "direct_play") NetflixRed else NetflixGray.copy(alpha = 0.5f))
                ) {
                    Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.Top) {
                        Surface(
                            shape = androidx.compose.foundation.shape.CircleShape,
                            color = if (currentMode == "direct_play") NetflixRed else androidx.compose.ui.graphics.Color.Transparent,
                            border = androidx.compose.foundation.BorderStroke(2.dp, if (currentMode == "direct_play") NetflixRed else androidx.compose.ui.graphics.Color.Gray),
                            modifier = Modifier.size(20.dp).padding(top = 2.dp)
                        ) {
                            if (currentMode == "direct_play") {
                                Box(contentAlignment = Alignment.Center) {
                                    Surface(color = NetflixWhite, shape = androidx.compose.foundation.shape.CircleShape, modifier = Modifier.size(8.dp)) {}
                                }
                            }
                        }
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text(stringResource(R.string.playback_direct), color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium)
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(stringResource(R.string.playback_direct_desc), color = NetflixLightGray, fontSize = 12.sp)
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun CinemaSectionRoute(
    viewModel: MediaSettingsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    
    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.loadCinema()
    }
    
    CinemaSectionContent(
        viewModel = viewModel,
        uiState = uiState
    )
}

@Composable
internal fun CinemaSectionContent(viewModel: MediaSettingsViewModel, uiState: MediaSettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        Column {
            Text(stringResource(R.string.cinema_mode), color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text(stringResource(R.string.cinema_mode_desc), color = NetflixLightGray, fontSize = 14.sp)
        }

        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = androidx.compose.ui.graphics.Color(0xFFEAB308).copy(alpha = 0.1f),
            shape = RoundedCornerShape(8.dp),
            border = androidx.compose.foundation.BorderStroke(1.dp, androidx.compose.ui.graphics.Color(0xFFEAB308).copy(alpha = 0.5f))
        ) {
            Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                Icon(LucideIcons.Info, contentDescription = null, tint = androidx.compose.ui.graphics.Color(0xFFEAB308), modifier = Modifier.size(24.dp))
                Spacer(modifier = Modifier.width(12.dp))
                Text(stringResource(R.string.cinema_mode_warning), color = androidx.compose.ui.graphics.Color(0xFFFEF08A), fontSize = 13.sp)
            }
        }

        // Enable Cinema Mode
        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = NetflixDark,
            shape = RoundedCornerShape(12.dp),
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text(stringResource(R.string.cinema_enable), color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(stringResource(R.string.cinema_enable_desc), color = NetflixLightGray, fontSize = 12.sp)
                    }
                    Switch(
                        checked = uiState.adminCinemaSettings?.enabled ?: false,
                        onCheckedChange = { viewModel.updateCinemaSettings(enabled = it) },
                        colors = SwitchDefaults.colors(
                            checkedThumbColor = NetflixWhite,
                            checkedTrackColor = NetflixRed,
                            uncheckedThumbColor = NetflixLightGray,
                            uncheckedTrackColor = NetflixGray
                        )
                    )
                }
            }
        }

        // Max Trailers
        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = NetflixDark,
            shape = RoundedCornerShape(12.dp),
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text(stringResource(R.string.cinema_max_trailers), color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.height(4.dp))
                Text(stringResource(R.string.cinema_max_trailers_desc), color = NetflixLightGray, fontSize = 12.sp)
                Spacer(modifier = Modifier.height(16.dp))

                SettingsDropdown(
                    label = "",
                    value = uiState.adminCinemaSettings?.maxTrailers ?: "2",
                    options = listOf("0" to stringResource(R.string.cinema_trailers_original), "1" to stringResource(R.string.cinema_trailers_1), "2" to stringResource(R.string.cinema_trailers_2), "3" to stringResource(R.string.cinema_trailers_3)),
                    onValueChange = { viewModel.updateCinemaSettings(maxTrailers = it) }
                )
            }
        }

        // Custom Intro
        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = NetflixDark,
            shape = RoundedCornerShape(12.dp),
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                    Text(stringResource(R.string.cinema_custom_intro), color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                    if (uiState.adminCinemaSettings?.hasIntro == true) {
                        Surface(color = androidx.compose.ui.graphics.Color(0xFF22C55E).copy(alpha = 0.2f), shape = RoundedCornerShape(4.dp)) {
                            Text(stringResource(R.string.cinema_intro_uploaded), color = androidx.compose.ui.graphics.Color(0xFF22C55E), fontSize = 10.sp, fontWeight = FontWeight.Medium, modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp))
                        }
                    }
                }
                Spacer(modifier = Modifier.height(4.dp))
                Text(stringResource(R.string.cinema_upload_intro_desc), color = NetflixLightGray, fontSize = 12.sp)
                Spacer(modifier = Modifier.height(16.dp))

                Button(
                    onClick = { /* File picker integration required */ },
                    colors = ButtonDefaults.buttonColors(containerColor = androidx.compose.ui.graphics.Color(0xFF2563EB)),
                    shape = RoundedCornerShape(8.dp),
                ) {
                    Icon(LucideIcons.Upload, contentDescription = null, modifier = Modifier.size(16.dp))
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(if (uiState.adminCinemaSettings?.hasIntro == true) stringResource(R.string.cinema_upload_new) else stringResource(R.string.cinema_upload), fontWeight = FontWeight.Medium)
                }
            }
        }
    }
}

