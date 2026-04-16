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
fun WebhooksSectionRoute(
    viewModel: SystemAdminViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    
    var showAddWebhookDialog by remember { mutableStateOf(false) }
    var editingWebhook by remember { mutableStateOf<Webhook?>(null) }
    
    LaunchedEffect(Unit) {
        viewModel.loadWebhooks()
    }
    
    if (showAddWebhookDialog || editingWebhook != null) {
        WebhookDialog(
            webhook = editingWebhook,
            onDismiss = {
                showAddWebhookDialog = false
                editingWebhook = null
            },
            onSave = { url, events, active ->
                if (editingWebhook != null) {
                    viewModel.updateWebhook(editingWebhook!!.id, url, events, active)
                } else {
                    viewModel.createWebhook(url, events, active)
                }
                showAddWebhookDialog = false
                editingWebhook = null
            },
        )
    }
    
    WebhooksSectionContent(
        uiState = uiState,
        onAddClick = { showAddWebhookDialog = true },
        onEditClick = { editingWebhook = it },
        onDeleteClick = { viewModel.deleteWebhook(it) },
    )
}

@Composable
internal fun WebhooksSectionContent(
    uiState: SystemAdminUiState,
    onAddClick: () -> Unit,
    onEditClick: (Webhook) -> Unit,
    onDeleteClick: (Int) -> Unit,
) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text("Webhooks", color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Button(
                onClick = onAddClick,
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                shape = RoundedCornerShape(8.dp),
            ) {
                Icon(LucideIcons.Add, contentDescription = null)
                Spacer(modifier = Modifier.width(4.dp))
                Text("Add Webhook")
            }
        }

        if (uiState.webhooks.isEmpty()) {
            Text("No webhooks configured", color = NetflixLightGray)
        } else {
            uiState.webhooks.forEach { webhook ->
                WebhookCard(
                    webhook = webhook,
                    onEdit = { onEditClick(webhook) },
                    onDelete = { onDeleteClick(webhook.id) },
                )
            }
        }
    }
}

@Composable
internal fun WebhookCard(
    webhook: Webhook,
    onEdit: () -> Unit,
    onDelete: () -> Unit,
) {
    var showMenu by remember { mutableStateOf(false) }

    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(8.dp),
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(webhook.url, color = NetflixWhite)
                Text(webhook.events.joinToString(", "), color = NetflixLightGray, fontSize = 12.sp)
            }
            Surface(
                shape = RoundedCornerShape(4.dp),
                color = if (webhook.active) NetflixRed.copy(alpha = 0.2f) else NetflixGray,
            ) {
                Text(
                    if (webhook.active) "Active" else "Inactive",
                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp),
                    color = if (webhook.active) NetflixRed else NetflixLightGray,
                    fontSize = 10.sp,
                )
            }
            Box {
                IconButton(onClick = { showMenu = true }) {
                    Icon(LucideIcons.MoreVert, contentDescription = "More", tint = NetflixLightGray)
                }
                DropdownMenu(
                    expanded = showMenu,
                    onDismissRequest = { showMenu = false },
                ) {
                    DropdownMenuItem(
                        text = { Text("Edit") },
                        onClick = { onEdit(); showMenu = false },
                        leadingIcon = { Icon(LucideIcons.Edit, contentDescription = null) },
                    )
                    DropdownMenuItem(
                        text = { Text("Delete", color = NetflixRed) },
                        onClick = { onDelete(); showMenu = false },
                        leadingIcon = { Icon(LucideIcons.Delete, contentDescription = null, tint = NetflixRed) },
                    )
                }
            }
        }
    }
}

// Dialogs

@Composable

internal fun WebhookDialog(
    webhook: Webhook?,
    onDismiss: () -> Unit,
    onSave: (url: String, events: List<String>, active: Boolean) -> Unit,
) {
    var url by remember { mutableStateOf(webhook?.url ?: "") }
    var eventsText by remember { mutableStateOf(webhook?.events?.joinToString(", ") ?: "") }
    var active by remember { mutableStateOf(webhook?.active ?: true) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(if (webhook != null) "Edit Webhook" else "Add Webhook", color = NetflixWhite) },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                OutlinedTextField(
                    value = url,
                    onValueChange = { url = it },
                    label = { Text("URL") },
                    modifier = Modifier.fillMaxWidth(),
                    colors = textFieldColors(),
                )
                OutlinedTextField(
                    value = eventsText,
                    onValueChange = { eventsText = it },
                    label = { Text("Events (comma separated)") },
                    modifier = Modifier.fillMaxWidth(),
                    colors = textFieldColors(),
                    placeholder = { Text("media_added, library_scan_complete", color = NetflixGray) },
                )
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text("Active", color = NetflixWhite)
                    Switch(
                        checked = active,
                        onCheckedChange = { active = it },
                        colors = SwitchDefaults.colors(checkedThumbColor = NetflixRed, checkedTrackColor = NetflixRed),
                    )
                }
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    if (url.isNotBlank()) {
                        onSave(
                            url,
                            eventsText.split(",").map { it.trim() }.filter { it.isNotEmpty() },
                            active,
                        )
                    }
                },
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
            ) {
                Text("Save")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text(stringResource(R.string.action_cancel), color = NetflixLightGray)
            }
        },
        containerColor = NetflixDark,
    )
}

// Helper components
