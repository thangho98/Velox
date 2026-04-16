package com.velox.app.presentation.ui.screens.settings.components

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
import com.velox.app.presentation.viewmodel.SettingsAction
import com.velox.app.presentation.viewmodel.SettingsSection
import com.velox.app.ui.theme.*
import com.velox.app.ui.theme.TextMuted
import java.net.HttpURLConnection
import java.net.URL
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.json.JSONObject

@Composable
internal fun SettingsToggle(
    label: String,
    description: String,
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
) {
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
                Text(label, color = NetflixWhite)
                Text(description, color = NetflixLightGray, fontSize = 12.sp)
            }
            Switch(
                checked = checked,
                onCheckedChange = onCheckedChange,
                colors = SwitchDefaults.colors(checkedThumbColor = NetflixRed, checkedTrackColor = NetflixRed),
            )
        }
    }
}

@OptIn(androidx.compose.material3.ExperimentalMaterial3Api::class)
@Composable
internal fun SettingsDropdown(
    label: String,
    value: String,
    options: List<Pair<String, String>>,
    onValueChange: (String) -> Unit,
) {
    var expanded by remember { mutableStateOf(false) }
    val displayValue = options.find { it.first == value }?.second ?: value

    Column {
        Text(label, color = NetflixWhite, fontSize = 14.sp)
        Spacer(modifier = Modifier.height(4.dp))
        Box(modifier = Modifier.fillMaxWidth()) {
            OutlinedTextField(
                value = displayValue,
                onValueChange = {},
                readOnly = true,
                modifier = Modifier.fillMaxWidth(),
                colors = textFieldColors(),
                shape = RoundedCornerShape(8.dp),
                trailingIcon = {
                    Icon(LucideIcons.ArrowDropDown, contentDescription = null, tint = NetflixWhite)
                },
            )
            // Transparent touch interceptor overlay
            Surface(
                modifier = Modifier
                    .matchParentSize()
                    .clickable { expanded = true },
                color = androidx.compose.ui.graphics.Color.Transparent
            ) {}
        }

        if (expanded) {
            ModalBottomSheet(
                onDismissRequest = { expanded = false },
                containerColor = NetflixDark,
                dragHandle = { BottomSheetDefaults.DragHandle() }
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(bottom = 32.dp)
                ) {
                    Text(
                        text = "Select $label",
                        color = NetflixLightGray,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.SemiBold,
                        modifier = Modifier.padding(horizontal = 24.dp, vertical = 8.dp)
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    options.forEach { option ->
                        val isSelected = option.first == value
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .clickable {
                                    onValueChange(option.first)
                                    expanded = false
                                }
                                .padding(horizontal = 24.dp, vertical = 16.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            Text(
                                text = option.second,
                                color = if (isSelected) NetflixWhite else NetflixLightGray,
                                fontSize = 16.sp,
                                fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal
                            )
                            if (isSelected) {
                                Icon(
                                    imageVector = LucideIcons.Check,
                                    contentDescription = "Selected",
                                    tint = NetflixRed,
                                    modifier = Modifier.size(20.dp)
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
internal fun textFieldColors() = OutlinedTextFieldDefaults.colors(
    focusedTextColor = NetflixWhite,
    unfocusedTextColor = NetflixWhite,
    focusedBorderColor = NetflixRed,
    unfocusedBorderColor = NetflixGray,
    focusedLabelColor = NetflixRed,
    unfocusedLabelColor = NetflixLightGray,
)

@Preview(showBackground = true)
@Composable
internal fun SettingsGroupCard(
    sections: List<SettingsSection>,
    onAction: (SettingsAction) -> Unit
) {
    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp),
        color = NetflixDark,
        shape = RoundedCornerShape(16.dp)
    ) {
        Column {
            sections.forEachIndexed { index, section ->
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { onAction(SettingsAction.SelectSection(section)) }
                        .padding(horizontal = 16.dp, vertical = 16.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(
                        imageVector = iconForSection(section),
                        contentDescription = null,
                        tint = NetflixLightGray,
                        modifier = Modifier.size(22.dp)
                    )
                    Spacer(modifier = Modifier.width(16.dp))
                    Text(section.title, color = NetflixWhite, fontSize = 16.sp, fontWeight = FontWeight.Medium)
                    Spacer(Modifier.weight(1f))
                    Icon(LucideIcons.ChevronRight, contentDescription = null, tint = NetflixGray, modifier = Modifier.size(20.dp))
                }

                if (index < sections.size - 1) {
                    androidx.compose.material3.HorizontalDivider(
                        color = NetflixBlack,
                        modifier = Modifier.padding(start = 54.dp)
                    )
                }
            }
        }
    }
}

internal fun iconForSection(section: SettingsSection): androidx.compose.ui.graphics.vector.ImageVector {
    return when (section) {
        SettingsSection.PROFILE -> LucideIcons.Person
        SettingsSection.PREFERENCES -> LucideIcons.Settings
        SettingsSection.SECURITY -> LucideIcons.Lock
        SettingsSection.SESSIONS -> LucideIcons.Devices
        SettingsSection.METADATA -> LucideIcons.Info
        SettingsSection.SUBTITLES -> LucideIcons.Subtitles
        SettingsSection.PLAYBACK -> LucideIcons.PlayCircle
        SettingsSection.CINEMA -> LucideIcons.Movie
        SettingsSection.PRETRANSCODE -> LucideIcons.FlashOn
        SettingsSection.MARKERS -> LucideIcons.SkipNext
        SettingsSection.DASHBOARD -> LucideIcons.GridView
        SettingsSection.LIBRARIES -> LucideIcons.Folder
        SettingsSection.USERS -> LucideIcons.Person
        SettingsSection.ACTIVITY -> LucideIcons.ShowChart
        SettingsSection.TASKS -> LucideIcons.CheckCircle
        SettingsSection.WEBHOOKS -> LucideIcons.Link
    }
}

