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
fun ActivitySectionRoute(
    viewModel: SystemAdminViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.loadActivity()
    }

    ActivitySectionContent(
        viewModel = viewModel,
        uiState = uiState
    )
}

@Composable
internal fun ActivitySectionContent(viewModel: SystemAdminViewModel, uiState: SystemAdminUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Text(stringResource(R.string.settings_title_activity), color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)

        if (uiState.isLoading && uiState.activityLogs.isEmpty()) {
            Text("Loading activity...", color = NetflixLightGray)
        } else if (uiState.activityLogs.isEmpty()) {
            Surface(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(160.dp),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Text(stringResource(R.string.activity_empty), color = NetflixLightGray, fontSize = 14.sp)
                }
            }
        } else {
            Surface(
                modifier = Modifier.fillMaxWidth(),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Column {
                    // Header Row
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp, vertical = 12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(stringResource(R.string.activity_time), color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.8f))
                        Text("User", color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.2f))
                        Text(stringResource(R.string.tasks_action), color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.5f))
                        Text(stringResource(R.string.activity_media), color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(2.5f))
                    }

                    HorizontalDivider(color = Color.White.copy(alpha = 0.05f))

                    // Activity List
                    uiState.activityLogs.forEachIndexed { index, log ->
                        if (index > 0) {
                            HorizontalDivider(color = Color.White.copy(alpha = 0.02f))
                        }
                        ActivityTableRow(log)
                    }
                }
            }
        }
    }
}

@Composable
internal fun ActivityTableRow(log: com.velox.app.data.model.ActivityLogDto) {
    val fmtDate = try {
        // e.g. "2026-04-10T11:22:25Z" -> "4/10/2026, 11:22:25"
        val parts = log.createdAt.split("T")
        val dateParts = parts[0].split("-")
        val yyyy = dateParts[0]
        val mm = dateParts[1].toInt()
        val dd = dateParts[2].toInt()
        val timePart = parts[1].replace("Z", "")
        "$mm/$dd/$yyyy, $timePart"
    } catch (e: Exception) {
        log.createdAt
    }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        // Time
        Text(fmtDate, color = NetflixLightGray, fontSize = 13.sp, modifier = Modifier.weight(1.8f))

        // User
        Text(log.username ?: "System", color = NetflixWhite, fontSize = 14.sp, modifier = Modifier.weight(1.2f))

        // Action
        Box(modifier = Modifier.weight(1.5f), contentAlignment = Alignment.CenterStart) {
            ActionBadge(log.action)
        }

        // Media
        val targetStr = log.mediaTitle?.takeIf { it.isNotBlank() } ?: "-"
        Text(targetStr, color = NetflixLightGray, fontSize = 13.sp, maxLines = 1, overflow = androidx.compose.ui.text.style.TextOverflow.Ellipsis, modifier = Modifier.weight(2.5f))
    }
}

@Composable
internal fun ActionBadge(action: String) {
    val (bgColor, textColor) = when (action) {
        "login" -> Pair(Color(0x333B82F6), Color(0xFF60A5FA)) // blue 500/20, blue 400
        "play_start" -> Pair(Color(0x3322C55E), Color(0xFF4ADE80)) // green 500/20, green 400
        "play_stop" -> Pair(Color(0x33EAB308), Color(0xFFFACC15)) // yellow 500/20, yellow 400
        "library_scan" -> Pair(Color(0x33A855F7), Color(0xFFC084FC)) // purple 500/20, purple 400
        "media_added" -> Pair(Color(0x3314B8A6), Color(0xFF2DD4BF)) // teal 500/20, teal 400
        else -> Pair(Color(0x336B7280), Color(0xFF9CA3AF)) // gray
    }

    Surface(
        color = bgColor,
        shape = RoundedCornerShape(4.dp)
    ) {
        Text(
            text = action,
            color = textColor,
            fontSize = 10.sp,
            fontWeight = FontWeight.Medium,
            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
        )
    }
}
