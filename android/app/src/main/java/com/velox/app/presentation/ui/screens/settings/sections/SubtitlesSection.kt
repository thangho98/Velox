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
fun SubtitlesSectionRoute(
    viewModel: MediaSettingsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    
    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.loadSubtitles()
    }
    
    SubtitlesSectionContent(
        viewModel = viewModel,
        uiState = uiState
    )
}

@Composable
internal fun SubtitlesSectionContent(viewModel: MediaSettingsViewModel, uiState: MediaSettingsUiState) {
    Column(verticalArrangement = Arrangement.spacedBy(24.dp)) {
        Column {
            Text(stringResource(R.string.settings_title_subtitles), color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text(stringResource(R.string.subtitles_desc), color = NetflixLightGray, fontSize = 14.sp)
        }

        // OpenSubtitles Settings
        OpenSubsSettingsCard(
            settings = uiState.openSubsSettings,
            onSave = { apiKey, username, password ->
                viewModel.updateOpenSubsSettings(
                    com.velox.app.data.model.UpdateOpenSubsRequest(
                        apiKey = apiKey,
                        username = username,
                        password = password.ifEmpty { null }
                    )
                )
            }
        )

        // Subdl Settings
        ProviderSettingsCard(
            name = "Subdl",
            description = "Provides subtitles using the Subdl API.",
            apiKey = uiState.subdlSettings?.apiKey ?: "",
            hasBuiltin = uiState.subdlSettings?.hasBuiltin ?: false,
            isRequired = false,
            placeholder = "Optional",
            url = "https://subdl.com/panel/api",
            onSave = { apiKey -> viewModel.updateProviderApiKey("subdl", apiKey) }
        )

        AITranslationSettingsCard(
            settings = uiState.aiTranslationSettings,
            onSave = { provider, apiKey, baseUrl, model ->
                viewModel.updateAITranslationSettings(
                    provider = provider,
                    apiKey = apiKey,
                    baseUrl = baseUrl,
                    model = model,
                )
            }
        )

        // DeepL Settings
        ProviderSettingsCard(
            name = "DeepL Translation",
            description = "Translate subtitles using DeepL instead of Google Translate.",
            apiKey = uiState.deepLSettings?.apiKey ?: "",
            hasBuiltin = false,
            isRequired = false,
            placeholder = "Leave empty to use Google Translate",
            url = "https://www.deepl.com/pro-api",
            onSave = { apiKey -> viewModel.updateProviderApiKey("deepl", apiKey) }
        )

        // Auto-Download Languages
        AutoSubSettingsCard(
            languages = uiState.autoSubSettings?.languages ?: "",
            onSave = { langs -> viewModel.updateAutoSubSettings(langs) }
        )
    }
}

internal fun defaultAITranslationBaseUrl(provider: String): String =
    when (provider) {
        "openai_compatible" -> "https://api.openai.com/v1"
        "gemini_compatible" -> "https://generativelanguage.googleapis.com/v1beta"
        "anthropic_compatible" -> "https://api.anthropic.com"
        else -> ""
    }

@Composable
internal fun AITranslationSettingsCard(
    settings: com.velox.app.data.model.AITranslationSettingsDto?,
    onSave: (String, String, String, String) -> Unit,
) {
    var provider by remember(settings?.provider) { mutableStateOf(settings?.provider ?: "") }
    var apiKey by remember(settings?.apiKey) { mutableStateOf(settings?.apiKey ?: "") }
    var baseUrl by remember(settings?.baseUrl, settings?.provider) {
        mutableStateOf(settings?.baseUrl ?: defaultAITranslationBaseUrl(settings?.provider ?: ""))
    }
    var model by remember(settings?.model) { mutableStateOf(settings?.model ?: "") }
    var apiKeyVisible by remember { mutableStateOf(false) }

    val providerOptions = listOf(
        "" to "Disabled",
        "openai_compatible" to "OpenAI Compatible",
        "gemini_compatible" to "Gemini Compatible",
        "anthropic_compatible" to "Anthropic Compatible",
    )
    val isEnabled = provider.isNotEmpty()
    val statusText = providerOptions.find { it.first == provider }?.second ?: "Disabled"
    val statusColor = if (isEnabled) Color(0xFF22C55E) else Color(0xFF6B7280)

    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(12.dp),
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(stringResource(R.string.subtitles_ai), color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
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
            Text(
                "Use an LLM provider to localize subtitles with better tone and context. Velox manages the system prompt and validates AI JSON output before saving.",
                color = NetflixLightGray,
                fontSize = 12.sp
            )
            Spacer(modifier = Modifier.height(16.dp))

            SettingsDropdown(
                label = "Provider",
                value = provider,
                options = providerOptions,
                onValueChange = { selected ->
                    provider = selected
                    baseUrl = defaultAITranslationBaseUrl(selected)
                },
            )

            Spacer(modifier = Modifier.height(12.dp))
            Text(stringResource(R.string.subtitles_api_key), color = NetflixLightGray, fontSize = 12.sp, fontWeight = FontWeight.Medium)
            Spacer(modifier = Modifier.height(4.dp))
            OutlinedTextField(
                value = apiKey,
                onValueChange = { apiKey = it },
                modifier = Modifier.fillMaxWidth(),
                enabled = isEnabled,
                placeholder = { Text("Provider API key") },
                visualTransformation = if (apiKeyVisible) androidx.compose.ui.text.input.VisualTransformation.None else androidx.compose.ui.text.input.PasswordVisualTransformation(),
                trailingIcon = {
                    val image = if (apiKeyVisible) LucideIcons.Visibility else LucideIcons.VisibilityOff
                    IconButton(onClick = { apiKeyVisible = !apiKeyVisible }) { Icon(image, "Toggle API Key visibility") }
                },
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )

            Spacer(modifier = Modifier.height(12.dp))
            OutlinedTextField(
                value = baseUrl,
                onValueChange = { baseUrl = it },
                label = { Text("Base URL") },
                modifier = Modifier.fillMaxWidth(),
                enabled = isEnabled,
                placeholder = { Text(defaultAITranslationBaseUrl(provider)) },
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )

            Spacer(modifier = Modifier.height(12.dp))
            OutlinedTextField(
                value = model,
                onValueChange = { model = it },
                label = { Text("Model") },
                modifier = Modifier.fillMaxWidth(),
                enabled = isEnabled,
                placeholder = { Text("e.g. gpt-4.1-mini") },
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )

            Spacer(modifier = Modifier.height(8.dp))
            Text(
                "Prompt is system-managed. The response must pass strict JSON validation before a subtitle file is created.",
                color = NetflixLightGray,
                fontSize = 11.sp
            )

            Spacer(modifier = Modifier.height(16.dp))
            Button(
                onClick = { onSave(provider, apiKey, baseUrl, model) },
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

@Composable
internal fun OpenSubsSettingsCard(
    settings: com.velox.app.data.model.OpenSubsSettingsDto?,
    onSave: (String, String, String) -> Unit
) {
    var apiKey by androidx.compose.runtime.remember(settings?.apiKey) { androidx.compose.runtime.mutableStateOf(settings?.apiKey ?: "") }
    var username by androidx.compose.runtime.remember(settings?.username) { androidx.compose.runtime.mutableStateOf(settings?.username ?: "") }
    var password by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf("") }
    var passwordVisible by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf(false) }

    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(12.dp),
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            val statusColor = if (settings?.passwordSet == true && apiKey.isNotEmpty()) androidx.compose.ui.graphics.Color(0xFF22C55E) else androidx.compose.ui.graphics.Color(0xFF6B7280)

            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(stringResource(R.string.subtitles_os), color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.width(8.dp))
                if (settings?.passwordSet == true && apiKey.isNotEmpty()) {
                    Surface(color = statusColor.copy(alpha = 0.2f), shape = RoundedCornerShape(4.dp)) {
                        Text(stringResource(R.string.webhooks_enabled), color = statusColor, fontSize = 10.sp, fontWeight = FontWeight.Medium, modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp))
                    }
                }
            }
            Spacer(modifier = Modifier.height(4.dp))
            Text(stringResource(R.string.subtitles_os_desc), color = NetflixLightGray, fontSize = 12.sp)
            Spacer(modifier = Modifier.height(16.dp))

            var apiKeyVisible by androidx.compose.runtime.remember { androidx.compose.runtime.mutableStateOf(false) }

            // Step 1: API Key
            Row(verticalAlignment = Alignment.CenterVertically) {
                Surface(color = NetflixGray, shape = androidx.compose.foundation.shape.CircleShape, modifier = Modifier.size(20.dp)) {
                    Box(contentAlignment = Alignment.Center) { Text("1", color = NetflixWhite, fontSize = 10.sp, fontWeight = FontWeight.Bold) }
                }
                Spacer(modifier = Modifier.width(8.dp))
                Text(stringResource(R.string.subtitles_api_key), color = NetflixLightGray, fontSize = 12.sp, fontWeight = FontWeight.Medium)
            }
            Spacer(modifier = Modifier.height(4.dp))
            OutlinedTextField(
                value = apiKey,
                onValueChange = { apiKey = it },
                modifier = Modifier.fillMaxWidth().padding(start = 28.dp),
                visualTransformation = if (apiKeyVisible) androidx.compose.ui.text.input.VisualTransformation.None else androidx.compose.ui.text.input.PasswordVisualTransformation(),
                trailingIcon = {
                    val image = if (apiKeyVisible) LucideIcons.Visibility else LucideIcons.VisibilityOff
                    IconButton(onClick = { apiKeyVisible = !apiKeyVisible }) { Icon(image, "Toggle API Key visibility") }
                },
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )
            Spacer(modifier = Modifier.height(16.dp))

            // Step 2: Account
            Row(verticalAlignment = Alignment.CenterVertically) {
                Surface(color = NetflixGray, shape = androidx.compose.foundation.shape.CircleShape, modifier = Modifier.size(20.dp)) {
                    Box(contentAlignment = Alignment.Center) { Text("2", color = NetflixWhite, fontSize = 10.sp, fontWeight = FontWeight.Bold) }
                }
                Spacer(modifier = Modifier.width(8.dp))
                Text(stringResource(R.string.subtitles_account), color = NetflixLightGray, fontSize = 12.sp, fontWeight = FontWeight.Medium)
            }
            Spacer(modifier = Modifier.height(4.dp))
            OutlinedTextField(
                value = username,
                onValueChange = { username = it },
                label = { Text(stringResource(R.string.login_username)) },
                modifier = Modifier.fillMaxWidth().padding(start = 28.dp),
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )
            Spacer(modifier = Modifier.height(8.dp))
            OutlinedTextField(
                value = password,
                onValueChange = { password = it },
                label = { Text(if (settings?.passwordSet == true) "Leave blank to keep current" else "Password") },
                modifier = Modifier.fillMaxWidth().padding(start = 28.dp),
                visualTransformation = if (passwordVisible) androidx.compose.ui.text.input.VisualTransformation.None else androidx.compose.ui.text.input.PasswordVisualTransformation(),
                trailingIcon = {
                    val image = if (passwordVisible) LucideIcons.Visibility else LucideIcons.VisibilityOff
                    IconButton(onClick = { passwordVisible = !passwordVisible }) { Icon(image, "Toggle password visibility") }
                },
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                singleLine = true
            )

            Spacer(modifier = Modifier.height(16.dp))
            Button(
                onClick = { onSave(apiKey, username, password) },
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                shape = RoundedCornerShape(8.dp),
                modifier = Modifier.padding(start = 28.dp),
                contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
            ) {
                Icon(LucideIcons.Check, contentDescription = null, modifier = Modifier.size(16.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text(stringResource(R.string.action_save), fontWeight = FontWeight.Medium)
            }
        }
    }
}

