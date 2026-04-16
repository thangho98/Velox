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
import com.velox.app.ui.theme.*
import com.velox.app.ui.theme.TextMuted
import java.net.HttpURLConnection
import java.net.URL
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.json.JSONObject

@Composable
fun FeedbackCard() {
    val context = LocalContext.current
    val appVersion = com.velox.app.BuildConfig.VERSION_NAME
    val vecode = com.velox.app.BuildConfig.VERSION_CODE
    val device = android.os.Build.MODEL
    val osVersion = android.os.Build.VERSION.RELEASE
    val apiLevel = android.os.Build.VERSION.SDK_INT

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 24.dp, vertical = 8.dp)
    ) {
        Text(
            text = "SUPPORT",
            color = NetflixLightGray,
            fontSize = 12.sp,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(bottom = 8.dp)
        )
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .clickable {
                    val intent = Intent(Intent.ACTION_SENDTO).apply {
                        data = Uri.parse("mailto:thanglong2098@gmail.com")
                        putExtra(Intent.EXTRA_SUBJECT, "[Velox Bug Report] Feedback from Android App")
                        putExtra(
                            Intent.EXTRA_TEXT,
                            "Vui lòng mô tả lỗi hoặc góp ý của bạn ở dưới đây:\n\n\n\n\n\n" +
                            "========================\n" +
                            "App Version: \$appVersion (\$vecode)\n" +
                            "Device Model: \$device\n" +
                            "Android Version: \$osVersion (API \$apiLevel)\n" +
                            "========================\n"
                        )
                    }
                    try {
                        context.startActivity(intent)
                    } catch (e: Exception) {
                        android.widget.Toast.makeText(context, "No email app installed.", android.widget.Toast.LENGTH_SHORT).show()
                    }
                },
            colors = CardDefaults.cardColors(containerColor = Color(0xFF1E1E1E)),
            shape = RoundedCornerShape(12.dp)
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(
                    imageVector = Icons.Default.Email,
                    contentDescription = "Report Bug",
                    tint = NetflixWhite,
                    modifier = Modifier.size(24.dp)
                )
                Spacer(modifier = Modifier.width(16.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Text("Report a Bug / Feedback", color = NetflixWhite, fontWeight = FontWeight.SemiBold, fontSize = 16.sp)
                    Text("Gửi email trực tiếp cho tác giả", color = NetflixLightGray, fontSize = 13.sp)
                }
                Icon(
                    imageVector = LucideIcons.ChevronRight,
                    contentDescription = null,
                    tint = NetflixGray,
                    modifier = Modifier.size(20.dp)
                )
            }
        }
    }
}

