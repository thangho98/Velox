package com.velox.app.presentation.ui.screens.auth

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Clear
import androidx.compose.material.icons.filled.Cloud
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Visibility
import androidx.compose.material.icons.filled.VisibilityOff
import androidx.compose.material.icons.filled.Wifi
import androidx.compose.material.icons.filled.WifiOff
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusDirection
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.velox.app.R
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.viewmodel.ConnectionInfo
import com.velox.app.presentation.viewmodel.ConnectionStatus
import com.velox.app.presentation.viewmodel.LoginUiState
import com.velox.app.presentation.viewmodel.LoginViewModel
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixGray
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme
import kotlinx.coroutines.launch

@Composable
fun LoginScreen(
    onLoginSuccess: () -> Unit,
    viewModel: LoginViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val connectionInfo by viewModel.connectionInfo.collectAsStateWithLifecycle()

    LoginScreenContent(
        uiState = uiState,
        connectionInfo = connectionInfo,
        initialServerUrl = remember { viewModel.getSavedServerUrl() },
        onVerifyServerUrl = { viewModel.verifyServerUrl(it) },
        onLogin = { username, password -> viewModel.login(username, password) },
        onLoginSuccess = onLoginSuccess,
        onSaveServerUrl = { viewModel.saveServerUrl(it) },
        onCheckConnection = { viewModel.checkServerConnection() },
        onClearConnectionStatus = { viewModel.clearConnectionStatus() }
    )
}

@Composable
private fun LoginScreenContent(
    uiState: LoginUiState,
    connectionInfo: ConnectionInfo,
    initialServerUrl: String,
    onVerifyServerUrl: suspend (String) -> Boolean,
    onLogin: (String, String) -> Unit,
    onLoginSuccess: () -> Unit,
    onSaveServerUrl: suspend (String) -> Unit,
    onCheckConnection: () -> Unit,
    onClearConnectionStatus: () -> Unit,
    initialStep: LoginStep = LoginStep.SERVER_URL,
) {
    val focusManager = LocalFocusManager.current
    val scope = rememberCoroutineScope()

    // Two-step flow state
    var currentStep by remember { mutableStateOf(initialStep) }
    var serverUrl by remember { mutableStateOf(initialServerUrl) }
    var username by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var passwordVisible by remember { mutableStateOf(false) }
    var serverUrlError by remember { mutableStateOf<String?>(null) }
    var isVerifyingServer by remember { mutableStateOf(false) }

    // Save server URL when step changes to credentials
    LaunchedEffect(currentStep) {
        if (currentStep == LoginStep.CREDENTIALS && serverUrl.isNotBlank()) {
            onSaveServerUrl(serverUrl)
            // Auto-check connection when entering credentials step
            onCheckConnection()
        }
    }

    LaunchedEffect(uiState) {
        if (uiState is LoginUiState.Success) {
            onLoginSuccess()
        }
    }

    // Clear connection status when changing server URL
    LaunchedEffect(serverUrl) {
        if (currentStep == LoginStep.CREDENTIALS) {
            onClearConnectionStatus()
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(NetflixBlack)
            .systemBarsPadding(),
    ) {
        // Back button - only shown on CREDENTIALS step
        if (currentStep == LoginStep.CREDENTIALS) {
            IconButton(
                onClick = { currentStep = LoginStep.SERVER_URL },
                modifier = Modifier
                    .align(Alignment.TopStart)
                    .padding(start = 8.dp, top = 8.dp)
            ) {
                Icon(
                    imageVector = LucideIcons.ChevronLeft,
                    contentDescription = "Back",
                    tint = NetflixRed,
                )
            }
        }

        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp),
            contentAlignment = Alignment.Center,
        ) {
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                modifier = Modifier.widthIn(max = 400.dp),
            ) {
                // Logo
                val screenWidth = LocalConfiguration.current.screenWidthDp
                Text(
                    text = "VELOX",
                    fontSize = if (screenWidth < 600) 36.sp else 48.sp,
                    fontWeight = FontWeight.Bold,
                    color = NetflixRed,
                    textAlign = TextAlign.Center,
                )

                Spacer(modifier = Modifier.height(48.dp))

                when (currentStep) {
                    LoginStep.SERVER_URL -> {
                        // Step 1: Server URL
                        ServerUrlStep(
                            serverUrl = serverUrl,
                            onServerUrlChange = {
                                serverUrl = it
                                serverUrlError = null
                            },
                            serverUrlError = serverUrlError,
                            isVerifying = isVerifyingServer,
                            onContinue = {
                                if (serverUrl.isBlank()) {
                                    serverUrlError = "Please enter server URL"
                                    return@ServerUrlStep
                                }

                                var requestUrl = serverUrl.trim()
                                if (!requestUrl.startsWith("http://") && !requestUrl.startsWith("https://")) {
                                    requestUrl = "http://$requestUrl"
                                    serverUrl = requestUrl
                                }

                                isVerifyingServer = true
                                serverUrlError = null

                                scope.launch {
                                    val result = onVerifyServerUrl(requestUrl)
                                    isVerifyingServer = false
                                    if (result) {
                                        currentStep = LoginStep.CREDENTIALS
                                    } else {
                                        serverUrlError = "Cannot connect to server. Please check the URL."
                                    }
                                }
                            },
                        )
                    }

                    LoginStep.CREDENTIALS -> {
                        // Step 2: Credentials
                        CredentialsStep(
                            serverUrl = serverUrl,
                            username = username,
                            onUsernameChange = { username = it },
                            password = password,
                            onPasswordChange = { password = it },
                            passwordVisible = passwordVisible,
                            onPasswordVisibilityToggle = { passwordVisible = !passwordVisible },
                            isLoading = uiState is LoginUiState.Loading,
                            connectionInfo = connectionInfo,
                            errorMessage = if (uiState is LoginUiState.Error) {
                                uiState.message
                            } else null,
                            onSignIn = {
                                focusManager.clearFocus()
                                onLogin(username, password)
                            },
                            onChangeServer = { currentStep = LoginStep.SERVER_URL },
                            onRetryConnection = onCheckConnection,
                        )
                    }
                }
            }
        }
    }
}

