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
fun UsersSectionRoute(
    viewModel: SystemAdminViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    var showAddUserDialog by remember { mutableStateOf(false) }
    var editingUser by remember { mutableStateOf<AdminUser?>(null) }

    LaunchedEffect(Unit) {
        viewModel.loadUsers()
    }

    if (showAddUserDialog || editingUser != null) {
        UserDialog(
            user = editingUser,
            onDismiss = {
                showAddUserDialog = false
                editingUser = null
            },
            onSave = { username, password, displayName, isAdmin ->
                if (editingUser != null) {
                    viewModel.updateUser(editingUser!!.id, displayName, isAdmin)
                } else {
                    viewModel.createUser(username, password, displayName, isAdmin)
                }
                showAddUserDialog = false
                editingUser = null
            },
        )
    }

    UsersSectionContent(
        uiState = uiState,
        onAddClick = { showAddUserDialog = true },
        onEditClick = { editingUser = it },
        onDeleteClick = { viewModel.deleteUser(it) },
    )
}

@Composable
internal fun UsersSectionContent(
    uiState: SystemAdminUiState,
    onAddClick: () -> Unit,
    onEditClick: (AdminUser) -> Unit,
    onDeleteClick: (Int) -> Unit,
) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column {
                Text(stringResource(R.string.settings_title_users), color = NetflixWhite, fontSize = 20.sp, fontWeight = FontWeight.Bold)
                val count = uiState.users.size
                Text("$count ${if (count == 1) "user" else "users"}", color = NetflixLightGray, fontSize = 14.sp)
            }
            Button(
                onClick = onAddClick,
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                shape = RoundedCornerShape(8.dp),
            ) {
                Icon(LucideIcons.Add, contentDescription = null, modifier = Modifier.size(16.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text(stringResource(R.string.action_add_user), fontWeight = FontWeight.SemiBold)
            }
        }

        if (uiState.isLoading) {
            CircularProgressIndicator(color = NetflixRed, modifier = Modifier.padding(16.dp))
        } else if (uiState.users.isEmpty()) {
            Surface(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(160.dp),
                color = NetflixDark,
                shape = RoundedCornerShape(8.dp),
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Text(stringResource(R.string.users_empty), color = NetflixLightGray, fontSize = 14.sp)
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
                        Text(stringResource(R.string.user_user), color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(2.5f))
                        Text(stringResource(R.string.user_role), color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.2f))
                        Text(stringResource(R.string.user_created), color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1.5f))
                        Text(stringResource(R.string.user_actions), color = NetflixLightGray, fontSize = 13.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(2f))
                    }

                    HorizontalDivider(color = Color.White.copy(alpha = 0.05f))

                    // Users List
                    uiState.users.forEachIndexed { index, user ->
                        if (index > 0) {
                            HorizontalDivider(color = Color.White.copy(alpha = 0.02f))
                        }
                        UserTableRow(
                            user = user,
                            canDelete = uiState.currentUserId != user.id,
                            onEdit = { onEditClick(user) },
                            onDelete = { onDeleteClick(user.id) },
                        )
                    }
                }
            }
        }
    }
}

