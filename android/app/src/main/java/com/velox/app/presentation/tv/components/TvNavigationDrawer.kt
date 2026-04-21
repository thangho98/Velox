package com.velox.app.presentation.tv.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Search
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.Icon
import androidx.tv.material3.NavigationDrawer
import androidx.tv.material3.NavigationDrawerItem
import androidx.tv.material3.Text

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun TvNavigationDrawer(
    initialSelectedIndex: Int = 0,
    onNavigateHome: () -> Unit,
    onNavigateSearch: () -> Unit,
    content: @Composable () -> Unit
) {
    var selectedIndex by remember { mutableIntStateOf(initialSelectedIndex) }

    NavigationDrawer(
        drawerContent = { drawerValue ->
            Column(
                modifier = Modifier
                    .background(Color(0xD9000000))
                    .fillMaxHeight()
                    .padding(12.dp)
            ) {
                NavigationDrawerItem(
                    selected = selectedIndex == 0,
                    onClick = {
                        selectedIndex = 0
                        onNavigateHome()
                    },
                    leadingContent = {
                        Icon(imageVector = Icons.Default.Home, contentDescription = "Home")
                    }
                ) {
                    Text(text = "Home")
                }

                Spacer(modifier = Modifier.padding(16.dp))

                NavigationDrawerItem(
                    selected = selectedIndex == 1,
                    onClick = {
                        selectedIndex = 1
                        onNavigateSearch()
                    },
                    leadingContent = {
                        Icon(imageVector = Icons.Default.Search, contentDescription = "Search")
                    }
                ) {
                    Text(text = "Search")
                }
            }
        }
    ) {
        content()
    }
}
