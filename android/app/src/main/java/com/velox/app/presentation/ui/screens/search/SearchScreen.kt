package com.velox.app.presentation.ui.screens.search

import androidx.compose.foundation.background
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.ui.components.NotificationBell
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Clear
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import coil.compose.AsyncImage
import com.velox.app.domain.model.MediaItem
import com.velox.app.domain.model.SeriesItem
import com.velox.app.presentation.viewmodel.SearchUiState
import com.velox.app.presentation.viewmodel.SearchViewModel
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixGray
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

@Composable
fun SearchScreen(
    onMediaClick: (Int) -> Unit,
    onSeriesClick: (Int) -> Unit,
    onBackClick: () -> Unit,
    viewModel: SearchViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()

    SearchContent(
        uiState = uiState,
        onMediaClick = onMediaClick,
        onSeriesClick = onSeriesClick,
        onBackClick = onBackClick,
        onQueryChange = viewModel::setQuery,
        onClearSearch = viewModel::clearSearch,
        onRemoveRecentSearch = viewModel::removeRecentSearch,
        onTypeFilterChange = viewModel::setTypeFilter,
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SearchContent(
    uiState: SearchUiState,
    onMediaClick: (Int) -> Unit,
    onSeriesClick: (Int) -> Unit,
    onBackClick: () -> Unit,
    onQueryChange: (String) -> Unit,
    onClearSearch: () -> Unit,
    onRemoveRecentSearch: (String) -> Unit,
    onTypeFilterChange: (String) -> Unit,
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    OutlinedTextField(
                        value = uiState.query,
                        onValueChange = onQueryChange,
                        modifier = Modifier.fillMaxWidth(),
                        placeholder = { Text("Search movies, series...", color = NetflixLightGray) },
                        leadingIcon = { Icon(LucideIcons.Search, contentDescription = "Search", tint = NetflixLightGray) },
                        trailingIcon = {
                            if (uiState.query.isNotEmpty()) {
                                IconButton(onClick = onClearSearch) {
                                    Icon(LucideIcons.Close, contentDescription = "Clear", tint = NetflixLightGray)
                                }
                            }
                        },
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedTextColor = NetflixWhite,
                            unfocusedTextColor = NetflixWhite,
                            focusedBorderColor = NetflixRed,
                            unfocusedBorderColor = NetflixGray,
                            cursorColor = NetflixRed,
                        ),
                        singleLine = true,
                        shape = RoundedCornerShape(8.dp),
                    )
                },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(LucideIcons.ChevronLeft, contentDescription = "Back", tint = NetflixWhite)
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
            if (uiState.isSearching) {
                // Loading state
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator(color = NetflixRed)
                }
            } else if (uiState.query.isEmpty()) {
                // Empty state - show recent searches
                if (uiState.recentSearches.isNotEmpty()) {
                    Column(
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(16.dp),
                    ) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically,
                        ) {
                            Text(
                                text = "Recent Searches",
                                color = NetflixWhite,
                                fontSize = 18.sp,
                                fontWeight = FontWeight.SemiBold,
                            )
                            TextButton(onClick = onClearSearch) {
                                Text(
                                    text = "Clear all",
                                    color = NetflixRed,
                                    fontSize = 14.sp,
                                )
                            }
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        uiState.recentSearches.forEach { search ->
                            Surface(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .clickable { onQueryChange(search) },
                                color = NetflixDark,
                                shape = RoundedCornerShape(8.dp),
                            ) {
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(12.dp),
                                    verticalAlignment = Alignment.CenterVertically,
                                ) {
                                    Icon(
                                        LucideIcons.Search,
                                        contentDescription = null,
                                        tint = NetflixLightGray,
                                    )
                                    Spacer(modifier = Modifier.width(12.dp))
                                    Text(
                                        text = search,
                                        color = NetflixWhite,
                                        fontSize = 14.sp,
                                        modifier = Modifier.weight(1f),
                                    )
                                    IconButton(
                                        onClick = { onRemoveRecentSearch(search) },
                                        modifier = Modifier.size(24.dp),
                                    ) {
                                        Icon(
                                            LucideIcons.Close,
                                            contentDescription = "Remove",
                                            tint = NetflixLightGray,
                                            modifier = Modifier.size(16.dp),
                                        )
                                    }
                                }
                            }
                            Spacer(modifier = Modifier.height(8.dp))
                        }
                    }
                } else {
                    // No recent searches
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center,
                    ) {
                        Column(horizontalAlignment = Alignment.CenterHorizontally) {
                            Icon(
                                LucideIcons.Search,
                                contentDescription = null,
                                tint = NetflixGray,
                                modifier = Modifier.size(64.dp),
                            )
                            Spacer(modifier = Modifier.height(16.dp))
                            Text(
                                text = "Search for movies and series",
                                color = NetflixLightGray,
                                fontSize = 16.sp,
                            )
                        }
                    }
                }
            } else if (uiState.movies.isEmpty() && uiState.series.isEmpty()) {
                // No results
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            text = "No results found",
                            color = NetflixWhite,
                            fontSize = 18.sp,
                            fontWeight = FontWeight.Medium,
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = "Try searching with different keywords",
                            color = NetflixLightGray,
                            fontSize = 14.sp,
                        )
                    }
                }
            } else {
                // Has results - show with filters
                val filteredMovies = if (uiState.typeFilter == "series") emptyList() else uiState.movies
                val filteredSeries = if (uiState.typeFilter == "movie") emptyList() else uiState.series
                val totalCount = filteredMovies.size + filteredSeries.size

                Column(modifier = Modifier.fillMaxSize()) {
                    // Results count
                    Text(
                        text = "Found $totalCount results for \"${uiState.query}\"",
                        color = NetflixLightGray,
                        fontSize = 14.sp,
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                    )

                    // Type filter chips
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .horizontalScroll(rememberScrollState())
                            .padding(horizontal = 16.dp, vertical = 4.dp),
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        listOf(
                            "" to "All",
                            "movie" to "Movies",
                            "series" to "Series",
                        ).forEach { (type, label) ->
                            FilterChip(
                                selected = uiState.typeFilter == type,
                                onClick = { onTypeFilterChange(type) },
                                label = { Text(label) },
                                colors = FilterChipDefaults.filterChipColors(
                                    selectedContainerColor = NetflixRed,
                                    selectedLabelColor = NetflixWhite,
                                    containerColor = NetflixDark,
                                    labelColor = NetflixLightGray,
                                ),
                            )
                        }
                    }

                    // Results grid
                    if (totalCount == 0) {
                        Box(
                            modifier = Modifier.fillMaxSize(),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text(
                                text = "No ${if (uiState.typeFilter == "movie") "movies" else if (uiState.typeFilter == "series") "series" else "results"} found",
                                color = NetflixLightGray,
                                fontSize = 16.sp,
                            )
                        }
                    } else {
                        LazyVerticalGrid(
                            columns = GridCells.Adaptive(minSize = 100.dp),
                            contentPadding = PaddingValues(16.dp),
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                            verticalArrangement = Arrangement.spacedBy(16.dp),
                        ) {
                            // Movies section
                            if (filteredMovies.isNotEmpty()) {
                                item {
                                    Text(
                                        text = "Movies (${filteredMovies.size})",
                                        color = NetflixWhite,
                                        fontSize = 18.sp,
                                        fontWeight = FontWeight.SemiBold,
                                        modifier = Modifier.padding(vertical = 8.dp),
                                    )
                                }
                                items(filteredMovies) { movie ->
                                    SearchMovieCard(
                                        item = movie,
                                        onClick = { onMediaClick(movie.id) },
                                    )
                                }
                            }

                            // Series section
                            if (filteredSeries.isNotEmpty()) {
                                item {
                                    Text(
                                        text = "Series (${filteredSeries.size})",
                                        color = NetflixWhite,
                                        fontSize = 18.sp,
                                        fontWeight = FontWeight.SemiBold,
                                        modifier = Modifier.padding(vertical = 8.dp),
                                    )
                                }
                                items(filteredSeries) { series ->
                                    SearchSeriesCard(
                                        item = series,
                                        onClick = { onSeriesClick(series.id) },
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun SearchMovieCard(
    item: MediaItem,
    onClick: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
    ) {
        Surface(
            modifier = Modifier
                .fillMaxWidth()
                .aspectRatio(2f / 3f),
            color = NetflixGray,
            shape = RoundedCornerShape(8.dp),
        ) {
            if (item.posterPath != null) {
                AsyncImage(
                    model = item.posterPath,
                    contentDescription = item.title,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Crop,
                )
            } else {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        text = item.title.take(2).uppercase(),
                        color = NetflixLightGray,
                        fontSize = 24.sp,
                    )
                }
            }
        }
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = item.title,
            color = NetflixWhite,
            fontSize = 14.sp,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        item.year?.let {
            Text(
                text = it.toString(),
                color = NetflixLightGray,
                fontSize = 12.sp,
            )
        }
    }
}

@Composable
fun SearchSeriesCard(
    item: SeriesItem,
    onClick: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
    ) {
        Surface(
            modifier = Modifier
                .fillMaxWidth()
                .aspectRatio(2f / 3f),
            color = NetflixGray,
            shape = RoundedCornerShape(8.dp),
        ) {
            if (item.posterPath != null) {
                AsyncImage(
                    model = item.posterPath,
                    contentDescription = item.title,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Crop,
                )
            } else {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        text = item.title.take(2).uppercase(),
                        color = NetflixLightGray,
                        fontSize = 24.sp,
                    )
                }
            }
        }
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = item.title,
            color = NetflixWhite,
            fontSize = 14.sp,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        item.year?.let {
            Text(
                text = it.toString(),
                color = NetflixLightGray,
                fontSize = 12.sp,
            )
        }
    }
}

@Preview(showBackground = true)
@Composable
fun SearchScreenPreview() {
    VeloxTheme {
        SearchContent(
            uiState = SearchUiState(
                query = "Inception",
                movies = listOf(SampleSearchMovieItem),
                series = listOf(SampleSearchSeriesItem),
                recentSearches = listOf("Batman", "Interstellar"),
            ),
            onMediaClick = {},
            onSeriesClick = {},
            onBackClick = {},
            onQueryChange = {},
            onClearSearch = {},
            onRemoveRecentSearch = {},
            onTypeFilterChange = {},
        )
    }
}

private val SampleSearchMovieItem = MediaItem(
    id = 1,
    title = "Inception",
    posterPath = null,
    backdropPath = null,
    year = 2010,
    rating = 8.8f,
    mediaType = "movie",
    overview = null,
)

private val SampleSearchSeriesItem = SeriesItem(
    id = 2,
    title = "Breaking Bad",
    posterPath = null,
    backdropPath = null,
    year = 2008,
    rating = 9.5f,
    overview = null,
    seasonCount = 5,
    episodeCount = 62,
)