private enum class LoginStep {
    SERVER_URL,
    CREDENTIALS,
}

@Composable
private fun ServerUrlStep(
    serverUrl: String,
    onServerUrlChange: (String) -> Unit,
    serverUrlError: String?,
    isVerifying: Boolean,
    onContinue: () -> Unit,
) {
    val focusManager = LocalFocusManager.current

    // Title
    val screenWidth = LocalConfiguration.current.screenWidthDp
    Text(
        text = stringResource(R.string.login_title_connect),
        fontSize = if (screenWidth < 600) 22.sp else 28.sp,
        fontWeight = FontWeight.SemiBold,
        color = NetflixWhite,
    )

    Spacer(modifier = Modifier.height(8.dp))

    Text(
        text = stringResource(R.string.login_subtitle_connect),
        fontSize = 14.sp,
        color = NetflixLightGray,
    )

    Spacer(modifier = Modifier.height(32.dp))

    // Server URL field
    OutlinedTextField(
        value = serverUrl,
        onValueChange = onServerUrlChange,
        label = { Text(stringResource(R.string.login_server_url)) },
        placeholder = { Text(stringResource(R.string.login_server_url_hint)) },
        singleLine = true,
        modifier = Modifier.fillMaxWidth(),
        leadingIcon = {
            Icon(
                imageVector = LucideIcons.Cloud,
                contentDescription = null,
                tint = NetflixLightGray,
            )
        },
        trailingIcon = {
            if (serverUrl.isNotBlank()) {
                IconButton(onClick = { onServerUrlChange("") }) {
                    Icon(
                        imageVector = LucideIcons.Close,
                        contentDescription = "Clear URL",
                        tint = NetflixLightGray,
                    )
                }
            }
        },
        isError = serverUrlError != null,
        supportingText = {
            if (serverUrlError != null) {
                Text(
                    text = serverUrlError,
                    color = MaterialTheme.colorScheme.error,
                )
            } else {
                Text(
                    text = stringResource(R.string.login_no_api_suffix),
                    color = NetflixLightGray,
                )
            }
        },
        colors = OutlinedTextFieldDefaults.colors(
            focusedBorderColor = NetflixRed,
            unfocusedBorderColor = Color.Gray,
            focusedLabelColor = NetflixRed,
            unfocusedLabelColor = Color.Gray,
            cursorColor = NetflixRed,
            focusedTextColor = NetflixWhite,
            unfocusedTextColor = NetflixWhite,
            focusedLeadingIconColor = NetflixRed,
            unfocusedLeadingIconColor = NetflixLightGray,
            focusedTrailingIconColor = NetflixRed,
            unfocusedTrailingIconColor = NetflixLightGray,
        ),
        keyboardOptions = KeyboardOptions(
            keyboardType = KeyboardType.Uri,
            imeAction = ImeAction.Done,
        ),
        keyboardActions = KeyboardActions(
            onDone = {
                focusManager.clearFocus()
                onContinue()
            },
        ),
    )

    Spacer(modifier = Modifier.height(32.dp))

    // Continue button
    Button(
        onClick = onContinue,
        modifier = Modifier
            .fillMaxWidth()
            .height(48.dp),
        enabled = serverUrl.isNotBlank() && !isVerifying,
        colors = ButtonDefaults.buttonColors(
            containerColor = NetflixRed,
            contentColor = NetflixWhite,
        ),
    ) {
        if (isVerifying) {
            CircularProgressIndicator(
                modifier = Modifier.size(24.dp),
                color = NetflixWhite,
            )
        } else {
            Text(
                text = stringResource(R.string.login_continue),
                fontSize = 16.sp,
                fontWeight = FontWeight.SemiBold,
            )
        }
    }

    Spacer(modifier = Modifier.height(32.dp))

    // Footer
    Text(
        text = stringResource(R.string.login_contact_admin),
        fontSize = 12.sp,
        color = NetflixGray,
        textAlign = TextAlign.Center,
    )
}

