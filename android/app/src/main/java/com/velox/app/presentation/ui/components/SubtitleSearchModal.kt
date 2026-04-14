package com.velox.app.presentation.ui.components

import androidx.compose.ui.res.stringResource
import com.velox.app.R
import androidx.compose.foundation.background
import com.velox.app.presentation.ui.components.LucideIcons
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Download
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.presentation.viewmodel.SubtitleSearchResultUi
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

/**
 * SubtitleSearchModal - Search and download subtitles from external providers.
 * Shows language selector, search results, and download button.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SubtitleSearchModal(
    mediaId: Int,
    onDismiss: () -> Unit,
    onDownloaded: () -> Unit,
    searchSubtitles: (Int, String) -> Unit,
    downloadSubtitle: (Int, String, String) -> Unit,
    isSearching: Boolean,
    isDownloading: Boolean,
    searchResults: List<SubtitleSearchResultUi>,
    selectedLang: String,
    onLangSelected: (String) -> Unit,
    modifier: Modifier = Modifier,
) {
    val languages = listOf(
        "en" to "English",
        "vi" to "Vietnamese",
        "fr" to "French",
        "de" to "German",
        "es" to "Spanish",
        "ja" to "Japanese",
        "ko" to "Korean",
        "zh" to "Chinese",
    )

    var downloadingId by remember { mutableStateOf<String?>(null) }

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        containerColor = Color(0xFF1A1A1A),
        modifier = modifier,
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 20.dp)
                .padding(bottom = 32.dp),
        ) {
            // Header
            Text(
                text = "Search Subtitles",
                color = NetflixWhite,
                fontSize = 18.sp,
                fontWeight = FontWeight.Bold,
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Language selector
            Text(
                text = "Language",
                color = NetflixWhite.copy(alpha = 0.7f),
                fontSize = 14.sp,
            )
            Spacer(modifier = Modifier.height(8.dp))
            LazyRow(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                items(languages) { (code, name) ->
                    FilterChip(
                        selected = selectedLang == code,
                        onClick = { onLangSelected(code) },
                        label = { Text(name) },
                        colors = FilterChipDefaults.filterChipColors(
                            containerColor = Color(0xFF2A2A2A),
                            labelColor = NetflixWhite,
                            selectedContainerColor = NetflixRed,
                            selectedLabelColor = NetflixWhite,
                        ),
                    )
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Search button
            Button(
                onClick = { searchSubtitles(mediaId, selectedLang) },
                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(8.dp),
            ) {
                Icon(LucideIcons.Search, contentDescription = null)
                Spacer(modifier = Modifier.width(8.dp))
                Text(stringResource(R.string.action_search))
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Loading state
            if (isSearching) {
                Box(
                    modifier = Modifier.fillMaxWidth(),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator(color = NetflixRed)
                }
            }

            // Search results
            if (searchResults.isNotEmpty()) {
                Text(
                    text = "${searchResults.size} results",
                    color = NetflixWhite.copy(alpha = 0.6f),
                    fontSize = 12.sp,
                )
                Spacer(modifier = Modifier.height(8.dp))
                LazyColumn(
                    modifier = Modifier.heightIn(max = 300.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    items(searchResults) { result ->
                        SubtitleSearchResultItem(
                            result = result,
                            isDownloading = isDownloading && downloadingId == result.id,
                            onDownload = {
                                downloadingId = result.id
                                downloadSubtitle(mediaId, result.id, result.lang)
                            },
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun SubtitleSearchResultItem(
    result: SubtitleSearchResultUi,
    isDownloading: Boolean,
    onDownload: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(
                color = Color(0xFF2A2A2A),
                shape = RoundedCornerShape(8.dp),
            )
            .padding(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = result.title,
                color = NetflixWhite,
                fontSize = 14.sp,
                fontWeight = FontWeight.Medium,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
            Spacer(modifier = Modifier.height(4.dp))
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(
                    text = result.lang.uppercase(),
                    color = NetflixRed,
                    fontSize = 12.sp,
                )
                if (result.hi) {
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "HI",
                        color = NetflixWhite.copy(alpha = 0.5f),
                        fontSize = 10.sp,
                    )
                }
                if (result.aiTranslated) {
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "AI",
                        color = Color(0xFF4CAF50),
                        fontSize = 10.sp,
                    )
                }
            }
        }

        Spacer(modifier = Modifier.width(12.dp))

        IconButton(
            onClick = onDownload,
            enabled = !isDownloading,
        ) {
            if (isDownloading) {
                CircularProgressIndicator(
                    color = NetflixRed,
                    modifier = Modifier.size(20.dp),
                    strokeWidth = 2.dp,
                )
            } else {
                Icon(
                    LucideIcons.Download,
                    contentDescription = "Download",
                    tint = NetflixRed,
                )
            }
        }
    }
}

private val SampleSubtitleResult = SubtitleSearchResultUi(
    id = "123",
    title = "The.Dark.Knight.2008.1080p.BluRay.x264-GROUP",
    lang = "en",
    hi = false,
    aiTranslated = false,
)

private val SampleSubtitleResultHI = SubtitleSearchResultUi(
    id = "456",
    title = "The.Dark.Knight.2008.HDR.2160p.WEB.H265-GROUP",
    lang = "vi",
    hi = true,
    aiTranslated = true,
)

@Preview(showBackground = true, backgroundColor = 0xFF1A1A1A)
@Composable
private fun SubtitleSearchResultItemPreview() {
    VeloxTheme {
        SubtitleSearchResultItem(
            result = SampleSubtitleResult,
            isDownloading = false,
            onDownload = {},
        )
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF1A1A1A)
@Composable
private fun SubtitleSearchResultItemDownloadingPreview() {
    VeloxTheme {
        SubtitleSearchResultItem(
            result = SampleSubtitleResultHI,
            isDownloading = true,
            onDownload = {},
        )
    }
}
