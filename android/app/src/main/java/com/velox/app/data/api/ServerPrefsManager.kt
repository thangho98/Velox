package com.velox.app.data.api

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.runBlocking
import javax.inject.Inject
import javax.inject.Singleton

private val Context.serverDataStore: DataStore<Preferences> by preferencesDataStore(name = "velox_server")

@Singleton
class ServerPrefsManager @Inject constructor(
    @ApplicationContext private val context: Context,
) {
    companion object {
        private val SERVER_URL_KEY = stringPreferencesKey("server_url")
        private val DEFAULT_SERVER_URL = com.velox.app.BuildConfig.BASE_URL

        // Allow custom port for development
        const val DEFAULT_PORT = 8098
    }

    val serverUrl: Flow<String> = context.serverDataStore.data.map { prefs ->
        prefs[SERVER_URL_KEY] ?: DEFAULT_SERVER_URL
    }

    fun getServerUrlSync(): String = runBlocking {
        serverUrl.first()
    }

    suspend fun setServerUrl(url: String): String {
        val normalizedUrl = normalizeUrl(url)
        context.serverDataStore.edit { prefs ->
            prefs[SERVER_URL_KEY] = normalizedUrl
        }
        return normalizedUrl
    }

    suspend fun clearServerUrl() {
        context.serverDataStore.edit { prefs ->
            prefs.remove(SERVER_URL_KEY)
        }
    }

    private fun normalizeUrl(url: String): String {
        var normalized = url.trim()

        // Remove trailing slash
        if (normalized.endsWith("/")) {
            normalized = normalized.dropLast(1)
        }

        // Add /api/ suffix if not present
        if (!normalized.endsWith("/api")) {
            normalized = "$normalized/api"
        }

        // Ensure it ends with /
        if (!normalized.endsWith("/")) {
            normalized = "$normalized/"
        }

        return normalized
    }

    fun isDefaultUrl(): Boolean = getServerUrlSync() == DEFAULT_SERVER_URL
}
