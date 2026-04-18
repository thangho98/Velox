package com.velox.app.presentation.ui.screens.movies

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.itemsIndexed
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.velox.app.presentation.ui.components.ResponsiveImage
import com.velox.app.R
import com.velox.app.domain.model.MediaItem
import com.velox.app.presentation.ui.components.FilterBottomSheet
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.ui.components.NotificationBell
import com.velox.app.presentation.ui.components.QuickActionsMenu
import com.velox.app.presentation.viewmodel.MoviesUiState
import com.velox.app.presentation.viewmodel.MoviesViewModel
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixGray
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme
import kotlinx.coroutines.launch

private val ALPHABET = ('A'..'Z').toList() + '#'

@Composable
fun MoviesScreen(
    onMediaClick: (Int) -> Unit,
    onSearchClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    onSettingsClick: () -> Unit,
    viewModel: MoviesViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    MoviesContent(
        uiState = uiState,
        onMediaClick = onMediaClick,
        onSearchClick = onSearchClick,
        onNotificationsClick = onNotificationsClick,
        onSettingsClick = onSettingsClick,
        onSearchQueryChange = viewModel::setSearchQuery,
        onSortByChange = viewModel::setSortBy,
        onClearFilters = viewModel::clearFilters,
        onRefresh = viewModel::refresh,
        onGenreChange = viewModel::setGenre,
        onYearChange = viewModel::setYear,
        onStartCharChange = viewModel::setStartChar,
        onLoadMore = viewModel::loadMore,
        onSearch = viewModel::search,
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MoviesContent(
    uiState: MoviesUiState,
    onMediaClick: (Int) -> Unit,
    onSearchClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    onSettingsClick: () -> Unit,
    onSearchQueryChange: (String) -> Unit,
    onSortByChange: (String) -> Unit,
    onClearFilters: () -> Unit,
    onRefresh: () -> Unit,
    onGenreChange: (String?) -> Unit,
    onYearChange: (String?) -> Unit,
    onStartCharChange: (String?) -> Unit,
    onLoadMore: () -> Unit,
    onSearch: () -> Unit,
) {
    // Filter bottom sheet states
    var showGenreSheet by remember { mutableStateOf(false) }
    var showYearSheet by remember { mutableStateOf(false) }

    // Quick actions menu state
    var selectedMovie by remember { mutableStateOf<MediaItem?>(null) }
    var showQuickActions by remember { mutableStateOf(false) }

    // Year options (current year down to 1970)
    val yearOptions = remember {
        val currentYear = java.util.Calendar.getInstance().get(java.util.Calendar.YEAR)
        (currentYear downTo 1970).map { it.toString() }
    }

    // Group movies by first letter for A-Z index
    val activeLetters = remember(uiState.alphabetCounts, uiState.movies, uiState.sortBy) {
        val currentLetters = mutableSetOf<String>()
        if (uiState.sortBy == "az" && uiState.alphabetCounts.isNotEmpty()) {
            uiState.alphabetCounts.filter { it.count > 0 }.forEach { currentLetters.add(it.letter) }
        } else if (uiState.sortBy == "az") {
            uiState.movies.forEach { movie ->
                val firstChar = (movie.title ?: "Unknown").firstOrNull()?.uppercaseChar() ?: '#'
                currentLetters.add(if (firstChar in 'A'..'Z') firstChar.toString() else "#")
            }
        }
        currentLetters
    }

    // Grid state for A-Z scroll tracking
    val gridState = rememberLazyGridState()
    val activeLetter = uiState.startChar ?: "All"


    val keyboardController = androidx.compose.ui.platform.LocalSoftwareKeyboardController.current

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.nav_movies), color = NetflixWhite, fontWeight = FontWeight.Bold) },
                actions = {
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
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
        ) {
            Column(
                modifier = Modifier.fillMaxSize(),
            ) {
                // Search bar
                OutlinedTextField(
                    value = uiState.searchQuery,
                    onValueChange = onSearchQueryChange,
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp, vertical = 8.dp),
                    placeholder = { Text(stringResource(R.string.nav_movies_search), color = NetflixLightGray) },
                    leadingIcon = { Icon(LucideIcons.Search, contentDescription = "Search", tint = NetflixLightGray) },
                    trailingIcon = {
                        if (uiState.searchQuery.isNotEmpty()) {
                            IconButton(onClick = {
                                onSearchQueryChange("")
                            }) {
                                Icon(LucideIcons.Close, contentDescription = "Clear", tint = NetflixLightGray)
                            }
                        }
                    },
                    keyboardOptions = androidx.compose.foundation.text.KeyboardOptions(
                        imeAction = androidx.compose.ui.text.input.ImeAction.Search
                    ),
                    keyboardActions = androidx.compose.foundation.text.KeyboardActions(
                        onSearch = {
                            onSearch()
                            keyboardController?.hide()
                        }
                    ),
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

                // Filter chips
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .horizontalScroll(rememberScrollState())
                        .padding(horizontal = 16.dp, vertical = 8.dp),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    // Genre filter
                    FilterChip(
                        selected = uiState.selectedGenre != null,
                        onClick = { showGenreSheet = true },
                        label = { Text(uiState.selectedGenre ?: "Genre") },
                        colors = FilterChipDefaults.filterChipColors(
                            selectedContainerColor = NetflixRed,
                            selectedLabelColor = NetflixWhite,
                            containerColor = NetflixDark,
                            labelColor = NetflixLightGray,
                        ),
                    )

                    // Year filter
                    FilterChip(
                        selected = uiState.selectedYear != null,
                        onClick = { showYearSheet = true },
                        label = { Text(uiState.selectedYear ?: "Year") },
                        colors = FilterChipDefaults.filterChipColors(
                            selectedContainerColor = NetflixRed,
                            selectedLabelColor = NetflixWhite,
                            containerColor = NetflixDark,
                            labelColor = NetflixLightGray,
                        ),
                    )

                    // Sort chips
                    listOf("newest" to "Newest", "oldest" to "Oldest", "rating" to "Rating", "az" to "A-Z").forEach { (value, label) ->
                        FilterChip(
                            selected = uiState.sortBy == value,
                            onClick = { onSortByChange(value) },
                            label = { Text(label) },
                            colors = FilterChipDefaults.filterChipColors(
                                selectedContainerColor = NetflixRed,
                                selectedLabelColor = NetflixWhite,
                                containerColor = NetflixDark,
                                labelColor = NetflixLightGray,
                            ),
                        )
                    }

                    // Clear filters
                    if (uiState.selectedGenre != null || uiState.selectedYear != null) {
                        FilterChip(
                            selected = false,
                            onClick = onClearFilters,
                            label = { Text(stringResource(R.string.action_clear)) },
                            colors = FilterChipDefaults.filterChipColors(
                                selectedContainerColor = NetflixRed,
                                selectedLabelColor = NetflixWhite,
                                containerColor = NetflixDark,
                                labelColor = NetflixLightGray,
                            ),
                        )
                    }
                }

                // Movies count
                Text(
                    text = "${uiState.movies.size} ${if (uiState.movies.size == 1) "movie" else "movies"}",
                    color = NetflixLightGray,
                    fontSize = 14.sp,
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 4.dp),
                )

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
                                onClick = onRefresh,
                                colors = ButtonDefaults.buttonColors(containerColor = NetflixRed),
                            ) {
                                Text(stringResource(R.string.action_retry))
                            }
                        }
                    }
                } else if (uiState.movies.isEmpty()) {
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text(
                            text = "No movies found",
                            color = NetflixLightGray,
                            fontSize = 16.sp,
                        )
                    }
                } else {
                    PullToRefreshBox(
                        isRefreshing = uiState.isLoading,
                        onRefresh = onRefresh,
                        modifier = Modifier.fillMaxSize()
                    ) {
                        // Movies grid
                        LazyVerticalGrid(
                            columns = GridCells.Adaptive(minSize = 100.dp),
                            state = gridState,
                            contentPadding = PaddingValues(
                                start = 16.dp,
                                end = if (uiState.sortBy == "az") 48.dp else 16.dp, // Extra padding for A-Z sidebar
                                top = 8.dp,
                                bottom = 80.dp,
                            ),
                        horizontalArrangement = Arrangement.spacedBy(12.dp),
                        verticalArrangement = Arrangement.spacedBy(16.dp),
                    ) {
                        itemsIndexed(uiState.movies, key = { _, movie -> movie.id }) { index, movie ->
                            // Trigger pagination when reaching near the end
                            if (index >= uiState.movies.size - 40 && !uiState.isLoadingMore && !uiState.hasReachedMax) {
                                LaunchedEffect(uiState.movies.size) {
                                    onLoadMore()
                                }
                            }

                            MovieCard(
                                item = movie,
                                onClick = { onMediaClick(movie.id) },
                                onLongPress = {
                                    selectedMovie = movie
                                    showQuickActions = true
                                },
                            )
                        }

                        if (uiState.isLoadingMore) {
                            item(span = { androidx.compose.foundation.lazy.grid.GridItemSpan(maxLineSpan) }) {
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(16.dp),
                                    contentAlignment = Alignment.Center
                                ) {
                                    CircularProgressIndicator(color = NetflixRed, modifier = Modifier.size(24.dp))
                                }
                            }
                        }
                        }

                    }
                }
            }

            // A-Z Sidebar (only when sort = az)
            if (uiState.sortBy == "az" && (uiState.movies.isNotEmpty() || uiState.alphabetCounts.isNotEmpty())) {
                Column(
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .padding(end = 8.dp)
                        .verticalScroll(rememberScrollState()),
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    ALPHABET.forEach { char ->
                        val letter = char.toString()
                        val isActive = activeLetters.contains(letter)
                        val isSelected = (activeLetter == letter) || (activeLetter == "All" && activeLetters.isEmpty() && letter == "#") // Basic selected handling

                        if (isActive || activeLetter == letter) {
                            Text(
                                text = letter,
                                color = if (activeLetter == letter) NetflixRed else NetflixLightGray.copy(alpha = 0.6f),
                                fontSize = 12.sp,
                                fontWeight = if (activeLetter == letter) FontWeight.Bold else FontWeight.Normal,
                                modifier = Modifier
                                    .width(32.dp)
                                    .height(24.dp)
                                    .clip(RoundedCornerShape(4.dp))
                                    .background(
                                        if (activeLetter == letter) NetflixRed.copy(alpha = 0.2f)
                                        else NetflixBlack.copy(alpha = 0.5f)
                                    )
                                    .clickable {
                                        if (activeLetter == letter && activeLetter != "All") {
                                            onStartCharChange(null) // reset to All
                                        } else {
                                            onStartCharChange(letter)
                                        }
                                    },
                                textAlign = TextAlign.Center,
                            )
                        }
                    }
                }
            }
        }
    }

    // Genre filter bottom sheet
    if (showGenreSheet) {
        FilterBottomSheet(
            title = "Select Genre",
            options = uiState.genres,
            selectedValue = uiState.selectedGenre,
            onSelect = { genre ->
                onGenreChange(genre)
                showGenreSheet = false
            },
            onDismiss = { showGenreSheet = false },
        )
    }

    // Year filter bottom sheet
    if (showYearSheet) {
        FilterBottomSheet(
            title = "Select Year",
            options = yearOptions,
            selectedValue = uiState.selectedYear,
            onSelect = { year ->
                onYearChange(year)
                showYearSheet = false
            },
            onDismiss = { showYearSheet = false },
        )
    }

    // Quick actions menu
    QuickActionsMenu(
        title = selectedMovie?.title ?: "Options",
        visible = showQuickActions,
        onPlay = {
            selectedMovie?.let { onMediaClick(it.id) }
        },
        onViewDetails = {
            selectedMovie?.let { onMediaClick(it.id) }
        },
        onDismiss = { showQuickActions = false },
    )
}

