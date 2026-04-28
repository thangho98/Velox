package com.velox.app.presentation.navigation

sealed class Screen(val route: String) {
    data object Splash : Screen("splash")
    data object Login : Screen("login")
    data object Home : Screen("home")
    data object Movies : Screen("movies")
    data object Series : Screen("series")
    data object Browse : Screen("browse")
    data object Favorites : Screen("favorites")
    data object Search : Screen("search")
    data object MediaDetail : Screen("media/{mediaId}") {
        fun createRoute(mediaId: Int) = "media/$mediaId"
    }
    data object SeriesDetail : Screen("series/{seriesId}") {
        fun createRoute(seriesId: Int) = "series/$seriesId"
    }
    data object Player : Screen("player/{mediaId}") {
        fun createRoute(mediaId: Int) = "player/$mediaId"
    }
    data object Settings : Screen("settings")
    data object Profile : Screen("profile")
    data object Notifications : Screen("notifications")
    data object LiveTv : Screen("livetv")
    data object LiveTvPlayer : Screen("livetv_player/{channelId}") {
        fun createRoute(channelId: Int) = "livetv_player/$channelId"
    }
}
