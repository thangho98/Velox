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
import com.velox.app.presentation.viewmodel.SystemAdminUiState
import com.velox.app.presentation.viewmodel.SystemAdminViewModel
import com.velox.app.ui.theme.*
import com.velox.app.ui.theme.TextMuted
import java.net.HttpURLConnection
import java.net.URL
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.json.JSONObject

@Composable
fun DashboardSectionRoute(
    viewModel: SystemAdminViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    
    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.loadDashboardData()
    }
    
    DashboardSectionContent(
        viewModel = viewModel,
        uiState = uiState
    )
}

@Composable
internal fun DashboardSectionContent(viewModel: SystemAdminViewModel, uiState: SystemAdminUiState) {
    val serverInfo = uiState.serverInfo
    val libraryStats = uiState.libraryStats

    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        // Section header (matches web SectionHeader)
        Column {
            Text("Dashboard", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text("Server information and status", color = NetflixLightGray, fontSize = 14.sp)
        }

        if (uiState.isLoading) {
            CircularProgressIndicator(color = NetflixRed, modifier = Modifier.padding(16.dp))
            return
        }

        // Stats Cards
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            DashboardStatCard("Total Media", serverInfo?.mediaCount?.toString() ?: "0", modifier = Modifier.weight(1f))
            DashboardStatCard("Series", serverInfo?.seriesCount?.toString() ?: "0", modifier = Modifier.weight(1f))
            DashboardStatCard("Users", serverInfo?.userCount?.toString() ?: "0", modifier = Modifier.weight(1f))
            DashboardStatCard("Total Size", formatBytes(serverInfo?.totalSize ?: 0L), modifier = Modifier.weight(1f))
        }

        // Server Information
        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = NetflixDark,
            shape = RoundedCornerShape(8.dp),
        ) {
            Column(modifier = Modifier.padding(20.dp)) {
                Text("Server Information", color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.height(16.dp))

                DashboardInfoRow("Version", serverInfo?.version ?: "Unknown")
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                DashboardInfoRow("Uptime", serverInfo?.uptime ?: "Unknown")
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                DashboardInfoRow("Go Version", serverInfo?.goVersion ?: "Unknown")
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                DashboardInfoRow("OS / Arch", "${serverInfo?.os ?: "?"} / ${serverInfo?.arch ?: "?"}")
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                DashboardInfoRow("FFmpeg", serverInfo?.ffmpegVersion ?: "Unknown")
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                DashboardInfoRow("HW Acceleration", serverInfo?.hwAccel?.takeIf { it.isNotBlank() } ?: "None")
                HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                DashboardInfoRow("Database", serverInfo?.database ?: "SQLite")
            }
        }

        // Library Statistics
        if (libraryStats.isNotEmpty()) {
            Surface(
                modifier = Modifier.fillMaxWidth(),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Column {
                    Text(
                        "Library Statistics",
                        color = NetflixWhite,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.SemiBold,
                        modifier = Modifier.padding(20.dp)
                    )

                    // Header Row
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(Color.Black.copy(alpha = 0.5f))
                            .padding(horizontal = 20.dp, vertical = 10.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text("Library", color = NetflixLightGray, fontSize = 12.sp, modifier = Modifier.weight(1.5f))
                        Text("Type", color = NetflixLightGray, fontSize = 12.sp, modifier = Modifier.weight(1f))
                        Text("Items", color = NetflixLightGray, fontSize = 12.sp, modifier = Modifier.weight(0.5f), textAlign = androidx.compose.ui.text.style.TextAlign.End)
                        Text("Size", color = NetflixLightGray, fontSize = 12.sp, modifier = Modifier.weight(1.2f), textAlign = androidx.compose.ui.text.style.TextAlign.End)
                    }

                    // List Rows
                    libraryStats.forEachIndexed { index, lib ->
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 20.dp, vertical = 12.dp),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Text(lib.name, color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.5f))

                            val typeColor = when(lib.type.lowercase(java.util.Locale.US)) {
                                "movies" -> Color(0xFF60A5FA)
                                "tvshows" -> Color(0xFFC084FC)
                                else -> Color(0xFF4ADE80)
                            }
                            Text(lib.type, color = typeColor, fontSize = 12.sp, modifier = Modifier.weight(1f))

                            Text(lib.itemCount.toString(), color = Color(0xFFD1D5DB), fontSize = 14.sp, modifier = Modifier.weight(0.5f), textAlign = androidx.compose.ui.text.style.TextAlign.End)
                            Text(formatBytes(lib.totalSize), color = Color(0xFFD1D5DB), fontSize = 14.sp, modifier = Modifier.weight(1.2f), textAlign = androidx.compose.ui.text.style.TextAlign.End)
                        }
                        if (index < libraryStats.lastIndex) {
                            HorizontalDivider(color = Color.White.copy(alpha = 0.1f))
                        }
                    }
                }
            }
        }
    }
}

@Composable
internal fun DashboardStatCard(label: String, value: String, modifier: Modifier = Modifier) {
    Surface(
        modifier = modifier,
        color = NetflixDark,
        shape = RoundedCornerShape(8.dp),
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            horizontalAlignment = Alignment.Start,
        ) {
            Text(label, color = NetflixLightGray, fontSize = 12.sp)
            Spacer(modifier = Modifier.height(4.dp))
            Text(value, color = NetflixWhite, fontSize = 24.sp, fontWeight = FontWeight.Bold, maxLines = 1)
        }
    }
}

@Composable
internal fun DashboardInfoRow(label: String, value: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 12.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(label, color = NetflixLightGray, fontSize = 14.sp)
        Text(value, color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium)
    }
}