@Composable
private fun CredentialsStep(
    serverUrl: String,
    username: String,
    onUsernameChange: (String) -> Unit,
    password: String,
    onPasswordChange: (String) -> Unit,
    passwordVisible: Boolean,
    onPasswordVisibilityToggle: () -> Unit,
    isLoading: Boolean,
    connectionInfo: ConnectionInfo,
    errorMessage: String?,
    onSignIn: () -> Unit,
    onChangeServer: () -> Unit,
    onRetryConnection: () -> Unit,
) {
    val focusManager = LocalFocusManager.current

    // Title
    val screenWidth = LocalConfiguration.current.screenWidthDp
    Text(
        text = stringResource(R.string.login_sign_in),
        fontSize = if (screenWidth < 600) 22.sp else 28.sp,
        fontWeight = FontWeight.SemiBold,
        color = NetflixWhite,
    )

    Spacer(modifier = Modifier.height(8.dp))

    // Server info chip with connection status
    Surface(
        shape = androidx.compose.foundation.shape.RoundedCornerShape(16.dp),
        color = NetflixDark,
        modifier = Modifier.clickable { onChangeServer() },
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(
                imageVector = LucideIcons.Cloud,
                contentDescription = null,
                tint = NetflixLightGray,
                modifier = Modifier.size(16.dp),
            )
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                text = serverUrl.removeSuffix("/api/"),
                color = NetflixLightGray,
                fontSize = 12.sp,
                modifier = Modifier.weight(1f),
            )
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                text = stringResource(R.string.login_change),
                color = NetflixRed,
                fontSize = 12.sp,
                fontWeight = FontWeight.SemiBold,
            )
        }
    }

    Spacer(modifier = Modifier.height(16.dp))

    // Connection status banner
    ConnectionStatusBanner(
        connectionInfo = connectionInfo,
        onRetry = onRetryConnection,
    )

    // Error message (only show if not showing connection error)
    if (errorMessage != null && connectionInfo.status != ConnectionStatus.FAILED) {
        Spacer(modifier = Modifier.height(12.dp))
        Text(
            text = errorMessage,
            color = MaterialTheme.colorScheme.error,
            fontSize = 14.sp,
        )
    }

    // Connection failed - show detailed guidance
    if (connectionInfo.status == ConnectionStatus.FAILED && connectionInfo.errorMessage != null) {
        Spacer(modifier = Modifier.height(12.dp))
        Surface(
            shape = androidx.compose.foundation.shape.RoundedCornerShape(8.dp),
            color = NetflixDark,
        ) {
            Column(
                modifier = Modifier.padding(12.dp)
            ) {
                Text(
                    text = connectionInfo.errorMessage,
                    color = NetflixRed,
                    fontSize = 14.sp,
                )
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = "• Check that the server is running\n• Verify the URL is correct\n• Make sure you're on the same network",
                    color = NetflixLightGray,
                    fontSize = 12.sp,
                    lineHeight = 18.sp,
                )
                Spacer(modifier = Modifier.height(12.dp))
                Button(
                    onClick = onRetryConnection,
                    modifier = Modifier.height(36.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = NetflixRed,
                        contentColor = NetflixWhite,
                    ),
                ) {
                    Icon(
                        imageVector = LucideIcons.Refresh,
                        contentDescription = null,
                        modifier = Modifier.size(16.dp),
                    )
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = stringResource(R.string.login_retry_connection),
                        fontSize = 13.sp,
                        fontWeight = FontWeight.SemiBold,
                    )
                }
            }
        }
    }

    Spacer(modifier = Modifier.height(24.dp))

    // Username field
    OutlinedTextField(
        value = username,
        onValueChange = onUsernameChange,
        label = { Text(stringResource(R.string.login_username)) },
        singleLine = true,
        modifier = Modifier.fillMaxWidth(),
        colors = OutlinedTextFieldDefaults.colors(
            focusedBorderColor = NetflixRed,
            unfocusedBorderColor = Color.Gray,
            focusedLabelColor = NetflixRed,
            unfocusedLabelColor = Color.Gray,
            cursorColor = NetflixRed,
            focusedTextColor = NetflixWhite,
            unfocusedTextColor = NetflixWhite,
        ),
        keyboardOptions = KeyboardOptions(
            keyboardType = KeyboardType.Text,
            imeAction = ImeAction.Next,
        ),
        keyboardActions = KeyboardActions(
            onNext = { focusManager.moveFocus(FocusDirection.Down) },
        ),
    )

    Spacer(modifier = Modifier.height(16.dp))

    // Password field
    OutlinedTextField(
        value = password,
        onValueChange = onPasswordChange,
        label = { Text(stringResource(R.string.login_password)) },
        singleLine = true,
        modifier = Modifier.fillMaxWidth(),
        colors = OutlinedTextFieldDefaults.colors(
            focusedBorderColor = NetflixRed,
            unfocusedBorderColor = Color.Gray,
            focusedLabelColor = NetflixRed,
            unfocusedLabelColor = Color.Gray,
            cursorColor = NetflixRed,
            focusedTextColor = NetflixWhite,
            unfocusedTextColor = NetflixWhite,
        ),
        visualTransformation = if (passwordVisible) {
            VisualTransformation.None
        } else {
            PasswordVisualTransformation()
        },
        trailingIcon = {
            IconButton(onClick = onPasswordVisibilityToggle) {
                Icon(
                    imageVector = if (passwordVisible) {
                        LucideIcons.VisibilityOff
                    } else {
                        LucideIcons.Visibility
                    },
                    contentDescription = if (passwordVisible) "Hide password" else "Show password",
                    tint = Color.Gray,
                )
            }
        },
        keyboardOptions = KeyboardOptions(
            keyboardType = KeyboardType.Password,
            imeAction = ImeAction.Done,
        ),
        keyboardActions = KeyboardActions(
            onDone = {
                focusManager.clearFocus()
                if (username.isNotBlank() && password.isNotBlank()) {
                    onSignIn()
                }
            },
        ),
    )

    Spacer(modifier = Modifier.height(32.dp))

    // Sign In button
    Button(
        onClick = onSignIn,
        modifier = Modifier
            .fillMaxWidth()
            .height(48.dp),
        enabled = username.isNotBlank() && password.isNotBlank() && !isLoading,
        colors = ButtonDefaults.buttonColors(
            containerColor = NetflixRed,
            contentColor = NetflixWhite,
        ),
    ) {
        if (isLoading) {
            CircularProgressIndicator(
                modifier = Modifier.size(24.dp),
                color = NetflixWhite,
            )
        } else {
            Text(
                text = stringResource(R.string.login_sign_in),
                fontSize = 16.sp,
                fontWeight = FontWeight.SemiBold,
            )
        }
    }

    Spacer(modifier = Modifier.height(32.dp))

    // Footer
    Text(
        text = stringResource(R.string.login_questions_contact_admin),
        fontSize = 12.sp,
        color = NetflixGray,
        textAlign = TextAlign.Center,
    )
}

