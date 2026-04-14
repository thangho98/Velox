package com.velox.app.presentation.ui.screens.series

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
import com.velox.app.domain.model.SeriesItem
import com.velox.app.presentation.ui.components.FilterBottomSheet
import com.velox.app.presentation.ui.components.QuickActionsMenu
import com.velox.app.presentation.viewmodel.SeriesUiState
import com.velox.app.presentation.viewmodel.SeriesViewModel
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixDark
import com.velox.app.ui.theme.NetflixGray
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

private val ALPHABET = ('A'..'Z').toList() + '#'

@Composable
fun SeriesScreen(
    onSeriesClick: (Int) -> Unit,
    onSearchClick: () -> Unit,
    onNotificationsClick: () -> Unit,
    onSettingsClick: () -> Unit,
    viewModel: SeriesViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()

    SeriesContent(
        uiState = uiState,
        onSeriesClick = onSeriesClick,
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
fun SeriesContent(
    uiState: SeriesUiState,
    onSeriesClick: (Int) -> Unit,
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
    var selectedSeries by remember { mutableStateOf<SeriesItem?>(null) }
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

    // Group series by first letter for A-Z index
    val groupedSeries = remember(uiState.series) {
        uiState.series.groupBy { series ->
            val firstChar = (series.title ?: "Unknown").firstOrNull()?.uppercaseChar() ?: '#'
            if (firstChar in 'A'..'Z') firstChar else '#'
        }
    }

    // Calculate active letter based on scroll position
    LaunchedEffect(gridState.firstVisibleItemIndex) {
        if (uiState.sortBy == "az") {
            val visibleIndex = gridState.firstVisibleItemIndex
            if (visibleIndex < uiState.series.size && visibleIndex >= 0) {
                val series = uiState.series.getOrNull(visibleIndex)
                series?.let {
                    val firstChar = (it.title ?: "Unknown").firstOrNull()?.uppercaseChar() ?: '#'
                    activeLetter = if (firstChar in 'A'..'Z') firstChar else '#'
                }
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Series", color = NetflixWhite, fontWeight = FontWeight.Bold) },
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
                    placeholder = { Text("Search series...", color = NetflixLightGray) },
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

                // Series count
                Text(
                    text = "${uiState.series.size} ${if (uiState.series.size == 1) "series" else "series"}",
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
                } else if (uiState.series.isEmpty()) {
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text(
                            text = "No series found",
                            color = NetflixLightGray,
                            fontSize = 16.sp,
                        )
                    }
                } else {
                    // Series grid
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
                        itemsIndexed(uiState.series, key = { _, series -> series.id }) { index, series ->
                            SeriesCard(
                                item = series,
                                onClick = { onSeriesClick(series.id) },
                                onLongPress = {
                                    selectedSeries = series
                                    showQuickActions = true
                                },
                            )
                        }
                    }
                }
            }

            // A-Z Sidebar (only when sort = az)
            if (uiState.sortBy == "az" && uiState.series.isNotEmpty()) {
                val activeLetters = remember(uiState.series) {
                    ALPHABET.filter { letter ->
                        groupedSeries.containsKey(letter)
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
                                    // Find first series with this letter
                                    val firstIndex = uiState.series.indexOfFirst { s ->
                                        val firstChar = (s.title ?: "Unknown").firstOrNull()?.uppercaseChar() ?: '#'
                                        val seriesLetter = if (firstChar in 'A'..'Z') firstChar else '#'
                                        seriesLetter == letter
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
        title = selectedSeries?.title ?: "Options",
        visible = showQuickActions,
        onPlay = {
            selectedSeries?.let { onSeriesClick(it.id) }
        },
        onViewDetails = {
            selectedSeries?.let { onSeriesClick(it.id) }
        },
        onDismiss = { showQuickActions = false },
    )
}

@Composable
fun SeriesCard(
    item: SeriesItem,
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
        item.seasonCount?.let {
            Text(
                text = "$it season${if (it > 1) "s" else ""}",
                color = NetflixLightGray,
                fontSize = 12.sp,
            )
        }
    }
}

@Preview(showBackground = true)
@Composable
fun SeriesScreenPreview() {
    VeloxTheme {
        SeriesContent(
            uiState = SeriesUiState(
                series = listOf(SampleSeriesItem),
                isLoading = false,
                genres = listOf("Action", "Comedy", "Drama", "Horror", "Sci-Fi", "Thriller"),
            ),
            onSeriesClick = {},
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
fun SeriesCardPreview() {
    VeloxTheme {
        SeriesCard(
            item = SampleSeriesItem,
            onClick = {}
        )
    }
}

private val SampleSeriesItem = SeriesItem(
    id = 1,
    title = "The Last of Us",
    posterPath = null,
    backdropPath = null,
    year = 2023,
    rating = 8.8f,
    overview = "A survivors of a cordyceps infection.",
    seasonCount = 1,
    episodeCount = 9,
)
