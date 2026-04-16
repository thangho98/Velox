package com.velox.app.presentation.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.RectangleShape
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.Density
import androidx.compose.foundation.Image
import coil.compose.AsyncImage
import com.vanniktech.blurhash.BlurHash
import com.velox.app.domain.model.ImageResource

@Composable
fun ResponsiveImage(
    data: ImageResource?,
    contentDescription: String?,
    modifier: Modifier = Modifier,
    contentScale: ContentScale = ContentScale.Crop,
) {
    if (data == null) {
        Box(modifier.background(Color.Black.copy(alpha = 0.4f)))
        return
    }

    val density = LocalDensity.current
    var containerWidthPx by remember { mutableIntStateOf(0) }

    Box(
        modifier = modifier
            .aspectRatio(aspectToRatio(data.aspect))
            .onSizeChanged { containerWidthPx = it.width }
            .clip(RectangleShape),
    ) {
        // Blurhash placeholder (always drawn underneath; image fades in on top)
        if (data.blurhash != null) {
            val bitmap = remember(data.blurhash) {
                try {
                    BlurHash.decode(data.blurhash, 32, 32)
                } catch (e: Exception) {
                    null
                }
            }
            if (bitmap != null) {
                Image(
                    bitmap = bitmap.asImageBitmap(),
                    contentDescription = null,
                    modifier = Modifier.matchParentSize(),
                    contentScale = ContentScale.Crop,
                )
            }
        }

        val url = remember(containerWidthPx, data.srcset) {
            pickSrcsetVariant(data, containerWidthPx, density)
        }
        AsyncImage(
            model = url,
            contentDescription = contentDescription,
            modifier = Modifier.matchParentSize(),
            contentScale = contentScale,
        )
    }
}

private fun aspectToRatio(aspect: String): Float = when (aspect) {
    "2:3" -> 2f / 3f
    "16:9" -> 16f / 9f
    else -> 1f
}

// Pick the smallest srcset variant >= container width (in px).
private fun pickSrcsetVariant(data: ImageResource, containerPx: Int, density: Density): String {
    if (containerPx == 0) return data.url
    val candidates = data.srcset.entries
        .mapNotNull { (k, url) -> k.toIntOrNull()?.let { it to url } }
        .sortedBy { it.first }
    return candidates.firstOrNull { it.first >= containerPx }?.second
        ?: candidates.lastOrNull()?.second
        ?: data.url
}
