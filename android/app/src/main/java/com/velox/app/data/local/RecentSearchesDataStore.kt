package com.velox.app.data.local

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringSetPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "velox_prefs")

@Singleton
class RecentSearchesDataStore @Inject constructor(
    @ApplicationContext private val context: Context,
) {
    private val recentSearchesKey = stringSetPreferencesKey("recent_searches")
    private val maxRecentSearches = 10

    val recentSearches: Flow<List<String>> = context.dataStore.data.map { preferences ->
        preferences[recentSearchesKey]?.toList() ?: emptyList()
    }

    suspend fun addSearch(query: String) {
        context.dataStore.edit { preferences ->
            val current = preferences[recentSearchesKey]?.toMutableSet() ?: mutableSetOf()
            // Remove if exists to avoid duplicates, then add to front
            current.remove(query)
            val newList = listOf(query) + current.toList()
            preferences[recentSearchesKey] = newList.take(maxRecentSearches).toSet()
        }
    }

    suspend fun removeSearch(query: String) {
        context.dataStore.edit { preferences ->
            val current = preferences[recentSearchesKey]?.toMutableSet() ?: mutableSetOf()
            current.remove(query)
            preferences[recentSearchesKey] = current
        }
    }

    suspend fun clearAll() {
        context.dataStore.edit { preferences ->
            preferences.remove(recentSearchesKey)
        }
    }
}
