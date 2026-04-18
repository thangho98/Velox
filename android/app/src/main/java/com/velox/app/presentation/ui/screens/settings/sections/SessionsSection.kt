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
import com.velox.app.presentation.viewmodel.UserProfileUiState
import com.velox.app.presentation.viewmodel.UserProfileViewModel
import com.velox.app.ui.theme.*
import com.velox.app.ui.theme.TextMuted
import java.net.HttpURLConnection
import java.net.URL
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.json.JSONObject

@Composable
fun SessionsSectionRoute(
    viewModel: UserProfileViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.loadProfileData()
    }

    SessionsSectionContent(
        viewModel = viewModel,
        uiState = uiState
    )
}

@Composable
internal fun SessionsSectionContent(viewModel: UserProfileViewModel, uiState: UserProfileUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Text(stringResource(R.string.settings_title_sessions), color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
        Text(stringResource(R.string.settings_desc_sessions), color = NetflixLightGray, fontSize = 14.sp)

        if (uiState.isLoading) {
            Box(
                modifier = Modifier.fillMaxWidth().padding(32.dp),
                contentAlignment = Alignment.Center
            ) {
                CircularProgressIndicator(color = NetflixRed)
            }
        } else if (uiState.sessions.isEmpty()) {
            Text(stringResource(R.string.sessions_empty), color = NetflixLightGray)
        } else {
            uiState.sessions.forEach { session ->
                SessionCard(
                    session = session,
                    onRevoke = { viewModel.revokeSession(session.id) },
                )
            }
        }
    }
}

@Composable
internal fun SessionCard(session: Session, onRevoke: () -> Unit) {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(8.dp),
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Box(
                modifier = Modifier
                    .size(40.dp)
                    .clip(RoundedCornerShape(4.dp))
                    .background(NetflixGray),
                contentAlignment = Alignment.Center
            ) {
                Icon(LucideIcons.Devices, contentDescription = null, tint = NetflixLightGray, modifier = Modifier.size(20.dp))
            }
            Spacer(modifier = Modifier.width(16.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(session.deviceName, color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium)
                Text(session.ipAddress, color = NetflixLightGray, fontSize = 12.sp)
                Text(
                    if (session.isCurrent) "Current session" else "Last active: ${session.lastActive}",
                    color = NetflixLightGray,
                    fontSize = 12.sp,
                )
            }
            if (!session.isCurrent) {
                Button(
                    onClick = onRevoke,
                    colors = ButtonDefaults.buttonColors(containerColor = NetflixGray),
                    shape = RoundedCornerShape(4.dp),
                    contentPadding = PaddingValues(horizontal = 12.dp, vertical = 6.dp),
                    modifier = Modifier.height(32.dp)
                ) {
                    Text(stringResource(R.string.action_revoke), color = NetflixWhite, fontSize = 14.sp)
                }
            }
        }
    }
}