@OptIn(androidx.compose.foundation.layout.ExperimentalLayoutApi::class)
@Composable
internal fun AutoSubSettingsCard(
    languages: String,
    onSave: (String) -> Unit
) {
    val commonLangs = listOf("en" to "English", "vi" to "Vietnamese", "fr" to "French", "de" to "German", "es" to "Spanish", "ja" to "Japanese", "ko" to "Korean", "zh" to "Chinese")
    val selectedLangs = androidx.compose.runtime.remember(languages) { androidx.compose.runtime.mutableStateListOf(*languages.split(",").filter { it.isNotEmpty() }.toTypedArray()) }

    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(12.dp),
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(stringResource(R.string.subtitles_auto_download), color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                Spacer(modifier = Modifier.width(8.dp))
                if (selectedLangs.isNotEmpty()) {
                    Surface(color = androidx.compose.ui.graphics.Color(0xFF3B82F6).copy(alpha = 0.2f), shape = RoundedCornerShape(4.dp)) {
                        Text("${selectedLangs.size} languages", color = androidx.compose.ui.graphics.Color(0xFF3B82F6), fontSize = 10.sp, fontWeight = FontWeight.Medium, modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp))
                    }
                }
            }
            Spacer(modifier = Modifier.height(4.dp))
            Text(stringResource(R.string.subtitles_auto_download_desc), color = NetflixLightGray, fontSize = 12.sp)
            Spacer(modifier = Modifier.height(12.dp))

            androidx.compose.foundation.layout.FlowRow(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                for (lang in commonLangs) {
                    val code = lang.first
                    val label = lang.second
                    val isSelected = selectedLangs.contains(code)

                    Surface(
                        color = if (isSelected) NetflixRed else NetflixGray,
                        shape = androidx.compose.foundation.shape.CircleShape,
                        modifier = Modifier.clickable {
                            if (isSelected) selectedLangs.remove(code) else selectedLangs.add(code)
                        }
                    ) {
                        Text(
                            label,
                            color = if (isSelected) NetflixWhite else NetflixLightGray,
                            fontSize = 12.sp,
                            fontWeight = FontWeight.Medium,
                            modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp)
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))
            Button(
                onClick = { onSave(selectedLangs.joinToString(",")) },
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
