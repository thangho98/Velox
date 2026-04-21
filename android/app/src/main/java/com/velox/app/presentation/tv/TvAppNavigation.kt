package com.velox.app.presentation.tv

import androidx.compose.foundation.layout.Row
import androidx.compose.runtime.Composable
import androidx.navigation.NamedNavArgument
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import androidx.tv.material3.ExperimentalTvMaterial3Api
import com.velox.app.presentation.tv.components.TvNavigationDrawer
import com.velox.app.presentation.tv.screens.TvHomeScreen
import com.velox.app.presentation.tv.screens.TvLoginScreen
import com.velox.app.presentation.tv.screens.TvMediaDetailScreen
import com.velox.app.presentation.tv.screens.TvSearchScreen

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun TvAppNavigation(startDestination: String = "login") {
    val navController = rememberNavController()

    NavHost(navController = navController, startDestination = startDestination) {
        composable("login") {
            TvLoginScreen(
                onLoginSuccess = {
                    navController.navigate("home") {
                        popUpTo("login") { inclusive = true }
                    }
                }
            )
        }



        composable("home") {
            TvNavigationDrawer(
                initialSelectedIndex = 0,
                onNavigateHome = { /* Already here */ },
                onNavigateSearch = { navController.navigate("search") }
            ) {
                TvHomeScreen(
                    onNavigateToMedia = { mediaId ->
                        navController.navigate("details/$mediaId")
                    },
                    onNavigateToSeries = { seriesId ->
                        navController.navigate("series/$seriesId")
                    }
                )
            }
        }

        composable("search") {
            TvNavigationDrawer(
                initialSelectedIndex = 1,
                onNavigateHome = {
                    navController.navigate("home") {
                        popUpTo("home") { inclusive = true }
                    }
                },
                onNavigateSearch = { /* Already here */ }
            ) {
                TvSearchScreen(
                    onNavigateToMedia = { mediaId ->
                        navController.navigate("details/$mediaId")
                    },
                    onNavigateToSeries = { seriesId ->
                        navController.navigate("series/$seriesId")
                    }
                )
            }
        }

        composable(
            route = "details/{mediaId}",
            arguments = listOf(navArgument("mediaId") { type = NavType.IntType })
        ) { backStackEntry ->
            val mediaId = backStackEntry.arguments?.getInt("mediaId") ?: return@composable
            TvMediaDetailScreen(
                mediaId = mediaId,
                onNavigateToPlayer = { id ->
                    navController.navigate("player/$id")
                }
            )
        }

        composable(
            route = "series/{seriesId}",
            arguments = listOf(navArgument("seriesId") { type = NavType.IntType })
        ) { backStackEntry ->
            val seriesId = backStackEntry.arguments?.getInt("seriesId") ?: return@composable
            com.velox.app.presentation.tv.screens.TvSeriesDetailScreen(
                seriesId = seriesId,
                onNavigateToPlayer = { id ->
                    navController.navigate("player/$id")
                }
            )
        }

        composable(
            route = "player/{mediaId}",
            arguments = listOf(navArgument("mediaId") { type = NavType.IntType })
        ) { backStackEntry ->
            val mediaId = backStackEntry.arguments?.getInt("mediaId") ?: return@composable
            com.velox.app.presentation.tv.screens.TvPlayerScreen(
                mediaId = mediaId,
                onNavigateBack = {
                    navController.popBackStack()
                }
            )
        }
    }
}
