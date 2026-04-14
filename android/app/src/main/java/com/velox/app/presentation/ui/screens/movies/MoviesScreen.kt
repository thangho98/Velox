package com.velox.app.presentation.ui.screens.movies

import androidx.compose.ui.res.stringResource
import com.velox.app.R
import androidx.compose.foundation.background
import com.velox.app.presentation.ui.components.LucideIcons
import com.velox.app.presentation.ui.components.NotificationBell
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.coroutines.launch
import androidx.hilt.navigation.compose.hiltViewModel
import coil.compose.AsyncImage
import com.velox.app.domain.model.MediaItem
import com.velox.app.presentation.ui.components.FilterBottomSheet
import com.velox.app.presentation.ui.components.QuickActionsMenu
import com.velox.app.presentation.viewmodel.MoviesUiState
import com.velox.app.presentation.viewmodel.MoviesViewModel
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixGray
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.VeloxTheme
import com.velox.app.ui.theme.NetflixWhite

private val ALPHABET = ('A'..'Z').toList() + '#'

@Composable
fun MoviesScreen(
    onMediaClick: (Int) -> Unit,
    onSearchClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    onSettingsClick: () -> Unit,
    viewModel: MoviesViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()

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

    // Grid state for A-Z scroll tracking
    val gridState = rememberLazyGridState()
    val gridScope = rememberCoroutineScope()
    var activeLetter by remember { mutableStateOf<Char?>(null) }

    // Group movies by first letter for A-Z index
    val groupedMovies = remember(uiState.movies) {
        uiState.movies.groupBy { movie ->
            val firstChar = (movie.title ?: "Unknown").firstOrNull()?.uppercaseChar() ?: '#'
            if (firstChar in 'A'..'Z') firstChar else '#'
        }
    }

    // Calculate active letter based on scroll position
    LaunchedEffect(gridState.firstVisibleItemIndex) {
        if (uiState.sortBy == "az") {
            val visibleIndex = gridState.firstVisibleItemIndex
            if (visibleIndex < uiState.movies.size && visibleIndex >= 0) {
                val movie = uiState.movies.getOrNull(visibleIndex)
                movie?.let {
                    val firstChar = (it.title ?: "Unknown").firstOrNull()?.uppercaseChar() ?: '#'
                    activeLetter = if (firstChar in 'A'..'Z') firstChar else '#'
                }
            }
        }
    }

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
                            MovieCard(
                                item = movie,
                                onClick = { onMediaClick(movie.id) },
                                onLongPress = {
                                    selectedMovie = movie
                                    showQuickActions = true
                                },
                            )
                        }
                    }
                }
            }

            // A-Z Sidebar (only when sort = az)
            if (uiState.sortBy == "az" && uiState.movies.isNotEmpty()) {
                val activeLetters = remember(uiState.movies) {
                    ALPHABET.filter { letter ->
                        groupedMovies.containsKey(letter)
                    }
                }

                Column(
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .padding(end = 8.dp)
                        .verticalScroll(rememberScrollState(), enabled = false),
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    activeLetters.forEach { letter ->
                        val isActive = activeLetter == letter
                        Text(
                            text = letter.toString(),
                            color = if (isActive) NetflixRed else NetflixLightGray.copy(alpha = 0.6f),
                            fontSize = 12.sp,
                            fontWeight = if (isActive) FontWeight.Bold else FontWeight.Normal,
                            modifier = Modifier
                                .width(32.dp)
                                .height(24.dp)
                                .clip(RoundedCornerShape(4.dp))
                                .background(
                                    if (isActive) NetflixRed.copy(alpha = 0.2f)
                                    else NetflixBlack.copy(alpha = 0.5f)
                                )
                                .clickable {
                                    activeLetter = letter
                                    // Find first movie with this letter
                                    val firstIndex = uiState.movies.indexOfFirst { movie ->
                                        val firstChar = (movie.title ?: "Unknown").firstOrNull()?.uppercaseChar() ?: '#'
                                        val movieLetter = if (firstChar in 'A'..'Z') firstChar else '#'
                                        movieLetter == letter
                                    }
                                    if (firstIndex >= 0) {
                                        // Scroll to position
                                        gridScope.launch {
                                            gridState.scrollToItem(firstIndex)
                                        }
                                    }
                                },
                            textAlign = TextAlign.Center,
                        )
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
