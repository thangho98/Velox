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
fun TasksSectionRoute(
    viewModel: SystemAdminViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.loadTasks()
    }

    TasksSectionContent(
        viewModel = viewModel,
        uiState = uiState
    )
}

@Composable
internal fun TasksSectionContent(viewModel: SystemAdminViewModel, uiState: SystemAdminUiState) {
    var editingTask by remember { mutableStateOf<com.velox.app.data.model.TaskDto?>(null) }

    fun cleanInterval(valStr: String?): String {
        if (valStr == null) return ""
        return valStr.replace("0m0s", "").replace("0s", "")
    }

    if (editingTask != null) {
        val task = editingTask!!
        EditIntervalDialog(
            taskName = task.name,
            currentInterval = cleanInterval(task.interval),
            onDismiss = { editingTask = null },
            onSave = { interval ->
                viewModel.updateTaskInterval(task.name, interval)
                editingTask = null
            }
        )
    }

    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Text(stringResource(R.string.settings_title_tasks), color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)

        if (uiState.tasks.isEmpty()) {
            Surface(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(160.dp),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Text(stringResource(R.string.tasks_empty), color = NetflixLightGray, fontSize = 14.sp)
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
                        Text(stringResource(R.string.tasks_task), color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(2f))
                        Text(stringResource(R.string.tasks_interval), color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1f))
                        Text(stringResource(R.string.tasks_last_next), color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(2f))
                        Text(stringResource(R.string.tasks_status), color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1f))
                        Text(stringResource(R.string.tasks_action), color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.5f))
                    }

                    HorizontalDivider(color = Color.White.copy(alpha = 0.05f))

                    // Tasks List
                    val sortedTasks = uiState.tasks.sortedBy { it.name }
                    sortedTasks.forEachIndexed { index, task ->
                        if (index > 0) {
                            HorizontalDivider(color = Color.White.copy(alpha = 0.02f))
                        }
                        TaskTableRow(
                            task = task,
                            cleanedInterval = cleanInterval(task.interval),
                            runningTaskName = uiState.runningTaskName,
                            onRun = { viewModel.runTask(task.name) },
                            onEdit = { editingTask = task },
                        )
                    }
                }
            }
        }
    }
}

@Composable
internal fun TaskTableRow(
    task: com.velox.app.data.model.TaskDto,
    cleanedInterval: String,
    runningTaskName: String?,
    onRun: () -> Unit,
    onEdit: () -> Unit
) {
    val isRunning = task.running || runningTaskName == task.name

    fun formatDateTime(dt: String?): String {
        if (dt == null) return "Never"
        return try {
            val parts = dt.split("T")
            val dateParts = parts[0].split("-")
            val yyyy = dateParts[0]
            val mm = dateParts[1].toInt()
            val dd = dateParts[2].toInt()
            val timePart = parts[1].replace("Z", "").take(5)
            "$mm/$dd/$yyyy, $timePart"
        } catch (e: Exception) {
            dt
        }
    }

    val lastRun = formatDateTime(task.lastRun)
    val nextRun = formatDateTime(task.nextRun)

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        // Task
        Text(task.name, color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(2f))

        // Interval
        Row(modifier = Modifier.weight(1f), verticalAlignment = Alignment.CenterVertically) {
            Text(cleanedInterval, color = NetflixLightGray, fontSize = 13.sp)
            Spacer(modifier = Modifier.width(4.dp))
            IconButton(
                onClick = onEdit,
                modifier = Modifier.size(24.dp)
            ) {
                Icon(LucideIcons.Edit, contentDescription = "Edit Interval", tint = NetflixLightGray, modifier = Modifier.size(14.dp))
            }
        }

        // Timestamps
        Column(modifier = Modifier.weight(2f)) {
            Text("Last: $lastRun", color = NetflixLightGray, fontSize = 12.sp)
            Spacer(modifier = Modifier.height(2.dp))
            Text("Next: $nextRun", color = NetflixLightGray, fontSize = 12.sp)
        }

        // Status
        Box(modifier = Modifier.weight(1f), contentAlignment = Alignment.CenterStart) {
            if (isRunning) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    CircularProgressIndicator(modifier = Modifier.size(12.dp), color = Color(0xFFFACC15), strokeWidth = 1.5.dp)
                    Spacer(modifier = Modifier.width(6.dp))
                    Text("Running", color = Color(0xFFFACC15), fontSize = 12.sp)
                }
            } else {
                Text("Idle", color = Color.Gray, fontSize = 12.sp)
            }
        }

        // Action
        Box(modifier = Modifier.weight(1.5f), contentAlignment = Alignment.CenterStart) {
            Button(
                onClick = onRun,
                enabled = !isRunning,
                colors = ButtonDefaults.buttonColors(
                    containerColor = NetflixGray,
                    disabledContainerColor = NetflixGray.copy(alpha = 0.5f)
                ),
                contentPadding = PaddingValues(horizontal = 10.dp, vertical = 0.dp),
                shape = RoundedCornerShape(4.dp),
                modifier = Modifier.height(28.dp).defaultMinSize(minWidth = 1.dp)
            ) {
                Icon(if (isRunning) LucideIcons.Refresh else LucideIcons.PlayCircle, contentDescription = null, modifier = Modifier.size(12.dp), tint = NetflixWhite)
                Spacer(modifier = Modifier.width(4.dp))
                Text(if (isRunning) "Running..." else "Run Now", fontSize = 11.sp, color = NetflixWhite)
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun EditIntervalDialog(
    taskName: String,
    currentInterval: String,
    onDismiss: () -> Unit,
    onSave: (String) -> Unit
) {
    val presets = listOf("30m", "1h", "12h", "24h", "168h", "custom")
    val presetLabels = mapOf(
        "30m" to "Every 30 Minutes",
        "1h" to "Every 1 Hour",
        "12h" to "Every 12 Hours",
        "24h" to "Every 24 Hours",
        "168h" to "Every 7 Days",
        "custom" to "Custom..."
    )

    var selectedPreset by remember { mutableStateOf(if (presets.contains(currentInterval)) currentInterval else "custom") }
    var customValue by remember { mutableStateOf(if (!presets.contains(currentInterval)) currentInterval else "") }

    AlertDialog(
        onDismissRequest = onDismiss,
        containerColor = NetflixDark,
        title = {
            Text("Edit Schedule", color = NetflixWhite, fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
        },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                Text(taskName, color = NetflixLightGray, fontSize = 14.sp)

                SettingsDropdown(
                    label = "Interval",
                    value = selectedPreset,
                    options = presetLabels.map { it.key to it.value },
                    onValueChange = { selectedPreset = it }
                )

                if (selectedPreset == "custom") {
                    OutlinedTextField(
                        value = customValue,
                        onValueChange = { customValue = it },
                        label = { Text("Custom Value (e.g. 2h30m)") },
                        modifier = Modifier.fillMaxWidth(),
                        colors = textFieldColors(),
                        singleLine = true,
                    )
                }
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    val finalVal = if (selectedPreset == "custom") customValue else selectedPreset
                    if (finalVal.isNotBlank()) {
                        onSave(finalVal)
                    }
                },
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                shape = RoundedCornerShape(8.dp),
            ) {
                Text(stringResource(R.string.action_save), color = NetflixWhite)
            }
        },
        dismissButton = {
            TextButton(
                onClick = onDismiss,
                colors = ButtonDefaults.textButtonColors(contentColor = NetflixLightGray)
            ) {
                Text(stringResource(R.string.action_cancel))
            }
        }
    )
}
