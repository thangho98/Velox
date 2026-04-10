package com.velox.app.presentation.ui.screens.browse

import androidx.compose.foundation.background
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.ui.components.NotificationBell
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.GridView
import androidx.compose.material.icons.filled.List
import androidx.compose.material.icons.filled.Movie
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.Tv
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import coil.compose.AsyncImage
import androidx.compose.ui.tooling.preview.Preview
import com.velox.app.presentation.viewmodel.BrowseUiState
import com.velox.app.ui.theme.VeloxTheme
import com.velox.app.domain.model.BrowseItem
import com.velox.app.domain.model.Library
import com.velox.app.presentation.viewmodel.BrowseViewModel
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixGray
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite

@Composable
fun BrowseScreen(
    onMediaClick: (Int) -> Unit,
    onSeriesClick: (Int) -> Unit,
    onSearchClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    onSettingsClick: () -> Unit,
    viewModel: BrowseViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()

    BrowseScreenContent(
        uiState = uiState,
        onBackClick = { viewModel.navigateBack() },
        onLibraryClick = { viewModel.selectLibrary(it) },
        onPathClick = { viewModel.navigateToPath(it) },
        onRetryClick = { viewModel.refresh() },
        onItemClick = { item ->
            if (item.isFolder) {
                viewModel.navigateToPath(item.path)
            } else {
                item.mediaId?.let { id ->
                    if (item.mediaType == "series") {
                        onSeriesClick(id)
                    } else {
                        onMediaClick(id)
                    }
                }
            }
        },
        onSearchClick = onSearchClick,
        onNotificationsClick = onNotificationsClick,
        onSettingsClick = onSettingsClick,
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BrowseScreenContent(
    uiState: BrowseUiState,
    onBackClick: () -> Unit,
    onLibraryClick: (Int) -> Unit,
    onPathClick: (String) -> Unit,
    onRetryClick: () -> Unit,
    onItemClick: (BrowseItem) -> Unit,
    onSearchClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    onSettingsClick: () -> Unit,
) {
    var isGridView by remember { mutableStateOf(false) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Browse", color = NetflixWhite, fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    if (uiState.currentPath.isNotEmpty()) {
                        IconButton(onClick = onBackClick) {
                            Icon(LucideIcons.ChevronLeft, contentDescription = "Back", tint = NetflixWhite)
                        }
                    }
                },
                actions = {
                    IconButton(onClick = { isGridView = !isGridView }) {
                        Icon(
                            if (isGridView) LucideIcons.ListIcon else LucideIcons.GridView,
                            contentDescription = "Toggle view",
                            tint = NetflixWhite,
                        )
                    }
                    NotificationBell(onClick = onNotificationsClick)
                    IconButton(onClick = onSearchClick) {
                        Icon(LucideIcons.Search, contentDescription = "Search", tint = NetflixWhite)
                    }
                    IconButton(onClick = onSettingsClick) {
                        Icon(LucideIcons.Settings, contentDescription = "Settings", tint = NetflixWhite)
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = NetflixBlack),
            )
        },
        containerColor = NetflixBlack,
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
        ) {
            // Library selector
            if (uiState.libraries.isNotEmpty()) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .horizontalScroll(rememberScrollState())
                        .padding(horizontal = 16.dp, vertical = 8.dp),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    uiState.libraries.forEach { library ->
                        FilterChip(
                            selected = uiState.selectedLibraryId == library.id,
                            onClick = { onLibraryClick(library.id) },
                            label = { Text(library.name) },
                            colors = FilterChipDefaults.filterChipColors(
                                selectedContainerColor = NetflixRed,
                                selectedLabelColor = NetflixWhite,
                                containerColor = NetflixDark,
                                labelColor = NetflixLightGray,
                            ),
                        )
                    }
                }
            }

            // Breadcrumb
            if (uiState.currentPath.isNotEmpty()) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp, vertical = 4.dp),
                    horizontalArrangement = Arrangement.spacedBy(4.dp),
                ) {
                    Text(
                        text = "Home",
                        color = NetflixRed,
                        fontSize = 14.sp,
                        modifier = Modifier.clickable { onPathClick("") },
                    )
                    uiState.currentPath.forEachIndexed { index, part ->
                        Text(" / ", color = NetflixLightGray, fontSize = 14.sp)
                        Text(
                            text = part,
                            color = if (index == uiState.currentPath.lastIndex) NetflixWhite else NetflixRed,
                            fontSize = 14.sp,
                            modifier = Modifier.clickable {
                                val path = uiState.currentPath.take(index + 1).joinToString("/")
                                onPathClick(path)
                            },
                        )
                    }
                }
            }

            // Content
            if (uiState.isLoading) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator(color = NetflixRed)
                }
            } else if (uiState.error != null) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            text = uiState.error,
                            color = MaterialTheme.colorScheme.error,
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Button(
                            onClick = onRetryClick,
                            colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                        ) {
                            Text("Retry")
                        }
                    }
                }
            } else if (uiState.items.isEmpty()) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        text = "This folder is empty",
                        color = NetflixLightGray,
                        fontSize = 16.sp,
                    )
                }
            } else {
                if (isGridView) {
                    LazyVerticalGrid(
                        columns = GridCells.Adaptive(minSize = 120.dp),
                        contentPadding = PaddingValues(16.dp),
                        horizontalArrangement = Arrangement.spacedBy(12.dp),
                        verticalArrangement = Arrangement.spacedBy(16.dp),
                    ) {
                        items(uiState.items) { item ->
                            BrowseGridItem(
                                item = item,
                                onClick = { onItemClick(item) },
                            )
                        }
                    }
                } else {
                    LazyColumn(
                        contentPadding = PaddingValues(16.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        items(uiState.items) { item ->
                            BrowseListItem(
                                item = item,
                                onClick = { onItemClick(item) },
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun BrowseListItem(
    item: BrowseItem,
    onClick: () -> Unit,
) {
    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
        color = NetflixDark,
        shape = RoundedCornerShape(8.dp),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            if (item.isFolder) {
                Box(
                    modifier = Modifier
                        .size(48.dp)
                        .clip(RoundedCornerShape(4.dp))
                        .background(NetflixGray),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(
                        LucideIcons.Folder,
                        contentDescription = null,
                        tint = NetflixRed,
                        modifier = Modifier.size(24.dp),
                    )
                }
            } else {
                AsyncImage(
                    model = item.posterPath,
                    contentDescription = item.name,
                    modifier = Modifier
                        .size(48.dp)
                        .clip(RoundedCornerShape(4.dp))
                        .background(NetflixGray),
                    contentScale = ContentScale.Crop,
                )
                if (item.posterPath == null) {
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(
                            if (item.mediaType == "series") LucideIcons.Tv else LucideIcons.Movie,
                            contentDescription = null,
                            tint = NetflixLightGray,
                            modifier = Modifier.size(24.dp),
                        )
                    }
                }
            }
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = item.name,
                    color = NetflixWhite,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Medium,
                )
                if (!item.isFolder && item.type != null) {
                    Text(
                        text = item.type.uppercase(),
                        color = NetflixLightGray,
                        fontSize = 12.sp,
                    )
                }
            }
            if (item.isFolder) {
                Icon(
                    LucideIcons.ChevronLeft,
                    contentDescription = null,
                    tint = NetflixLightGray,
                    modifier = Modifier
                        .size(20.dp)
                        .padding(start = 0.dp),
                )
            }
        }
    }
}

@Composable
fun BrowseGridItem(
    item: BrowseItem,
    onClick: () -> Unit,
) {
    Column(
        modifier = Modifier
            .width(120.dp)
            .clickable(onClick = onClick),
    ) {
        Surface(
            modifier = Modifier
                .width(120.dp)
                .height(160.dp),
            color = NetflixGray,
            shape = RoundedCornerShape(8.dp),
        ) {
            if (item.isFolder) {
                Box(
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(
                        LucideIcons.Folder,
                        contentDescription = null,
                        tint = NetflixRed,
                        modifier = Modifier.size(48.dp),
                    )
                }
            } else {
                AsyncImage(
                    model = item.posterPath,
                    contentDescription = item.name,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Crop,
                )
                if (item.posterPath == null) {
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(
                            if (item.mediaType == "series") LucideIcons.Tv else LucideIcons.Movie,
                            contentDescription = null,
                            tint = NetflixLightGray,
                            modifier = Modifier.size(48.dp),
                        )
                    }
                }
            }
        }
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = item.name,
            color = NetflixWhite,
            fontSize = 14.sp,
            maxLines = 2,
        )
    }
}

