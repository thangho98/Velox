package com.velox.app.presentation.ui.components.player

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.material3.HorizontalDivider
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.R
import com.velox.app.presentation.ui.components.LucideIcons

private val MonoFamily = FontFamily.Monospace

@Composable
internal fun PlaybackStatsOverlay(
    visible: Boolean,
    playbackInfo: com.velox.app.domain.model.PlaybackInfo?,
    onClose: () -> Unit,
    modifier: Modifier = Modifier,
) {
    if (visible && playbackInfo != null) {
        val isTranscoding =
            playbackInfo.method == "FullTranscode" ||
                playbackInfo.method == "TranscodeAudio"
        val isPreTranscode = playbackInfo.method == "PreTranscode"

        val ptCodecValid =
            playbackInfo.ptVideoCodec != null &&
                playbackInfo.ptVideoCodec != "copy"
        val ptHeightValid = playbackInfo.ptHeight != null && playbackInfo.ptHeight > 0
        val videoCodec = if (isPreTranscode && ptCodecValid) {
            playbackInfo.ptVideoCodec
        } else {
            playbackInfo.videoCodec
        }
        val videoHeight = if (isPreTranscode && ptHeightValid) {
            playbackInfo.ptHeight!!
        } else {
            playbackInfo.height ?: 0
        }
        val videoWidth = if (isPreTranscode && ptHeightValid &&
            playbackInfo.width != null && playbackInfo.height != null &&
            playbackInfo.height > 0
        ) {
            (playbackInfo.width.toFloat() / playbackInfo.height *
                playbackInfo.ptHeight!!).toInt()
        } else {
            playbackInfo.width ?: 0
        }
        val videoBitrate = if (isPreTranscode && playbackInfo.ptVideoBitrate != null &&
            playbackInfo.ptVideoBitrate > 0
        ) {
            playbackInfo.ptVideoBitrate
        } else {
            playbackInfo.bitrate ?: 0
        }

        val selectedAudio = playbackInfo.audioTracks.find { it.selected }
            ?: playbackInfo.audioTracks.find { it.isDefault }
            ?: playbackInfo.audioTracks.firstOrNull()

        Box(
            modifier = modifier.fillMaxSize().padding(start = 16.dp, top = 80.dp),
        ) {
            Surface(
                modifier = Modifier.width(320.dp).align(Alignment.TopStart),
                shape = RoundedCornerShape(12.dp),
                color = Color.Black.copy(alpha = 0.7f),
            ) {
                Box {
                    Column {
                        PlaybackMethodSection(
                            playbackInfo = playbackInfo,
                        )

                        HorizontalDivider(
                            color = Color.White.copy(alpha = 0.1f),
                            thickness = 1.dp,
                        )

                        VideoSection(
                            playbackInfo = playbackInfo,
                            isPreTranscode = isPreTranscode,
                            videoCodec = videoCodec,
                            videoHeight = videoHeight,
                            videoWidth = videoWidth,
                            videoBitrate = videoBitrate,
                        )

                        if (selectedAudio != null) {
                            HorizontalDivider(
                                color = Color.White.copy(alpha = 0.1f),
                                thickness = 1.dp,
                            )
                            AudioSection(
                                selectedAudio = selectedAudio,
                                isTranscoding = isTranscoding,
                                isPreTranscode = isPreTranscode,
                            )
                        }

                        HorizontalDivider(
                            color = Color.White.copy(alpha = 0.1f),
                            thickness = 1.dp,
                        )
                        StreamSection(
                            playbackInfo = playbackInfo,
                            isPreTranscode = isPreTranscode,
                            isTranscoding = isTranscoding,
                        )
                    }

                    IconButton(
                        onClick = onClose,
                        modifier = Modifier.align(Alignment.TopEnd).size(36.dp),
                    ) {
                        Icon(
                            LucideIcons.Close,
                            contentDescription = "Close",
                            tint = Color.White.copy(alpha = 0.4f),
                            modifier = Modifier.size(16.dp),
                        )
                    }
                }
            }
        }
    }
}

