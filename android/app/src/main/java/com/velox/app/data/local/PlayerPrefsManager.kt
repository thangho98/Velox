package com.velox.app.data.local

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

private val Context.playerDataStore: DataStore<Preferences> by preferencesDataStore(name = "velox_player")

@Singleton
class PlayerPrefsManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    companion object {
        private val PRIMARY_SUB_LANG = stringPreferencesKey("primary_sub_lang")
        private val SECONDARY_SUB_LANG = stringPreferencesKey("secondary_sub_lang")
    }

    val primarySubLang: Flow<String?> = context.playerDataStore.data.map { it[PRIMARY_SUB_LANG] }
    val secondarySubLang: Flow<String?> = context.playerDataStore.data.map { it[SECONDARY_SUB_LANG] }

    suspend fun setPrimarySubLang(lang: String?) {
        context.playerDataStore.edit { prefs ->
            if (lang != null) prefs[PRIMARY_SUB_LANG] = lang else prefs.remove(PRIMARY_SUB_LANG)
        }
    }

    suspend fun setSecondarySubLang(lang: String?) {
        context.playerDataStore.edit { prefs ->
            if (lang != null) prefs[SECONDARY_SUB_LANG] = lang else prefs.remove(SECONDARY_SUB_LANG)
        }
    }
}