@Composable
private fun ConnectionStatusBanner(
    connectionInfo: ConnectionInfo,
    onRetry: () -> Unit,
) {
    val backgroundColor = when (connectionInfo.status) {
        ConnectionStatus.CONNECTED -> Color(0xFF1a2f1a) // Dark green tint
        ConnectionStatus.FAILED -> Color(0xFF2f1a1a) // Dark red tint
        ConnectionStatus.CHECKING -> NetflixDark
        ConnectionStatus.UNKNOWN -> NetflixDark
    }

    val borderColor = when (connectionInfo.status) {
        ConnectionStatus.CONNECTED -> Color(0xFF4ade80).copy(alpha = 0.3f)
        ConnectionStatus.FAILED -> NetflixRed.copy(alpha = 0.3f)
        ConnectionStatus.CHECKING -> Color.Gray
        ConnectionStatus.UNKNOWN -> Color(0xFF333333)
    }

    val textColor = when (connectionInfo.status) {
        ConnectionStatus.CONNECTED -> Color(0xFF4ade80)
        ConnectionStatus.FAILED -> NetflixRed
        ConnectionStatus.CHECKING -> NetflixLightGray
        ConnectionStatus.UNKNOWN -> NetflixLightGray
    }

    val icon = when (connectionInfo.status) {
        ConnectionStatus.CONNECTED -> LucideIcons.Wifi
        ConnectionStatus.FAILED -> LucideIcons.WifiOff
        ConnectionStatus.CHECKING -> LucideIcons.Wifi
        ConnectionStatus.UNKNOWN -> LucideIcons.Wifi
    }

    Surface(
        shape = androidx.compose.foundation.shape.RoundedCornerShape(8.dp),
        color = backgroundColor,
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(10.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            if (connectionInfo.status == ConnectionStatus.CHECKING) {
                CircularProgressIndicator(
                    modifier = Modifier.size(16.dp),
                    strokeWidth = 2.dp,
                    color = Color.Gray,
                )
            } else {
                Icon(
                    imageVector = icon,
                    contentDescription = null,
                    tint = textColor,
                    modifier = Modifier.size(16.dp),
                )
            }

            Spacer(modifier = Modifier.width(8.dp))

            Text(
                text = when (connectionInfo.status) {
                    ConnectionStatus.CHECKING -> stringResource(R.string.login_checking_connection)
                    ConnectionStatus.CONNECTED -> stringResource(R.string.login_connected) + "${connectionInfo.latencyMs?.let { " · ${it}ms" } ?: ""}"
                    ConnectionStatus.FAILED -> connectionInfo.errorMessage ?: stringResource(R.string.login_connection_failed)
                    ConnectionStatus.UNKNOWN -> stringResource(R.string.login_not_checked)
                },
                color = textColor,
                fontSize = 13.sp,
                modifier = Modifier.weight(1f),
            )

            // Retry button
            IconButton(
                onClick = onRetry,
                enabled = connectionInfo.status != ConnectionStatus.CHECKING,
                modifier = Modifier.size(24.dp),
            ) {
                Icon(
                    imageVector = LucideIcons.Refresh,
                    contentDescription = "Retry",
                    tint = if (connectionInfo.status == ConnectionStatus.CHECKING) Color(0xFF444444) else Color(0xFF888888),
                    modifier = Modifier.size(14.dp),
                )
            }
        }
    }
}