@Suppress("UNUSED_PARAMETER")
@Composable
private fun PlaybackMethodSection(
    playbackInfo: com.velox.app.domain.model.PlaybackInfo,
    isTranscoding: Boolean = false,
    isPreTranscode: Boolean = false,
) {
    Column(modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Text(
                stringResource(R.string.player_playback),
                color = Color.White,
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
            )
            MethodBadge(playbackInfo.method)
        }
        if (playbackInfo.decisionReason != null) {
            Spacer(Modifier.height(6.dp))
            Text(
                text = playbackInfo.decisionReason,
                color = Color.White.copy(alpha = 0.6f),
                fontSize = 12.sp,
                fontFamily = MonoFamily,
                lineHeight = 18.sp,
            )
        }
    }
}

@Composable
private fun VideoSection(
    playbackInfo: com.velox.app.domain.model.PlaybackInfo,
    isPreTranscode: Boolean,
    videoCodec: String?,
    videoHeight: Int,
    videoWidth: Int,
    videoBitrate: Int,
) {
    Column(modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Text(
                stringResource(R.string.player_video),
                color = Color.White,
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
            )
            when {
                playbackInfo.method == "FullTranscode" ->
                    StatusBadge("Transcoding", Color(0xFFF87171))
                isPreTranscode -> StatusBadge("PreTranscode", Color(0xFF22D3EE))
                else -> StatusBadge("Direct", Color(0xFF4ADE80))
            }
        }
        Spacer(Modifier.height(6.dp))

        val codecLine = buildString {
            append(videoCodec?.uppercase() ?: "—")
            if (videoHeight > 0) append("  $videoWidth×$videoHeight")
            if (!isPreTranscode && playbackInfo.videoProfile != null) {
                append("  ${playbackInfo.videoProfile}")
            }
            if (!isPreTranscode && playbackInfo.videoLevel != null) {
                append("  L${playbackInfo.videoLevel}")
            }
        }
        Text(
            codecLine,
            color = Color.White.copy(alpha = 0.8f),
            fontSize = 12.sp,
            fontFamily = MonoFamily,
            lineHeight = 18.sp,
        )

        val bitrateLineParts = mutableListOf<String>()
        if (videoBitrate > 0) {
            val brStr = if (videoBitrate >= 1000) {
                "%.1f Mbps".format(videoBitrate / 1000f)
            } else "$videoBitrate Kbps"
            bitrateLineParts.add(brStr)
        }
        if (playbackInfo.videoFps != null && playbackInfo.videoFps > 0) {
            val fpsVal = playbackInfo.videoFps
            val fpsStr = if (fpsVal == fpsVal.toInt().toFloat()) {
                "${fpsVal.toInt()} fps"
            } else {
                "%.2f fps".format(fpsVal)
            }
            bitrateLineParts.add(fpsStr)
        }
        if (bitrateLineParts.isNotEmpty()) {
            Text(
                bitrateLineParts.joinToString(" · "),
                color = Color.White.copy(alpha = 0.6f),
                fontSize = 12.sp,
                fontFamily = MonoFamily,
                lineHeight = 18.sp,
            )
        }
        Spacer(Modifier.height(4.dp))
        Text(
            "Dropped: 0",
            color = Color.White.copy(alpha = 0.6f),
            fontSize = 12.sp,
            fontFamily = MonoFamily,
        )
    }
}

@Composable
private fun AudioSection(
    selectedAudio: com.velox.app.domain.model.AudioTrack,
    isTranscoding: Boolean,
    isPreTranscode: Boolean,
) {
    Column(modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Text(
                stringResource(R.string.player_audio),
                color = Color.White,
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
            )
            when {
                isTranscoding -> StatusBadge("Transcoding", Color(0xFFFACC15))
                isPreTranscode -> StatusBadge("PreTranscode", Color(0xFF22D3EE))
                else -> StatusBadge("Direct", Color(0xFF4ADE80))
            }
        }
        Spacer(Modifier.height(6.dp))

        val channelLayout = formatChannelLayout(selectedAudio.channels)
        val audioLine = buildString {
            append(selectedAudio.codec.uppercase())
            append(" $channelLayout")
            if (selectedAudio.language.isNotEmpty()) {
                append(" · ${selectedAudio.language}")
            }
            if (selectedAudio.isDefault) append(" (Default)")
        }
        Text(
            audioLine,
            color = Color.White.copy(alpha = 0.8f),
            fontSize = 12.sp,
            fontFamily = MonoFamily,
            lineHeight = 18.sp,
        )

        val audioBitrateParts = mutableListOf<String>()
        if (selectedAudio.bitrate != null && selectedAudio.bitrate > 0) {
            val brStr = if (selectedAudio.bitrate >= 1000) {
                "${selectedAudio.bitrate / 1000} Kbps"
            } else "${selectedAudio.bitrate} bps"
            audioBitrateParts.add(brStr)
        }
        if (selectedAudio.sampleRate != null && selectedAudio.sampleRate > 0) {
            audioBitrateParts.add("${selectedAudio.sampleRate} Hz")
        }
        if (audioBitrateParts.isNotEmpty()) {
            Text(
                audioBitrateParts.joinToString(" · "),
                color = Color.White.copy(alpha = 0.6f),
                fontSize = 12.sp,
                fontFamily = MonoFamily,
                lineHeight = 18.sp,
            )
        }
    }
}

