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
fun ProfileSectionRoute(
    viewModel: UserProfileViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    
    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.loadProfileData()
    }
    
    ProfileSectionContent(
        viewModel = viewModel,
        uiState = uiState
    )
}

@Composable
internal fun ProfileSectionContent(viewModel: UserProfileViewModel, uiState: UserProfileUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Text("Profile", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
        Text("Manage your account information", color = NetflixLightGray, fontSize = 14.sp)

        OutlinedTextField(
            value = uiState.username,
            onValueChange = {},
            label = { Text(stringResource(R.string.login_username)) },
            enabled = false,
            modifier = Modifier.fillMaxWidth(),
            colors = OutlinedTextFieldDefaults.colors(
                disabledTextColor = NetflixWhite,
                disabledBorderColor = NetflixGray,
                disabledLabelColor = NetflixLightGray,
            ),
        )
        Text("Username cannot be changed", color = NetflixLightGray, fontSize = 12.sp)

        OutlinedTextField(
            value = uiState.displayName,
            onValueChange = viewModel::updateDisplayName,
            label = { Text("Display Name") },
            modifier = Modifier.fillMaxWidth(),
            colors = OutlinedTextFieldDefaults.colors(
                focusedTextColor = NetflixWhite,
                unfocusedTextColor = NetflixWhite,
                focusedBorderColor = NetflixRed,
                unfocusedBorderColor = NetflixGray,
                focusedLabelColor = NetflixRed,
                unfocusedLabelColor = NetflixLightGray,
            ),
        )

        // Role - read-only
        OutlinedTextField(
            value = if (uiState.isAdmin) "Administrator" else "User",
            onValueChange = {},
            label = { Text("Role") },
            enabled = false,
            modifier = Modifier.fillMaxWidth(),
            colors = OutlinedTextFieldDefaults.colors(
                disabledTextColor = NetflixLightGray,
                disabledBorderColor = NetflixGray,
                disabledLabelColor = NetflixLightGray,
            ),
        )

        uiState.error?.let {
            Text(it, color = NetflixRed, fontSize = 14.sp)
        }

        Button(
            onClick = viewModel::saveProfile,
            modifier = Modifier.wrapContentWidth(Alignment.Start),
            colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
            shape = RoundedCornerShape(8.dp),
            enabled = !uiState.isLoading,
        ) {
            if (uiState.isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.size(24.dp),
                    color = NetflixWhite,
                    strokeWidth = 2.dp
                )
            } else {
                Text("Save Changes")
            }
        }

        uiState.successMessage?.let {
            Text(it, color = androidx.compose.ui.graphics.Color(0xFF4CAF50), fontSize = 14.sp)
        }
    }
}

@Composable
fun PreferencesSectionRoute(
    viewModel: UserProfileViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    
    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.loadProfileData()
    }
    
    PreferencesSectionContent(
        viewModel = viewModel,
        uiState = uiState
    )
}

@Composable
internal fun PreferencesSectionContent(viewModel: UserProfileViewModel, uiState: UserProfileUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Text("Preferences", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
        Text("Customize your streaming experience", color = NetflixLightGray, fontSize = 14.sp)

        SettingsDropdown(
            label = "Subtitle Language",
            value = uiState.preferences.subtitleLanguage ?: "",
            options = listOf("" to "Auto", "vi" to "Tiếng Việt", "en" to "English"),
            onValueChange = { viewModel.updatePreference("subtitleLanguage", it) },
        )

        SettingsDropdown(
            label = "Audio Language",
            value = uiState.preferences.audioLanguage ?: "",
            options = listOf("" to "Auto", "vi" to "Tiếng Việt", "en" to "English"),
            onValueChange = { viewModel.updatePreference("audioLanguage", it) },
        )

        SettingsDropdown(
            label = "Max Streaming Quality",
            value = if (uiState.preferences.maxStreamingQuality.isNullOrBlank()) "original" else uiState.preferences.maxStreamingQuality,
            options = listOf("original" to "Original", "4k" to "4K (2160p)", "1080p" to "HD (1080p)", "720p" to "HD (720p)", "480p" to "SD (480p)"),
            onValueChange = { viewModel.updatePreference("maxQuality", it) },
        )

        SettingsDropdown(
            label = "Theme",
            value = if (uiState.preferences.theme.isNullOrBlank()) "system" else uiState.preferences.theme,
            options = listOf("system" to "System", "dark" to "Dark", "light" to "Light"),
            onValueChange = { viewModel.updatePreference("theme", it) },
        )

        SettingsDropdown(
            label = "Language",
            value = if (uiState.preferences.language.isNullOrBlank()) "en" else uiState.preferences.language,
            options = listOf("en" to "English", "vi" to "Tiếng Việt"),
            onValueChange = { viewModel.updatePreference("language", it) },
        )



        uiState.error?.let {
            Text(it, color = NetflixRed, fontSize = 14.sp)
        }

        Button(
            onClick = viewModel::savePreferences,
            modifier = Modifier.wrapContentWidth(Alignment.Start),
            colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
            shape = RoundedCornerShape(8.dp),
            enabled = !uiState.isLoading,
        ) {
            if (uiState.isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.size(24.dp),
                    color = NetflixWhite,
                    strokeWidth = 2.dp
                )
            } else {
                Text("Save Changes")
            }
        }

        uiState.successMessage?.let {
            Text(it, color = androidx.compose.ui.graphics.Color(0xFF4CAF50), fontSize = 14.sp)
        }
    }
}

@Composable
fun SecuritySectionRoute(
    viewModel: UserProfileViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    
    SecuritySectionContent(
        viewModel = viewModel,
        uiState = uiState
    )
}

@Composable
internal fun SecuritySectionContent(viewModel: UserProfileViewModel, uiState: UserProfileUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Text("Security", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
        Text("Change your password", color = NetflixLightGray, fontSize = 14.sp)

        OutlinedTextField(
            value = uiState.currentPassword,
            onValueChange = viewModel::updateCurrentPassword,
            label = { Text("Current Password") },
            modifier = Modifier.fillMaxWidth(),
            colors = textFieldColors(),
            shape = RoundedCornerShape(8.dp),
            visualTransformation = PasswordVisualTransformation(),
        )

        OutlinedTextField(
            value = uiState.newPassword,
            onValueChange = viewModel::updateNewPassword,
            label = { Text("New Password") },
            modifier = Modifier.fillMaxWidth(),
            colors = textFieldColors(),
            shape = RoundedCornerShape(8.dp),
            visualTransformation = PasswordVisualTransformation(),
        )

        OutlinedTextField(
            value = uiState.confirmPassword,
            onValueChange = viewModel::updateConfirmPassword,
            label = { Text("Confirm Password") },
            modifier = Modifier.fillMaxWidth(),
            colors = textFieldColors(),
            shape = RoundedCornerShape(8.dp),
            visualTransformation = PasswordVisualTransformation(),
        )

        uiState.securityError?.let {
            Text(it, color = NetflixRed, fontSize = 14.sp)
        }

        uiState.securitySuccess?.let {
            Text(it, color = androidx.compose.ui.graphics.Color(0xFF4CAF50), fontSize = 14.sp)
        }

        Button(
            onClick = viewModel::changePassword,
            modifier = Modifier.wrapContentWidth(Alignment.Start),
            colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
            shape = RoundedCornerShape(8.dp),
            enabled = !uiState.isLoading,
        ) {
            if (uiState.isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.size(24.dp),
                    color = NetflixWhite,
                    strokeWidth = 2.dp
                )
            } else {
                Text("Update Password")
            }
        }
    }
}
