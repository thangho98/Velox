package com.velox.app.presentation.ui.components

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.size
import com.velox.app.presentation.ui.components.LucideIcons
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.tooling.preview.Preview
import com.velox.app.presentation.navigation.Screen
import com.velox.app.presentation.ui.components.livetv.LivePulseDot
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

@Composable
fun VeloxBottomTabBar(
    currentRoute: String?,
    onTabClick: (Screen) -> Unit,
    modifier: Modifier = Modifier,
) {
    val tabs = listOf(
        TabItem(Screen.Home, "Home", LucideIcons.House),
        TabItem(Screen.Movies, "Movies", LucideIcons.Film),
        TabItem(Screen.Series, "Series", LucideIcons.Tv),
        TabItem(Screen.LiveTv, "Live TV", LucideIcons.PlayCircle),
        TabItem(Screen.Browse, "Browse", LucideIcons.FolderOpen),
        TabItem(Screen.Favorites, "Favorites", LucideIcons.Heart),
    )

    NavigationBar(
        modifier = modifier,
        containerColor = NetflixDark,
    ) {
        tabs.forEach { tab ->
            val selected = currentRoute == tab.screen.route

            NavigationBarItem(
                selected = selected,
                onClick = { onTabClick(tab.screen) },
                label = {
                    Text(
                        text = tab.label,
                        color = if (selected) NetflixWhite else NetflixLightGray,
                        style = TextStyle(fontSize = 12.sp),
                    )
                },
                icon = {
                    Box {
                        Icon(
                            imageVector = tab.icon,
                            contentDescription = tab.label,
                            modifier = Modifier.size(21.dp),
                            tint = if (selected) NetflixWhite else NetflixLightGray,
                        )
                        if (tab.screen == Screen.LiveTv) {
                            LivePulseDot(
                                color = NetflixRed,
                                size = 7.dp,
                                modifier = Modifier
                                    .align(Alignment.TopEnd)
                                    .offset(x = 4.dp, y = (-2).dp),
                            )
                        }
                    }
                },
                colors = NavigationBarItemDefaults.colors(
                    selectedIconColor = NetflixWhite,
                    unselectedIconColor = NetflixLightGray,
                    indicatorColor = NetflixDark,
                ),
            )
        }
    }
}

private data class TabItem(
    val screen: Screen,
    val label: String,
    val icon: ImageVector,
)

@Preview(showBackground = true)
@Composable
private fun VeloxBottomTabBarHomePreview() {
    VeloxTheme {
        VeloxBottomTabBar(
            currentRoute = Screen.Home.route,
            onTabClick = {},
        )
    }
}

@Preview(showBackground = true)
@Composable
private fun VeloxBottomTabBarMoviesPreview() {
    VeloxTheme {
        VeloxBottomTabBar(
            currentRoute = Screen.Movies.route,
            onTabClick = {},
        )
    }
}
