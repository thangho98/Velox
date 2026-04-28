package com.velox.app.presentation.ui.components.livetv

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.defaultMinSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.domain.model.livetv.LiveProgram
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter

private val epgTimeFormatter = DateTimeFormatter.ofPattern("HH:mm").withZone(ZoneId.systemDefault())

private fun LiveProgram.startMs(): Long = runCatching { Instant.parse(startTimeIso).toEpochMilli() }.getOrDefault(0L)
private fun LiveProgram.endMs(): Long = runCatching { Instant.parse(endTimeIso).toEpochMilli() }.getOrDefault(0L)

private fun LiveProgram.formatStart(): String = runCatching {
    epgTimeFormatter.format(Instant.parse(startTimeIso))
}.getOrDefault("")

private fun LiveProgram.formatEnd(): String = runCatching {
    epgTimeFormatter.format(Instant.parse(endTimeIso))
}.getOrDefault("")

private fun LiveProgram.progressPercent(nowMs: Long): Float {
    val s = startMs()
    val e = endMs()
    if (nowMs < s) return 0f
    if (nowMs >= e || e <= s) return 100f
    return ((nowMs - s).toFloat() / (e - s).toFloat() * 100f).coerceIn(0f, 100f)
}

private fun LiveProgram.isOnNow(nowMs: Long): Boolean {
    val s = startMs()
    val e = endMs()
    return s <= nowMs && nowMs < e
}

@Composable
fun LiveTvEpgTimeline(
    programs: List<LiveProgram>,
    channelName: String?,
    modifier: Modifier = Modifier,
) {
    if (programs.isEmpty()) return

    val nowMs = remember(programs) { System.currentTimeMillis() }

    Column(modifier = modifier.fillMaxWidth()) {
        Text(
            text = if (channelName.isNullOrBlank()) "UP NEXT" else "UP NEXT ON ${channelName.uppercase()}",
            color = NetflixLightGray,
            fontSize = 10.sp,
            fontWeight = FontWeight.Bold,
            letterSpacing = 3.sp,
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 4.dp),
        )
        Spacer(modifier = Modifier.height(6.dp))
        LazyRow(
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            contentPadding = androidx.compose.foundation.layout.PaddingValues(horizontal = 16.dp),
            modifier = Modifier.fillMaxWidth(),
        ) {
            items(programs, key = { it.id }) { program ->
                val onNow = program.isOnNow(nowMs)
                val progress = if (onNow) program.progressPercent(nowMs) else 0f
                EpgCard(
                    program = program,
                    isOnNow = onNow,
                    progressPercent = progress,
                )
            }
        }
    }
}

@Composable
private fun EpgCard(
    program: LiveProgram,
    isOnNow: Boolean,
    progressPercent: Float,
) {
    val borderMod = if (isOnNow) {
        Modifier.border(1.dp, NetflixRed, RoundedCornerShape(4.dp))
    } else {
        Modifier
    }

    Column(
        modifier = Modifier
            .defaultMinSize(minWidth = 260.dp)
            .width(280.dp)
            .clip(RoundedCornerShape(4.dp))
            .then(borderMod)
            .background(Color(0xFF181818))
            .padding(horizontal = 14.dp, vertical = 12.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            if (isOnNow) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    LivePulseDot(color = NetflixRed)
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = "ON NOW",
                        color = NetflixRed,
                        fontSize = 10.sp,
                        fontWeight = FontWeight.Bold,
                        letterSpacing = 2.sp,
                    )
                }
            } else {
                Text(
                    text = program.formatStart(),
                    color = NetflixLightGray,
                    fontSize = 10.sp,
                    fontWeight = FontWeight.Bold,
                    letterSpacing = 2.sp,
                )
            }
            Text(
                text = "${program.formatStart()} – ${program.formatEnd()}",
                color = NetflixLightGray,
                fontSize = 10.sp,
                fontWeight = FontWeight.Medium,
                letterSpacing = 1.sp,
            )
        }
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = program.title,
            color = NetflixWhite,
            fontSize = 14.sp,
            fontWeight = FontWeight.SemiBold,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
        )
        if (program.description.isNotBlank()) {
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = program.description,
                color = NetflixLightGray,
                fontSize = 11.sp,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
        }
        if (isOnNow) {
            Spacer(modifier = Modifier.height(10.dp))
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(3.dp)
                    .background(Color(0xFF2F2F2F)),
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth(progressPercent / 100f)
                        .height(3.dp)
                        .background(NetflixRed),
                )
            }
        }
    }
}
