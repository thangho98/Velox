package com.velox.app.presentation.ui.screens.settings

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.compose.ui.res.stringResource
import com.velox.app.R
import com.velox.app.domain.model.Library
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.viewmodel.*
import com.velox.app.ui.theme.*
import com.velox.app.presentation.ui.screens.settings.sections.*
import com.velox.app.presentation.ui.screens.settings.components.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(
    onBackClick: () -> Unit,
    onProfileClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    profileViewModel: UserProfileViewModel = hiltViewModel(),
    systemAdminViewModel: SystemAdminViewModel = hiltViewModel(),
    mediaSettingsViewModel: MediaSettingsViewModel = hiltViewModel(),
) {
    var selectedTab by remember { mutableStateOf<SettingsSection?>(null) }
    
    val profileState by profileViewModel.uiState.collectAsStateWithLifecycle()
    val adminState by systemAdminViewModel.uiState.collectAsStateWithLifecycle()
    val mediaState by mediaSettingsViewModel.uiState.collectAsStateWithLifecycle()

    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(profileState.error, adminState.error, mediaState.error) {
        val err = profileState.error ?: adminState.error ?: mediaState.error
        err?.let { 
            snackbarHostState.showSnackbar(it) 
            profileViewModel.clearMessages()
            systemAdminViewModel.clearMessages()
            mediaSettingsViewModel.clearMessages()
        }
    }

    LaunchedEffect(profileState.successMessage, adminState.successMessage, mediaState.successMessage) {
        val msg = profileState.successMessage ?: adminState.successMessage ?: mediaState.successMessage
        msg?.let { 
            snackbarHostState.showSnackbar(it) 
            profileViewModel.clearMessages()
            systemAdminViewModel.clearMessages()
            mediaSettingsViewModel.clearMessages()
        }
    }

    Scaffold(
        snackbarHost = { SnackbarHost(hostState = snackbarHostState) },
        topBar = {
            TopAppBar(
                title = {
                    val titleText = selectedTab?.let { stringResource(it.titleRes) } ?: stringResource(R.string.settings_title)
                    Text(titleText, color = NetflixWhite, fontWeight = FontWeight.Bold)
                },
                navigationIcon = {
                    IconButton(onClick = {
                        if (selectedTab == null) onBackClick()
                        else selectedTab = null
                    }) {
                        Icon(LucideIcons.ChevronLeft, contentDescription = "Back", tint = NetflixWhite)
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = NetflixBlack),
            )
        },
        containerColor = NetflixBlack,
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
        ) {
            profileState.user?.let { user ->
                UserCard(
                    displayName = profileState.displayName,
                    username = profileState.username,
                    isAdmin = profileState.isAdmin,
                    profilePath = user.profilePath,
                    profile = user.profile,
                    onClick = onProfileClick,
                )
            }

            if (selectedTab == null) {
                LazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    contentPadding = PaddingValues(bottom = 32.dp)
                ) {
                    item {
                        Text(
                            text = stringResource(R.string.settings_group_web),
                            color = NetflixLightGray,
                            fontSize = 12.sp,
                            fontWeight = FontWeight.Bold,
                            modifier = Modifier.padding(start = 24.dp, top = 24.dp, bottom = 8.dp)
                        )
                        val webSettings = listOf(SettingsSection.PROFILE, SettingsSection.PREFERENCES, SettingsSection.SECURITY, SettingsSection.SESSIONS)
                        SettingsGroupCard(sections = webSettings, onAction = { action ->
                            if (action is SettingsAction.SelectSection) selectedTab = action.section
                        })
                    }

                    if (profileState.isAdmin) {
                        item {
                            Text(
                                text = stringResource(R.string.settings_group_admin),
                                color = NetflixLightGray,
                                fontSize = 12.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.padding(start = 24.dp, top = 24.dp, bottom = 8.dp)
                            )
                            val adminPrefs = listOf(SettingsSection.METADATA, SettingsSection.SUBTITLES, SettingsSection.PLAYBACK, SettingsSection.CINEMA, SettingsSection.PRETRANSCODE, SettingsSection.MARKERS)
                            SettingsGroupCard(sections = adminPrefs, onAction = { action ->
                                if (action is SettingsAction.SelectSection) selectedTab = action.section
                            })
                        }

                        item {
                            Text(
                                text = stringResource(R.string.settings_group_server),
                                color = NetflixLightGray,
                                fontSize = 12.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.padding(start = 24.dp, top = 24.dp, bottom = 8.dp)
                            )
                            val serverPrefs = listOf(SettingsSection.DASHBOARD, SettingsSection.LIBRARIES, SettingsSection.USERS, SettingsSection.ACTIVITY, SettingsSection.TASKS, SettingsSection.WEBHOOKS)
                            SettingsGroupCard(sections = serverPrefs, onAction = { action ->
                                if (action is SettingsAction.SelectSection) selectedTab = action.section
                            })
                        }
                    }

                    item {
                        FeedbackCard()
                    }

                    item {
                        AppVersionCard(uiState = adminState, onCheckUpdate = { systemAdminViewModel.checkAppUpdate() })
                    }
                }
            } else {
                LazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(16.dp),
                ) {
                    item {
                        when (selectedTab) {
                            SettingsSection.PROFILE -> ProfileSectionRoute(profileViewModel)
                            SettingsSection.PREFERENCES -> PreferencesSectionRoute(profileViewModel)
                            SettingsSection.SECURITY -> SecuritySectionRoute(profileViewModel)
                            SettingsSection.SESSIONS -> SessionsSectionRoute(profileViewModel)
                            SettingsSection.METADATA -> MetadataSectionRoute(mediaSettingsViewModel)
                            SettingsSection.SUBTITLES -> SubtitlesSectionRoute(mediaSettingsViewModel)
                            SettingsSection.PLAYBACK -> PlaybackSectionRoute(mediaSettingsViewModel)
                            SettingsSection.CINEMA -> CinemaSectionRoute(mediaSettingsViewModel)
                            SettingsSection.PRETRANSCODE -> PretranscodeSectionRoute(mediaSettingsViewModel)
                            SettingsSection.MARKERS -> MarkersSectionRoute(mediaSettingsViewModel)
                            SettingsSection.DASHBOARD -> DashboardSectionRoute(systemAdminViewModel)
                            SettingsSection.LIBRARIES -> LibrariesSectionRoute(systemAdminViewModel)
                            SettingsSection.USERS -> UsersSectionRoute(systemAdminViewModel)
                            SettingsSection.ACTIVITY -> ActivitySectionRoute(systemAdminViewModel)
                            SettingsSection.TASKS -> TasksSectionRoute(systemAdminViewModel)
                            SettingsSection.WEBHOOKS -> WebhooksSectionRoute(systemAdminViewModel)
                            null -> {}
                        }
                    }
                }
            }
        }
    }
}
