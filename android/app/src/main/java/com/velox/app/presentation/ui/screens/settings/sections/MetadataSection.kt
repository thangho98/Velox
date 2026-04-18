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
fun MetadataSectionRoute(
    viewModel: MediaSettingsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.loadMetadata()
    }

    MetadataSectionContent(
        viewModel = viewModel,
        uiState = uiState
    )
}

@Composable
internal fun MetadataSectionContent(viewModel: MediaSettingsViewModel, uiState: MediaSettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        Column {
            Text(stringResource(R.string.settings_title_metadata), color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text(stringResource(R.string.settings_desc_metadata), color = NetflixLightGray, fontSize = 14.sp)
        }

        // AniList Settings
        ProviderSettingsCard(
            name = stringResource(R.string.metadata_provider_anilist_name),
            description = stringResource(R.string.metadata_provider_anilist_description),
            apiKey = uiState.anilistSettings?.apiKey ?: "",
            hasBuiltin = uiState.anilistSettings?.hasBuiltin ?: false,
            placeholder = stringResource(R.string.metadata_provider_anilist_placeholder),
            url = "https://anilist.gitbook.io/anilist-apiv2-docs",
            connectUrl = "https://project-s8tij.vercel.app",
            inputLabel = stringResource(R.string.metadata_provider_anilist_token_label),
            supportingText = if (uiState.anilistSettings?.hasBuiltin == true) {
                stringResource(R.string.metadata_provider_anilist_optional)
            } else {
                stringResource(R.string.metadata_provider_anilist_optional_no_builtin)
            },
            helperText = stringResource(R.string.metadata_provider_anilist_auth_note),
            configuredStatusText = stringResource(R.string.metadata_status_custom_token),
            builtinStatusText = stringResource(R.string.metadata_status_env_token),
            onSave = { apiKey -> viewModel.updateProviderApiKey("anilist", apiKey) }
        )

        // TMDb Settings
        ProviderSettingsCard(
            name = "The Movie Database (TMDb)",
            description = "Primary metadata provider for movies and TV shows.",
            apiKey = uiState.tmdbSettings?.apiKey ?: "",
            hasBuiltin = uiState.tmdbSettings?.hasBuiltin ?: true,
            placeholder = "Optional",
            url = "https://www.themoviedb.org/settings/api",
            inputLabel = stringResource(R.string.metadata_tmdb_api_label),
            onSave = { apiKey -> viewModel.updateProviderApiKey("tmdb", apiKey) }
        )

        // OMDb Settings
        ProviderSettingsCard(
            name = "Open Movie Database (OMDb)",
            description = "Provides additional ratings and metadata.",
            apiKey = uiState.omdbSettings?.apiKey ?: "",
            hasBuiltin = uiState.omdbSettings?.hasBuiltin ?: false,
            placeholder = "Required",
            url = "https://www.omdbapi.com/apikey.aspx",
            onSave = { apiKey -> viewModel.updateProviderApiKey("omdb", apiKey) }
        )

        // TVDB Settings
        ProviderSettingsCard(
            name = "TheTVDB",
            description = "Fall-back provider for TV shows and episodes.",
            apiKey = uiState.tvdbSettings?.apiKey ?: "",
            hasBuiltin = uiState.tvdbSettings?.hasBuiltin ?: false,
            placeholder = "Required",
            url = "https://thetvdb.com/api-information",
            onSave = { apiKey -> viewModel.updateProviderApiKey("tvdb", apiKey) }
        )

        // Fanart Settings
        ProviderSettingsCard(
            name = "Fanart.tv",
            description = "Provides high quality artwork, clearlogos, and backgrounds.",
            apiKey = uiState.fanartSettings?.apiKey ?: "",
            hasBuiltin = uiState.fanartSettings?.hasBuiltin ?: true,
            placeholder = "Optional",
            url = "https://fanart.tv/get-an-api-key/",
            onSave = { apiKey -> viewModel.updateProviderApiKey("fanart", apiKey) }
        )

        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = NetflixDark,
            shape = RoundedCornerShape(12.dp),
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text(stringResource(R.string.metadata_refresh_all), color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.height(4.dp))
                Text(stringResource(R.string.metadata_bulk_fetch), color = NetflixLightGray, fontSize = 14.sp)
                Spacer(modifier = Modifier.height(16.dp))
                Button(
                    onClick = { viewModel.refreshAllMetadata() },
                    colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                    shape = RoundedCornerShape(8.dp),
                ) {
                    Icon(LucideIcons.Refresh, contentDescription = null, modifier = Modifier.size(16.dp))
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(stringResource(R.string.action_refresh_metadata), fontWeight = FontWeight.SemiBold)
                }
            }
        }
    }
}

