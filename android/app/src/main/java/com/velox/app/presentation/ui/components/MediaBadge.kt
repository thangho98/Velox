package com.velox.app.presentation.ui.components

import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.ui.theme.MovieBlue
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.SeriesPurple

@Composable
fun MediaBadge(
    mediaType: String,
    modifier: Modifier = Modifier
) {
    val isMovie = mediaType.equals("movie", ignoreCase = true)
    Surface(
        modifier = modifier.padding(8.dp),
        color = if (isMovie) MovieBlue else SeriesPurple,
        shape = RoundedCornerShape(4.dp)
    ) {
        Text(
            text = if (isMovie) "Movie" else "Series",
            color = NetflixWhite,
            fontSize = 12.sp,
            fontWeight = FontWeight.Medium,
            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
        )
    }
}