@Composable
private fun StreamSection(
    playbackInfo: com.velox.app.domain.model.PlaybackInfo,
    isPreTranscode: Boolean,
    isTranscoding: Boolean,
) {
    Column(modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)) {
        Text(
            stringResource(R.string.player_stream),
            color = Color.White,
            fontSize = 14.sp,
            fontWeight = FontWeight.Bold,
        )
        Spacer(Modifier.height(6.dp))

        val streamType = when {
            playbackInfo.method == "DirectPlay" -> "HTTP Range"
            isPreTranscode -> "HTTP Range (PreTranscode)"
            else -> "HLS"
        }
        val containerStr = if (isPreTranscode) {
            "MP4"
        } else {
            playbackInfo.container?.uppercase() ?: "—"
        }

        val streamLine = buildString {
            append("$streamType · $containerStr")
            if (!isPreTranscode && playbackInfo.fileSize != null &&
                playbackInfo.fileSize > 0
            ) {
                val gbStr = "%.1f GB".format(
                    playbackInfo.fileSize / (1024.0 * 1024.0 * 1024.0),
                )
                append(" · $gbStr")
            }
        }
        Text(
            streamLine,
            color = Color.White.copy(alpha = 0.8f),
            fontSize = 12.sp,
            fontFamily = MonoFamily,
            lineHeight = 18.sp,
        )

        val hasEstimate = playbackInfo.estimatedBitrate != null &&
            playbackInfo.estimatedBitrate > 0 &&
            (isTranscoding || isPreTranscode)
        if (hasEstimate) {
            val estVal = playbackInfo.estimatedBitrate!!
            val estStr = if (estVal >= 1000) {
                "%.1f Mbps".format(estVal / 1000f)
            } else "$estVal Kbps"
            Text(
                "Estimated: $estStr",
                color = Color.White.copy(alpha = 0.6f),
                fontSize = 12.sp,
                fontFamily = MonoFamily,
                lineHeight = 18.sp,
            )
        }
    }
}

@Composable
private fun MethodBadge(method: String) {
    val (bgColor, textColor, label) = when (method) {
        "DirectPlay" -> Triple(
            Color(0xFF4ADE80).copy(alpha = 0.2f),
            Color(0xFF4ADE80),
            "Direct Play",
        )
        "DirectStream" -> Triple(
            Color(0xFF60A5FA).copy(alpha = 0.2f),
            Color(0xFF60A5FA),
            "Direct Stream",
        )
        "PreTranscode" -> Triple(
            Color(0xFF22D3EE).copy(alpha = 0.2f),
            Color(0xFF22D3EE),
            "PreTranscode",
        )
        "TranscodeAudio" -> Triple(
            Color(0xFFFACC15).copy(alpha = 0.2f),
            Color(0xFFFACC15),
            "Transcode Audio",
        )
        "FullTranscode" -> Triple(
            Color(0xFFF87171).copy(alpha = 0.2f),
            Color(0xFFF87171),
            "Full Transcode",
        )
        else -> Triple(
            Color.White.copy(alpha = 0.1f),
            Color.White.copy(alpha = 0.6f),
            method,
        )
    }
    Surface(shape = RoundedCornerShape(4.dp), color = bgColor) {
        Text(
            label,
            color = textColor,
            fontSize = 10.sp,
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
        )
    }
}

@Composable
private fun StatusBadge(label: String, color: Color) {
    Surface(shape = RoundedCornerShape(4.dp), color = color.copy(alpha = 0.2f)) {
        Text(
            label,
            color = color,
            fontSize = 10.sp,
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
        )
    }
}

internal fun formatChannelLayout(channels: Int): String {
    return when (channels) {
        1 -> "1.0"
        2 -> "2.0"
        6 -> "5.1"
        8 -> "7.1"
        else -> "$channels.0"
    }
}