@Preview(showBackground = true)
@Composable
fun BrowseScreenPreview() {
    val sampleLibraries = listOf(
        Library(id = 1, name = "Movies", type = "movie", paths = listOf("/movies"), itemCount = 120),
        Library(id = 2, name = "TV Shows", type = "tv", paths = listOf("/tv"), itemCount = 45)
    )
    val sampleItems = listOf(
        BrowseItem(name = "Action", path = "Movies/Action", type = null, isFolder = true),
        BrowseItem(name = "Inception", path = "Movies/Inception.mp4", type = "movie", isFolder = false, mediaId = 1, mediaType = "movie"),
        BrowseItem(name = "Breaking Bad", path = "TV Shows/Breaking Bad", type = "series", isFolder = false, mediaId = 2, mediaType = "series")
    )
    val uiState = BrowseUiState(
        isLoading = false,
        libraries = sampleLibraries,
        selectedLibraryId = 1,
        items = sampleItems,
        currentPath = listOf("Movies")
    )

    VeloxTheme {
        BrowseScreenContent(
            uiState = uiState,
            onBackClick = {},
            onLibraryClick = {},
            onPathClick = {},
            onRetryClick = {},
            onItemClick = {},
            onSearchClick = {},
            onNotificationsClick = {},
            onSettingsClick = {}
        )
    }
}

@Preview(showBackground = true)
@Composable
fun BrowseListItemPreview() {
    val sampleItem = BrowseItem(
        name = "Inception",
        path = "Movies/Inception.mp4",
        type = "movie",
        isFolder = false,
        mediaId = 1,
        mediaType = "movie"
    )
    VeloxTheme {
        Box(modifier = Modifier.padding(16.dp)) {
            BrowseListItem(item = sampleItem, onClick = {})
        }
    }
}

@Preview(showBackground = true)
@Composable
fun BrowseGridItemPreview() {
    val sampleItem = BrowseItem(
        name = "Inception",
        path = "Movies/Inception.mp4",
        type = "movie",
        isFolder = false,
        mediaId = 1,
        mediaType = "movie"
    )
    VeloxTheme {
        Box(modifier = Modifier.padding(16.dp)) {
            BrowseGridItem(item = sampleItem, onClick = {})
        }
    }
}
