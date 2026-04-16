package com.velox.app.presentation.navigation

import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.navigation.NavHostController
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.velox.app.data.api.AuthManager
import com.velox.app.presentation.ui.components.VeloxBottomTabBar
import com.velox.app.presentation.ui.screens.auth.LoginScreen
import com.velox.app.presentation.ui.screens.home.HomeScreen
import com.velox.app.presentation.ui.screens.movies.MoviesScreen
import com.velox.app.presentation.ui.screens.series.SeriesScreen
import com.velox.app.presentation.ui.screens.browse.BrowseScreen
import com.velox.app.presentation.ui.screens.favorites.FavoritesScreen
import com.velox.app.presentation.ui.screens.search.SearchScreen
import com.velox.app.presentation.ui.screens.detail.MediaDetailScreen
import com.velox.app.presentation.ui.screens.detail.SeriesDetailScreen
import com.velox.app.presentation.ui.screens.player.PlayerScreen
import com.velox.app.presentation.ui.screens.settings.SettingsScreen
import com.velox.app.presentation.ui.screens.settings.ProfileScreen
import com.velox.app.presentation.ui.screens.settings.NotificationsScreen
import com.velox.app.presentation.ui.screens.splash.SplashScreen

private val mainTabs = listOf(
    Screen.Home.route,
    Screen.Movies.route,
    Screen.Series.route,
    Screen.Browse.route,
    Screen.Favorites.route,
)

