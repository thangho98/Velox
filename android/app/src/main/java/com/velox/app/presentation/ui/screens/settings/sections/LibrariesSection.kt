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
fun LibrariesSectionRoute(
    viewModel: SystemAdminViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    
    var showAddLibraryDialog by remember { mutableStateOf(false) }
    var editingLibrary by remember { mutableStateOf<com.velox.app.domain.model.Library?>(null) }
    
    LaunchedEffect(Unit) {
        viewModel.loadLibraries()
    }
    
    if (showAddLibraryDialog || editingLibrary != null) {
        LibraryDialog(
            library = editingLibrary,
            onDismiss = {
                showAddLibraryDialog = false
                editingLibrary = null
            },
            onSave = { name, type, paths ->
                if (editingLibrary != null) {
                    viewModel.updateLibrary(editingLibrary!!.id, name, paths)
                } else {
                    viewModel.createLibrary(name, type, paths)
                }
                showAddLibraryDialog = false
                editingLibrary = null
            },
        )
    }
    
    LibrariesSectionContent(
        uiState = uiState,
        onAddClick = { showAddLibraryDialog = true },
        onEditClick = { editingLibrary = it },
        onDeleteClick = { viewModel.deleteLibrary(it) },
        onScanClick = { id, force -> viewModel.scanLibrary(id, force) }
    )
}

@Composable
internal fun LibrariesSectionContent(
    uiState: SystemAdminUiState,
    onAddClick: () -> Unit,
    onEditClick: (com.velox.app.domain.model.Library) -> Unit,
    onDeleteClick: (Int) -> Unit,
    onScanClick: (Int, Boolean) -> Unit,
) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column {
                Text(stringResource(R.string.settings_title_libraries), color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
                val count = uiState.libraries.size
                Text("$count ${if (count == 1) "library" else "libraries"} configured", color = NetflixLightGray, fontSize = 14.sp)
            }
            Button(
                onClick = onAddClick,
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                shape = RoundedCornerShape(8.dp),
            ) {
                Icon(LucideIcons.Add, contentDescription = null, modifier = Modifier.size(16.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text(stringResource(R.string.action_add_library), fontWeight = FontWeight.SemiBold)
            }
        }

        if (uiState.isLoading) {
            CircularProgressIndicator(color = NetflixRed, modifier = Modifier.padding(16.dp))
        } else if (uiState.libraries.isEmpty()) {
            Surface(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(160.dp),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    Icon(LucideIcons.Movie, contentDescription = null, tint = NetflixGray, modifier = Modifier.size(36.dp))
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(stringResource(R.string.libraries_empty), color = NetflixLightGray, fontSize = 14.sp)
                    Spacer(modifier = Modifier.height(12.dp))
                    Button(
                        onClick = onAddClick,
                        colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                        shape = RoundedCornerShape(8.dp),
                    ) {
                        Text(stringResource(R.string.action_add_library), fontWeight = FontWeight.Medium)
                    }
                }
            }
        } else {
            Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                uiState.libraries.forEach { library ->
                    LibraryCard(
                        library = library,
                        isScanning = uiState.scanningLibraryId == library.id,
                        onScan = { onScanClick(library.id, false) },
                        onForceScan = { onScanClick(library.id, true) },
                        onDelete = { onDeleteClick(library.id) }
                    )
                }
            }
        }
    }
}

@Composable
internal fun LibraryCard(
    library: com.velox.app.domain.model.Library,
    isScanning: Boolean,
    onScan: () -> Unit,
    onForceScan: () -> Unit,
    onDelete: () -> Unit,
) {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = NetflixDark,
        shape = RoundedCornerShape(8.dp),
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            // Left Content
            Row(
                modifier = Modifier.weight(1f),
                verticalAlignment = Alignment.CenterVertically
            ) {
                val typeDetails = when(library.type.lowercase(java.util.Locale.US)) {
                    "movies" -> Triple(LucideIcons.Movie, Color(0xFF3B82F6), Color(0xFF1E3A8A)) // Blue
                    "tvshows" -> Triple(LucideIcons.Tv, Color(0xFFA855F7), Color(0xFF4C1D95)) // Purple
                    else -> Triple(LucideIcons.ListIcon, Color(0xFF22C55E), Color(0xFF14532D)) // Green
                }

                // Icon Box
                Surface(
                    modifier = Modifier.size(40.dp),
                    shape = RoundedCornerShape(8.dp),
                    color = typeDetails.third.copy(alpha = 0.4f)
                ) {
                    Box(contentAlignment = Alignment.Center) {
                        Icon(typeDetails.first, contentDescription = null, tint = typeDetails.second, modifier = Modifier.size(20.dp))
                    }
                }

                Spacer(modifier = Modifier.width(12.dp))

                Column {
                    Text(library.name, color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.SemiBold)
                    Spacer(modifier = Modifier.height(2.dp))
                    library.paths.takeTop(3).forEach { p ->
                        Text(p, color = NetflixLightGray, fontSize = 12.sp, fontFamily = androidx.compose.ui.text.font.FontFamily.Monospace, maxLines = 1)
                    }
                    Spacer(modifier = Modifier.height(4.dp))
                    val prettyType = when(library.type.lowercase(java.util.Locale.US)) {
                        "movies" -> "Movies"
                        "tvshows" -> "TV Shows"
                        else -> "Mixed"
                    }
                    Surface(
                        shape = RoundedCornerShape(4.dp),
                        color = Color.Transparent,
                        border = androidx.compose.foundation.BorderStroke(1.dp, typeDetails.second)
                    ) {
                        Text(prettyType, modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp), color = typeDetails.second, fontSize = 10.sp)
                    }
                }
            }

            // Right Buttons
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                LibraryCardAction(
                    text = "Scan",
                    icon = LucideIcons.Refresh,
                    onClick = onScan,
                    isLoading = isScanning,
                    disabled = isScanning,
                )

                LibraryCardAction(
                    text = "Force Rescan",
                    icon = LucideIcons.Refresh,
                    onClick = onForceScan,
                    disabled = isScanning,
                )

                LibraryCardAction(
                    text = null,
                    icon = LucideIcons.Delete,
                    onClick = onDelete,
                    disabled = isScanning,
                )
            }
        }
    }
}

