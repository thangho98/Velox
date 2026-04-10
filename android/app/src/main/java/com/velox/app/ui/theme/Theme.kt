package com.velox.app.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.runtime.Composable

private val DarkColorScheme = darkColorScheme(
    primary = NetflixRed,
    onPrimary = NetflixWhite,
    primaryContainer = NetflixRedHover,
    onPrimaryContainer = NetflixWhite,
    secondary = NetflixBlue,
    onSecondary = NetflixWhite,
    secondaryContainer = NetflixDark,
    onSecondaryContainer = NetflixWhite,
    tertiary = NetflixLightGray,
    onTertiary = NetflixBlack,
    background = NetflixBlack,
    onBackground = NetflixWhite,
    surface = NetflixDark,
    onSurface = NetflixWhite,
    surfaceVariant = NetflixGray,
    onSurfaceVariant = NetflixLightGray,
    error = Error,
    onError = NetflixWhite,
    errorContainer = Error,
    onErrorContainer = NetflixWhite,
    outline = NetflixGray,
    outlineVariant = NetflixLightGray,
)

@Composable
fun VeloxTheme(
    content: @Composable () -> Unit,
) {
    MaterialTheme(
        colorScheme = DarkColorScheme,
        typography = Typography,
        content = content,
    )
}
