package com.velox.app.presentation.ui.screens.settings.components

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
import com.velox.app.presentation.viewmodel.SystemAdminUiState
import com.velox.app.ui.theme.*
import com.velox.app.ui.theme.TextMuted
import java.net.HttpURLConnection
import java.net.URL
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.json.JSONObject

@Composable
fun AppVersionCard(uiState: SystemAdminUiState, onCheckUpdate: () -> Unit) {
    val context = LocalContext.current

    val currentVersion = com.velox.app.BuildConfig.VERSION_NAME
    val currentVersionCode = com.velox.app.BuildConfig.VERSION_CODE

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 24.dp, vertical = 24.dp)
    ) {
        Text(
            text = "APP VERSION",
            color = NetflixLightGray,
            fontSize = 12.sp,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(bottom = 8.dp)
        )
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(containerColor = Color(0xFF1E1E1E)),
            shape = RoundedCornerShape(12.dp) // Match SettingsGroupCard
        ) {
            Column(
                modifier = Modifier.fillMaxWidth().padding(16.dp)
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Image(
                        painter = painterResource(id = com.velox.app.R.drawable.velox_logo),
                        contentDescription = "App Logo",
                        modifier = Modifier.size(48.dp).clip(RoundedCornerShape(8.dp))
                    )
                    Spacer(modifier = Modifier.width(16.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text("Velox for Android", color = NetflixWhite, fontWeight = FontWeight.Bold, fontSize = 16.sp)
                        Text("Version $currentVersion", color = NetflixLightGray, fontSize = 13.sp)
                    }

                    // Refresh Button
                    if (uiState.latestVersion == null) {
                        IconButton(
                            onClick = { if (!uiState.isCheckingUpdate) onCheckUpdate() },
                            modifier = Modifier.background(NetflixGray.copy(alpha = 0.2f), RoundedCornerShape(50))
                        ) {
                            if (uiState.isCheckingUpdate) {
                                CircularProgressIndicator(modifier = Modifier.size(20.dp), color = NetflixWhite, strokeWidth = 2.dp)
                            } else {
                                Icon(LucideIcons.Refresh, contentDescription = "Check for Updates", tint = NetflixWhite, modifier = Modifier.size(20.dp))
                            }
                        }
                    }
                }

                if (uiState.updateStatus.isNotEmpty()) {
                    Spacer(modifier = Modifier.height(16.dp))
                    HorizontalDivider(color = NetflixGray.copy(alpha = 0.2f))
                    Spacer(modifier = Modifier.height(16.dp))

                    if (uiState.latestVersion != null && uiState.updateDownloadUrl != null) {
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Column(modifier = Modifier.weight(1f)) {
                                Text("Update Available", color = NetflixWhite, fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
                                Text(uiState.updateStatus, color = if (uiState.updateIsMandatory) NetflixRed else NetflixLightGray, fontSize = 13.sp)
                            }
                            Spacer(modifier = Modifier.width(12.dp))
                            Button(
                                onClick = {
                                    com.velox.app.utils.AppUpdater.downloadAndInstall(context, uiState.updateDownloadUrl)
                                },
                                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                                contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp)
                            ) {
                                Text(if (uiState.updateIsMandatory) "UPDATE" else "DOWNLOAD", fontWeight = FontWeight.Bold)
                            }
                        }
                    } else {
                        Text(uiState.updateStatus, color = NetflixLightGray, fontSize = 13.sp, modifier = Modifier.fillMaxWidth(), textAlign = TextAlign.Center)
                    }
                }
            }
        }
    }


}