@Composable
fun MovieCard(
    item: MediaItem,
    onClick: () -> Unit,
    onLongPress: () -> Unit = {},
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
    ) {
        Surface(
            modifier = Modifier
                .fillMaxWidth()
                .aspectRatio(2f / 3f)
                .clickable(onClick = onClick),
            color = NetflixGray,
            shape = RoundedCornerShape(8.dp),
        ) {
            Box(modifier = Modifier.fillMaxSize()) {
                if (item.poster != null || item.posterPath != null) {
                    if (item.poster != null) {
                        ResponsiveImage(
                            data = item.poster,
                            contentDescription = item.title,
                            modifier = Modifier.fillMaxSize(),
                            contentScale = ContentScale.Crop,
                        )
                    } else {
                        AsyncImage(
                            model = item.posterPath,
                            contentDescription = item.title,
                            modifier = Modifier.fillMaxSize(),
                            contentScale = ContentScale.Crop,
                        )
                    }
                } else {
                    Column(
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(16.dp),
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.Center,
                    ) {
                        Icon(
                            imageVector = com.velox.app.presentation.ui.components.LucideIcons.Film,
                            contentDescription = null,
                            tint = NetflixLightGray,
                            modifier = Modifier.size(48.dp)
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = item.title,
                            color = NetflixLightGray,
                            fontSize = 14.sp,
                            maxLines = 2,
                            overflow = TextOverflow.Ellipsis,
                            textAlign = TextAlign.Center,
                        )
                    }
                }

                // Long press indicator
                Box(
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(4.dp)
                        .size(24.dp)
                        .clip(RoundedCornerShape(4.dp))
                        .background(NetflixBlack.copy(alpha = 0.6f))
                        .clickable(onClick = onLongPress),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        text = "⋮",
                        color = NetflixWhite,
                        fontSize = 16.sp,
                    )
                }

                // Media Type Badge
                com.velox.app.presentation.ui.components.MediaBadge(
                    mediaType = item.mediaType,
                    modifier = Modifier.align(Alignment.TopStart)
                )

                // Progress bar + percentage
                val position = item.position ?: 0f
                val duration = item.duration ?: 0f
                val hasProgress = position > 0 && duration > 0 && item.completed != true
                val isCompleted = item.completed == true
                if (hasProgress || isCompleted) {
                    val progressPercent = if (isCompleted) 100f else (position / duration * 100f).coerceIn(0f, 100f)
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .align(Alignment.BottomCenter)
                            .background(androidx.compose.ui.graphics.Color.Black.copy(alpha = 0.6f))
                            .padding(horizontal = 8.dp, vertical = 6.dp),
                    ) {
                        Column {
                            LinearProgressIndicator(
                                progress = { progressPercent / 100f },
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .height(3.dp),
                                color = NetflixRed,
                                trackColor = androidx.compose.ui.graphics.Color(0xFF4B5563),
                                drawStopIndicator = {}
                            )
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(
                                text = "${progressPercent.toInt()}% watched",
                                color = androidx.compose.ui.graphics.Color(0xFFD1D5DB),
                                fontSize = 11.sp,
                            )
                        }
                    }
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
fun MoviesScreenPreview() {
    VeloxTheme {
        MoviesContent(
            uiState = MoviesUiState(
                movies = listOf(SampleMovieItem),
                isLoading = false,
                genres = listOf("Action", "Comedy", "Drama", "Horror", "Sci-Fi", "Thriller"),
            ),
            onMediaClick = {},
            onSearchClick = {},
            onNotificationsClick = {},
            onSettingsClick = {},
            onSearchQueryChange = {},
            onSortByChange = {},
            onClearFilters = {},
            onRefresh = {},
            onGenreChange = {},
            onYearChange = {},
            onLoadMore = {},
            onStartCharChange = {},
            onSearch = {},
        )
    }
}

@Preview(showBackground = true)
@Composable
fun MovieCardPreview() {
    VeloxTheme {
        MovieCard(
            item = SampleMovieItem,
            onClick = {}
        )
    }
}

private val SampleMovieItem = MediaItem(
    id = 1,
    title = "Inception",
    posterPath = null,
    backdropPath = null,
    year = 2010,
    rating = 8.8f,
    mediaType = "movie",
    overview = "A thief who steals corporate secrets through dream-sharing technology.",

)