@Composable
internal fun LibraryCardAction(
    text: String?,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    onClick: () -> Unit,
    isLoading: Boolean = false,
    disabled: Boolean = false,
) {
    Surface(
        shape = RoundedCornerShape(4.dp),
        color = NetflixGray,
        modifier = Modifier
            .height(32.dp)
            .clip(RoundedCornerShape(4.dp))
            .clickable(onClick = onClick, enabled = !disabled)
    ) {
        Row(
            modifier = Modifier.padding(horizontal = if (text != null) 12.dp else 10.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.Center
        ) {
            if (isLoading) {
                CircularProgressIndicator(
                    color = NetflixWhite,
                    strokeWidth = 2.dp,
                    modifier = Modifier.size(13.dp)
                )
            } else {
                Icon(icon, contentDescription = null, modifier = Modifier.size(14.dp), tint = if (disabled) NetflixLightGray else NetflixWhite)
            }

            if (text != null) {
                Spacer(modifier = Modifier.width(6.dp))
                val label = if (isLoading && text == "Scan") "Scanning" else text
                Text(
                    text = label,
                    fontSize = 12.sp,
                    color = if (disabled && !isLoading) NetflixLightGray else NetflixWhite,
                    fontWeight = FontWeight.Medium
                )
            }
        }
    }
}

internal fun <T> List<T>.takeTop(n: Int): List<T> = if (size > n) take(n) else this

@Composable

internal fun LibraryDialog(
    library: Library?,
    onDismiss: () -> Unit,
    onSave: (name: String, type: String, paths: List<String>) -> Unit,
) {
    var name by remember { mutableStateOf(library?.name ?: "") }
    var type by remember { mutableStateOf(library?.type ?: "movie") }
    var paths by remember { mutableStateOf(library?.paths?.joinToString(",") ?: "") }
    var typeExpanded by remember { mutableStateOf(false) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(if (library != null) "Edit Library" else "Add Library", color = NetflixWhite) },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                OutlinedTextField(
                    value = name,
                    onValueChange = { name = it },
                    label = { Text(stringResource(R.string.library_name)) },
                    modifier = Modifier.fillMaxWidth(),
                    colors = textFieldColors(),
                )
                Box {
                    OutlinedTextField(
                        value = type.replaceFirstChar { it.uppercase() },
                        onValueChange = {},
                        label = { Text(stringResource(R.string.library_type)) },
                        modifier = Modifier.fillMaxWidth(),
                        colors = textFieldColors(),
                        readOnly = true,
                        trailingIcon = {
                            IconButton(onClick = { typeExpanded = true }) {
                                Icon(LucideIcons.ArrowDropDown, contentDescription = null)
                            }
                        },
                    )
                    DropdownMenu(
                        expanded = typeExpanded,
                        onDismissRequest = { typeExpanded = false },
                    ) {
                        listOf("movie", "series").forEach { t ->
                            DropdownMenuItem(
                                text = { Text(t.replaceFirstChar { it.uppercase() }) },
                                onClick = { type = t; typeExpanded = false },
                            )
                        }
                    }
                }
                OutlinedTextField(
                    value = paths,
                    onValueChange = { paths = it },
                    label = { Text(stringResource(R.string.library_paths)) },
                    modifier = Modifier.fillMaxWidth(),
                    colors = textFieldColors(),
                )
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    if (name.isNotBlank() && paths.isNotBlank()) {
                        onSave(name, type, paths.split(",").map { it.trim() }.filter { it.isNotEmpty() })
                    }
                },
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
            ) {
                Text(stringResource(R.string.action_save))
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