@Preview(showBackground = true)
@Composable
fun LoginScreenServerUrlPreview() {
    VeloxTheme {
        LoginScreenContent(
            uiState = LoginUiState.Idle,
            connectionInfo = ConnectionInfo(),
            initialServerUrl = "http://192.168.1.100:8098",
            onVerifyServerUrl = { true },
            onLogin = { _, _ -> },
            onLoginSuccess = {},
            onSaveServerUrl = {},
            onCheckConnection = {},
            onClearConnectionStatus = {}
        )
    }
}

@Preview(showBackground = true)
@Composable
fun LoginScreenCredentialsPreview() {
    VeloxTheme {
        LoginScreenContent(
            uiState = LoginUiState.Idle,
            connectionInfo = ConnectionInfo(status = ConnectionStatus.CONNECTED, latencyMs = 45),
            initialServerUrl = "http://192.168.1.100:8098",
            onVerifyServerUrl = { true },
            onLogin = { _, _ -> },
            onLoginSuccess = {},
            onSaveServerUrl = {},
            onCheckConnection = {},
            onClearConnectionStatus = {},
            initialStep = LoginStep.CREDENTIALS
        )
    }
}

@Preview(showBackground = true)
@Composable
fun LoginScreenLoadingPreview() {
    VeloxTheme {
        LoginScreenContent(
            uiState = LoginUiState.Loading,
            connectionInfo = ConnectionInfo(status = ConnectionStatus.CONNECTED, latencyMs = 45),
            initialServerUrl = "http://192.168.1.100:8098",
            onVerifyServerUrl = { true },
            onLogin = { _, _ -> },
            onLoginSuccess = {},
            onSaveServerUrl = {},
            onCheckConnection = {},
            onClearConnectionStatus = {},
            initialStep = LoginStep.CREDENTIALS
        )
    }
}

@Preview(showBackground = true)
@Composable
fun LoginScreenErrorPreview() {
    VeloxTheme {
        LoginScreenContent(
            uiState = LoginUiState.Error("Invalid credentials"),
            connectionInfo = ConnectionInfo(status = ConnectionStatus.FAILED, errorMessage = "Cannot reach server"),
            initialServerUrl = "http://192.168.1.100:8098",
            onVerifyServerUrl = { true },
            onLogin = { _, _ -> },
            onLoginSuccess = {},
            onSaveServerUrl = {},
            onCheckConnection = {},
            onClearConnectionStatus = {},
            initialStep = LoginStep.CREDENTIALS
        )
    }
}
