package com.velox.app.presentation.ui.components

import android.annotation.SuppressLint
import com.velox.app.presentation.ui.components.LucideIcons
import android.view.ViewGroup
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.compose.ui.res.stringResource
import com.velox.app.R
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.SkipNext
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.viewinterop.AndroidView
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

@Composable
fun YouTubePlayer(
    videoKey: String,
    title: String,
    onClose: () -> Unit,
    onSkip: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Box(
        modifier = modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        // YouTube WebView
        YouTubeWebView(
            videoKey = videoKey,
            modifier = Modifier.fillMaxSize(),
        )

        // Top bar
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.Black.copy(alpha = 0.7f))
                .padding(8.dp)
                .align(Alignment.TopCenter),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = "Trailer",
                color = NetflixWhite,
                fontSize = 14.sp,
                fontWeight = FontWeight.Medium,
            )
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                text = title,
                color = NetflixWhite.copy(alpha = 0.7f),
                fontSize = 12.sp,
                modifier = Modifier.weight(1f),
            )
            IconButton(onClick = onClose) {
                Icon(
                    LucideIcons.Close,
                    contentDescription = "Close",
                    tint = NetflixWhite,
                )
            }
        }

        // Bottom controls
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.Black.copy(alpha = 0.7f))
                .padding(16.dp)
                .align(Alignment.BottomCenter),
            horizontalArrangement = Arrangement.End,
        ) {
            Button(
                onClick = onSkip,
                colors = ButtonDefaults.buttonColors(containerColor = Color.White.copy(alpha = 0.2f)),
            ) {
                Icon(
                    LucideIcons.SkipNext,
                    contentDescription = null,
                    tint = NetflixWhite,
                )
                Spacer(modifier = Modifier.width(4.dp))
                Text(stringResource(R.string.player_skip_to_movie), color = NetflixWhite)
            }
        }
    }
}

@SuppressLint("SetJavaScriptEnabled")
@Composable
private fun YouTubeWebView(
    videoKey: String,
    modifier: Modifier = Modifier,
) {
    AndroidView(
        factory = { context ->
            WebView(context).apply {
                layoutParams = ViewGroup.LayoutParams(
                    ViewGroup.LayoutParams.MATCH_PARENT,
                    ViewGroup.LayoutParams.MATCH_PARENT,
                )
                settings.apply {
                    javaScriptEnabled = true
                    mediaPlaybackRequiresUserGesture = false
                    domStorageEnabled = true
                }
                webViewClient = WebViewClient()
                loadDataWithBaseURL(
                    "https://www.youtube.com",
                    """
                    <!DOCTYPE html>
                    <html>
                    <head>
                        <style>
                            * { margin: 0; padding: 0; }
                            html, body { height: 100%; background: #000; }
                            iframe { position: absolute; top: 50%; left: 50%; width: 200%; height: 200%; transform: translate(-50%, -50%); border: none; }
                        </style>
                    </head>
                    <body>
                        <iframe
                            src="https://www.youtube.com/embed/$videoKey?autoplay=1&controls=1&showinfo=0&rel=0&modestbranding=1&iv_load_policy=3&fs=0&playsinline=1&enablejsapi=1"
                            allow="autoplay; encrypted-media; fullscreen"
                        ></iframe>
                    </body>
                    </html>
                    """.trimIndent(),
                    "text/html",
                    "UTF-8",
                    null,
                )
            }
        },
        modifier = modifier,
    )
}

// Note: YouTubePlayer uses Android WebView which cannot render in IDE preview.
// Run on device/emulator for full preview. The annotation is for documentation.
@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun YouTubePlayerPreview() {
    VeloxTheme {
        // Preview without WebView — shows the UI chrome (top bar + bottom controls)
        YouTubePlayer(
            videoKey = "dQw4w9WgXcQ",
            title = "Never Gonna Give You Up - Official Music Video",
            onClose = {},
            onSkip = {},
        )
    }
}