@Composable
internal fun UserTableRow(
    user: AdminUser,
    canDelete: Boolean,
    onEdit: () -> Unit,
    onDelete: () -> Unit,
) {
    val displayName = user.displayName.ifBlank { user.username }
    val initial = displayName.take(1).uppercase()

    val rawDate = user.createdAt.substringBefore("T")
    val fmtDate = try {
        val parts = rawDate.split("-")
        if (parts.size == 3) "${parts[1].toInt()}/${parts[2].toInt()}/${parts[0]}" else rawDate
    } catch (e: Exception) { rawDate }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        // Col 1: User (weight 2.5f)
        Row(
            modifier = Modifier.weight(2.5f),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Surface(
                modifier = Modifier.size(32.dp),
                shape = RoundedCornerShape(16.dp),
                color = NetflixGray,
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Text(initial, color = NetflixWhite, fontWeight = FontWeight.Medium, fontSize = 14.sp)
                }
            }
            Spacer(modifier = Modifier.width(12.dp))
            Column {
                Text(displayName, color = NetflixWhite, fontSize = 14.sp, fontWeight = FontWeight.SemiBold)
                Text(user.username, color = NetflixLightGray, fontSize = 12.sp)
            }
        }

        // Col 2: Role (weight 1.2f)
        Box(modifier = Modifier.weight(1.2f), contentAlignment = Alignment.CenterStart) {
            val roleColor = if (user.isAdmin) Color(0xFFA855F7) else Color(0xFF3B82F6)
            val roleBg = if (user.isAdmin) Color(0x33A855F7) else Color(0x333B82F6)
            Surface(color = roleBg, shape = RoundedCornerShape(4.dp)) {
                Text(
                    text = if (user.isAdmin) "Admin" else "User",
                    color = roleColor,
                    fontSize = 11.sp,
                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
                    fontWeight = FontWeight.Medium
                )
            }
        }

        // Col 3: Created (weight 1.5f)
        Text(fmtDate, color = NetflixLightGray, fontSize = 13.sp, modifier = Modifier.weight(1.5f))

        // Col 4: Actions (weight 2f)
        Row(
            modifier = Modifier.weight(2f),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Button(
                onClick = onEdit,
                colors = ButtonDefaults.buttonColors(containerColor = NetflixGray),
                contentPadding = PaddingValues(horizontal = 10.dp, vertical = 0.dp),
                shape = RoundedCornerShape(4.dp),
                modifier = Modifier.height(28.dp).defaultMinSize(minWidth = 1.dp)
            ) {
                Text(stringResource(R.string.action_edit), fontSize = 11.sp, color = NetflixWhite)
            }

            if (canDelete) {
                Button(
                    onClick = onDelete,
                    colors = ButtonDefaults.buttonColors(containerColor = Color(0x33EF4444)), // Red 400 / 20%
                    contentPadding = PaddingValues(horizontal = 10.dp, vertical = 0.dp),
                    shape = RoundedCornerShape(4.dp),
                    modifier = Modifier.height(28.dp).defaultMinSize(minWidth = 1.dp)
                ) {
                    Text(stringResource(R.string.action_delete), fontSize = 11.sp, color = Color(0xFFF87171)) // Red 400
                }
            }
        }
    }
}

@Composable

internal fun UserDialog(
    user: AdminUser?,
    onDismiss: () -> Unit,
    onSave: (username: String, password: String, displayName: String, isAdmin: Boolean) -> Unit,
) {
    var username by remember { mutableStateOf(user?.username ?: "") }
    var password by remember { mutableStateOf("") }
    var displayName by remember { mutableStateOf(user?.displayName ?: "") }
    var isAdmin by remember { mutableStateOf(user?.isAdmin ?: false) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(if (user != null) "Edit User" else "Add User", color = NetflixWhite) },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                OutlinedTextField(
                    value = username,
                    onValueChange = { username = it },
                    label = { Text(stringResource(R.string.login_username)) },
                    modifier = Modifier.fillMaxWidth(),
                    colors = textFieldColors(),
                    enabled = user == null, // Can't change username when editing
                )
                if (user == null) {
                    OutlinedTextField(
                        value = password,
                        onValueChange = { password = it },
                        label = { Text(stringResource(R.string.login_password)) },
                        modifier = Modifier.fillMaxWidth(),
                        colors = textFieldColors(),
                    )
                }
                OutlinedTextField(
                    value = displayName,
                    onValueChange = { displayName = it },
                    label = { Text(stringResource(R.string.user_display_name)) },
                    modifier = Modifier.fillMaxWidth(),
                    colors = textFieldColors(),
                )
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(stringResource(R.string.user_admin), color = NetflixWhite)
                    Switch(
                        checked = isAdmin,
                        onCheckedChange = { isAdmin = it },
                        colors = SwitchDefaults.colors(checkedThumbColor = NetflixRed, checkedTrackColor = NetflixRed),
                    )
                }
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    if (username.isNotBlank() && displayName.isNotBlank() && (user != null || password.isNotBlank())) {
                        onSave(username, password, displayName, isAdmin)
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