@Suppress("UNUSED_PARAMETER")
@Composable
fun VeloxNavHost(
    navController: NavHostController = rememberNavController(),
    authManager: AuthManager,
    startDestination: String = Screen.Splash.route,
) {
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route

    Scaffold(
        contentWindowInsets = WindowInsets(0),
        bottomBar = {
            if (currentRoute in mainTabs) {
                VeloxBottomTabBar(
                    currentRoute = currentRoute,
                    onTabClick = { screen ->
                        if (currentRoute != screen.route) {
                            navController.navigate(screen.route) {
                                popUpTo(Screen.Home.route) {
                                    saveState = true
                                }
                                launchSingleTop = true
                                restoreState = true
                            }
                        }
                    },
                )
            }
        },
    ) { padding ->
        NavHost(
            navController = navController,
            startDestination = startDestination,
            modifier = Modifier.padding(padding),
        ) {
            // Splash
            composable(Screen.Splash.route) {
                SplashScreen(
                    onNavigateToHome = {
                        navController.navigate(Screen.Home.route) {
                            popUpTo(Screen.Splash.route) { inclusive = true }
                            launchSingleTop = true
                        }
                    },
                    onNavigateToLogin = {
                        navController.navigate(Screen.Login.route) {
                            popUpTo(Screen.Splash.route) { inclusive = true }
                            launchSingleTop = true
                        }
                    }
                )
            }

            // Auth
            composable(Screen.Login.route) {
                LoginScreen(
                    onLoginSuccess = {
                        navController.navigate(Screen.Home.route) {
                            popUpTo(Screen.Login.route) { inclusive = true }
                        }
                    },
                )
            }

            // Main tabs
            composable(Screen.Home.route) {
                HomeScreen(
                    onMediaClick = { mediaId ->
                        navController.navigate(Screen.MediaDetail.createRoute(mediaId))
                    },
                    onPlayClick = { mediaId ->
                        navController.navigate(Screen.Player.createRoute(mediaId))
                    },
                    onSeriesClick = { seriesId ->
                        navController.navigate(Screen.SeriesDetail.createRoute(seriesId))
                    },
                    onNotificationsClick = {
                        navController.navigate(Screen.Notifications.route)
                    },
                    onSearchClick = {
                        navController.navigate(Screen.Search.route)
                    },
                    onSettingsClick = {
                        navController.navigate(Screen.Settings.route)
                    },
                    onNavigateToMovies = {
                        navController.navigate(Screen.Movies.route) {
                            popUpTo(Screen.Home.route) { saveState = true }
                            launchSingleTop = true
                            restoreState = true
                        }
                    },
                    onNavigateToSeries = {
                        navController.navigate(Screen.Series.route) {
                            popUpTo(Screen.Home.route) { saveState = true }
                            launchSingleTop = true
                            restoreState = true
                        }
                    },
                )
            }

            composable(Screen.Movies.route) {
                MoviesScreen(
                    onMediaClick = { mediaId ->
                        navController.navigate(Screen.MediaDetail.createRoute(mediaId))
                    },
                    onNotificationsClick = {
                        navController.navigate(Screen.Notifications.route)
                    },
                    onSearchClick = {
                        navController.navigate(Screen.Search.route)
                    },
                    onSettingsClick = {
                        navController.navigate(Screen.Settings.route)
                    },
                )
            }

            composable(Screen.Series.route) {
                SeriesScreen(
                    onSeriesClick = { seriesId ->
                        navController.navigate(Screen.SeriesDetail.createRoute(seriesId))
                    },
                    onNotificationsClick = {
                        navController.navigate(Screen.Notifications.route)
                    },
                    onSearchClick = {
                        navController.navigate(Screen.Search.route)
                    },
                    onSettingsClick = {
                        navController.navigate(Screen.Settings.route)
                    },
                )
            }

            composable(Screen.Browse.route) {
                BrowseScreen(
                    onMediaClick = { mediaId ->
                        navController.navigate(Screen.MediaDetail.createRoute(mediaId))
                    },
                    onSeriesClick = { seriesId ->
                        navController.navigate(Screen.SeriesDetail.createRoute(seriesId))
                    },
                    onNotificationsClick = {
                        navController.navigate(Screen.Notifications.route)
                    },
                    onSearchClick = {
                        navController.navigate(Screen.Search.route)
                    },
                    onSettingsClick = {
                        navController.navigate(Screen.Settings.route)
                    },
                )
            }

            composable(Screen.Favorites.route) {
                FavoritesScreen(
                    onMediaClick = { mediaId ->
                        navController.navigate(Screen.MediaDetail.createRoute(mediaId))
                    },
                    onSeriesClick = { seriesId ->
                        navController.navigate(Screen.SeriesDetail.createRoute(seriesId))
                    },
                    onNotificationsClick = {
                        navController.navigate(Screen.Notifications.route)
                    },
                    onSearchClick = {
                        navController.navigate(Screen.Search.route)
                    },
                    onSettingsClick = {
                        navController.navigate(Screen.Settings.route)
                    },
                )
            }

            composable(Screen.Search.route) {
                SearchScreen(
                    onMediaClick = { mediaId ->
                        navController.navigate(Screen.MediaDetail.createRoute(mediaId))
                    },
                    onSeriesClick = { seriesId ->
                        navController.navigate(Screen.SeriesDetail.createRoute(seriesId))
                    },
                    onBackClick = {
                        navController.popBackStack()
                    },
                )
            }

            // Detail screens
            composable(
                route = Screen.MediaDetail.route,
                arguments = listOf(navArgument("mediaId") { type = NavType.IntType }),
            ) { backStackEntry ->
                val mediaId = backStackEntry.arguments?.getInt("mediaId") ?: return@composable
                MediaDetailScreen(
                    mediaId = mediaId,
                    onBackClick = { navController.popBackStack() },
                    onPlayClick = {
                        navController.navigate(Screen.Player.createRoute(mediaId))
                    },
                )
            }

            composable(
                route = Screen.SeriesDetail.route,
                arguments = listOf(navArgument("seriesId") { type = NavType.IntType }),
            ) { backStackEntry ->
                val seriesId = backStackEntry.arguments?.getInt("seriesId") ?: return@composable
                SeriesDetailScreen(
                    seriesId = seriesId,
                    onBackClick = { navController.popBackStack() },
                    onEpisodeClick = { episodeId ->
                        navController.navigate(Screen.Player.createRoute(episodeId))
                    },
                )
            }

            // Player
            composable(
                route = Screen.Player.route,
                arguments = listOf(navArgument("mediaId") { type = NavType.IntType }),
            ) {
                PlayerScreen(
                    onBackClick = { navController.popBackStack() },
                    onNavigateToMedia = { nextMediaId ->
                        // Navigate to next episode — pop current player, push new one
                        navController.popBackStack()
                        navController.navigate(Screen.Player.createRoute(nextMediaId))
                    },
                )
            }

            // Settings
            composable(Screen.Settings.route) {
                SettingsScreen(
                    onBackClick = { navController.popBackStack() },
                    onProfileClick = {
                        navController.navigate(Screen.Profile.route)
                    },
                    onNotificationsClick = {
                        navController.navigate(Screen.Notifications.route)
                    },
                )
            }

            composable(Screen.Profile.route) {
                ProfileScreen(
                    onBackClick = { navController.popBackStack() },
                    onSettingsClick = {
                        navController.navigate(Screen.Settings.route)
                    },
                )
            }

            composable(Screen.Notifications.route) {
                NotificationsScreen(
                    onBackClick = { navController.popBackStack() },
                )
            }
        }
    }
}
