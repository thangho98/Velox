package com.velox.app

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.ui.Modifier
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.width
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import android.content.Intent
import android.net.Uri
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL
import com.velox.app.data.api.AuthManager
import com.velox.app.presentation.navigation.VeloxNavHost
import com.velox.app.ui.theme.VeloxTheme
import com.velox.app.ui.theme.NetflixRed
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : ComponentActivity() {

    @Inject
    lateinit var authManager: AuthManager

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            VeloxTheme {
                Surface(modifier = Modifier.fillMaxSize()) {
                    Box(modifier = Modifier.fillMaxSize()) {
                        VeloxNavHost(authManager = authManager)
                        MandatoryUpdateOverlay()
                        NotificationPermissionRequester()
                    }
                }
            }
        }
    }
}

@Composable
private fun NotificationPermissionRequester() {
    if (Build.VERSION.SDK_INT < Build.VERSION_CODES.TIRAMISU) return
    val context = LocalContext.current
    val launcher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { /* granted or not — playback still works, just no notification */ }

    LaunchedEffect(Unit) {
        val granted = ContextCompat.checkSelfPermission(
            context,
            Manifest.permission.POST_NOTIFICATIONS,
        ) == PackageManager.PERMISSION_GRANTED
        if (granted) return@LaunchedEffect

        // Ask at most once per install. If the user denies, they can re-enable
        // from system Settings — don't nag on every cold start.
        val prefs = context.getSharedPreferences(PERMISSION_PREFS, android.content.Context.MODE_PRIVATE)
        if (prefs.getBoolean(KEY_NOTIF_ASKED, false)) return@LaunchedEffect
        prefs.edit().putBoolean(KEY_NOTIF_ASKED, true).apply()
        launcher.launch(Manifest.permission.POST_NOTIFICATIONS)
    }
}

private const val PERMISSION_PREFS = "velox_permissions"
private const val KEY_NOTIF_ASKED = "notif_permission_asked"

@Composable
fun MandatoryUpdateOverlay() {
    val context = LocalContext.current
    var showOverlay by remember { mutableStateOf(false) }
    var downloadUrl by remember { mutableStateOf<String?>(null) }
    var releaseNotes by remember { mutableStateOf("") }
    var versionName by remember { mutableStateOf("") }

    val currentVersionCode = com.velox.app.BuildConfig.VERSION_CODE

    LaunchedEffect(Unit) {
        withContext(Dispatchers.IO) {
            try {
                val baseUrl = com.velox.app.BuildConfig.BASE_URL.trimEnd('/')
                val url = URL("$baseUrl/app-versions/latest?platform=android")
                val connection = url.openConnection() as HttpURLConnection
                connection.requestMethod = "GET"
                connection.setRequestProperty("Accept", "application/json")

                if (connection.responseCode == 200) {
                    val response = connection.inputStream.bufferedReader().readText()
                    val json = JSONObject(response)
                    val vCode = json.getInt("version_code")
                    val isMandatory = json.optBoolean("is_mandatory", false)

                    if (isMandatory && vCode > currentVersionCode) {
                        val dUrl = json.getString("download_url")
                        val vName = json.getString("version_name")
                        val rNotes = json.optString("release_notes", "")
                        withContext(Dispatchers.Main) {
                            downloadUrl = dUrl
                            versionName = vName
                            releaseNotes = rNotes
                            showOverlay = true
                        }
                    }
                }
            } catch (e: Exception) {
                // Ignore silent failure
            }
        }
    }

    if (showOverlay && downloadUrl != null) {
        AlertDialog(
            onDismissRequest = { /* Cannot dismiss */ },
            confirmButton = {
                Button(
                    onClick = {
                        val intent = Intent(Intent.ACTION_VIEW, Uri.parse(downloadUrl))
                        context.startActivity(intent)
                    },
                    colors = ButtonDefaults.buttonColors(containerColor = NetflixRed)
                ) {
                    Text("Update Now", color = androidx.compose.ui.graphics.Color.White)
                }
            },
            title = {
                Text(text = "Critical Update Required")
            },
            text = {
                Column {
                    Text("Velox version $versionName contains mandatory updates or breaking changes. Please download and install the new version to continue using the app.")
                    if (releaseNotes.isNotEmpty()) {
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(text = "Release notes:", style = androidx.compose.material3.MaterialTheme.typography.labelMedium)
                        Text(text = releaseNotes, style = androidx.compose.material3.MaterialTheme.typography.bodySmall)
                    }
                }
            },
            properties = androidx.compose.ui.window.DialogProperties(dismissOnBackPress = false, dismissOnClickOutside = false)
        )
    }
}