@Composable
internal fun ProviderSettingsCard(
    name: String,
    description: String,
    apiKey: String,
    hasBuiltin: Boolean,
    placeholder: String,
    url: String,
    connectUrl: String? = null,
    inputLabel: String? = null,
    supportingText: String? = null,
    helperText: String? = null,
    configuredStatusText: String? = null,
    builtinStatusText: String? = null,
    onSave: (String) -> Unit
) {
    val context = LocalContext.current
    val clipboardManager = context.getSystemService(android.content.ClipboardManager::class.java)
    var editedApiKey by androidx.compose.runtime.remember(apiKey) { androidx.compose.runtime.mutableStateOf(apiKey) }
    val statusCustomLabel = configuredStatusText ?: stringResource(R.string.metadata_status_custom_key)
    val statusBuiltinLabel = builtinStatusText ?: stringResource(R.string.metadata_status_env_key)
    val inputLabelText = inputLabel ?: stringResource(R.string.metadata_api_key_label)
    val toggleVisibilityText = stringResource(R.string.metadata_toggle_api_key_visibility)

    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(12.dp),
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            val statusText = if (apiKey.isNotEmpty()) statusCustomLabel else if (hasBuiltin) statusBuiltinLabel else stringResource(R.string.metadata_status_not_configured)
            val statusColor = if (apiKey.isNotEmpty()) androidx.compose.ui.graphics.Color(0xFF3B82F6) else if (hasBuiltin) androidx.compose.ui.graphics.Color(0xFF22C55E) else androidx.compose.ui.graphics.Color(0xFF6B7280)

            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(name, color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.width(8.dp))
                Surface(
                    color = statusColor.copy(alpha = 0.2f),
                    shape = RoundedCornerShape(4.dp),
                ) {
                    Text(
                        text = statusText,
                        color = statusColor,
                        fontSize = 10.sp,
                        fontWeight = FontWeight.Medium,
                        modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                    )
                }
            }
            Spacer(modifier = Modifier.height(4.dp))
            Text(description, color = NetflixLightGray, fontSize = 12.sp)
            Spacer(modifier = Modifier.height(4.dp))
            TextButton(
                onClick = {
                    context.startActivity(Intent(Intent.ACTION_VIEW, Uri.parse(url)))
                },
                contentPadding = PaddingValues(0.dp)
            ) {
                Text(
                    text = stringResource(R.string.action_learn_more),
                    color = NetflixRed,
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Medium
                )
            }
            Spacer(modifier = Modifier.height(16.dp))

            var apiKeyVisible by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf(false) }

            Text(inputLabelText, color = NetflixLightGray, fontSize = 12.sp, fontWeight = FontWeight.Medium)
            if (!supportingText.isNullOrBlank()) {
                Spacer(modifier = Modifier.height(4.dp))
                Text(supportingText, color = NetflixLightGray, fontSize = 11.sp)
            }
            if (!helperText.isNullOrBlank()) {
                Spacer(modifier = Modifier.height(4.dp))
                Text(helperText, color = NetflixLightGray, fontSize = 11.sp)
            }
            if (connectUrl != null) {
                Spacer(modifier = Modifier.height(4.dp))
                TextButton(
                    onClick = { context.startActivity(Intent(Intent.ACTION_VIEW, Uri.parse(connectUrl))) },
                    contentPadding = PaddingValues(0.dp),
                ) {
                    Text("Connect AniList to get your token", color = NetflixRed, fontSize = 11.sp, fontWeight = FontWeight.Medium)
                }
            }
            Spacer(modifier = Modifier.height(4.dp))
            OutlinedTextField(
                value = editedApiKey,
                onValueChange = { editedApiKey = it },
                placeholder = { Text(placeholder) },
                modifier = Modifier.fillMaxWidth(),
                visualTransformation = if (apiKeyVisible) androidx.compose.ui.text.input.VisualTransformation.None else androidx.compose.ui.text.input.PasswordVisualTransformation(),
                trailingIcon = {
                    Row {
                        IconButton(onClick = {
                            val clip = clipboardManager?.primaryClip
                            val text = clip?.getItemAt(0)?.text?.toString()?.trim()
                            if (!text.isNullOrEmpty()) editedApiKey = text
                        }) {
                            Icon(Icons.Default.ContentPaste, "Paste", tint = Color(0xFF9CA3AF))
                        }
                        val image = if (apiKeyVisible) LucideIcons.Visibility else LucideIcons.VisibilityOff
                        IconButton(onClick = { apiKeyVisible = !apiKeyVisible }) { Icon(image, toggleVisibilityText) }
                    }
                },
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )
            Spacer(modifier = Modifier.height(12.dp))
            Button(
                onClick = { onSave(editedApiKey) },
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                shape = RoundedCornerShape(8.dp),
                contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
            ) {
                Icon(LucideIcons.Check, contentDescription = null, modifier = Modifier.size(16.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text(stringResource(R.string.action_save), fontWeight = FontWeight.Medium)
            }
        }
    }
}
