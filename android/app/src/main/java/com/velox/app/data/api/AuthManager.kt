package com.velox.app.data.api

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.intPreferencesKey
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import com.velox.app.domain.model.ImageResource
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import okhttp3.Interceptor
import okhttp3.Response
import javax.inject.Inject
import javax.inject.Singleton

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "velox_auth")

@Singleton
class AuthManager @Inject constructor(
    @ApplicationContext private val context: Context,
) {
    companion object {
        private val ACCESS_TOKEN_KEY = stringPreferencesKey("access_token")
        private val REFRESH_TOKEN_KEY = stringPreferencesKey("refresh_token")
        private val TOKEN_EXPIRES_AT_KEY = intPreferencesKey("token_expires_at")
        private val USER_ID_KEY = intPreferencesKey("user_id")
        private val USERNAME_KEY = stringPreferencesKey("username")
        private val DISPLAY_NAME_KEY = stringPreferencesKey("display_name")
        private val IS_ADMIN_KEY = stringPreferencesKey("is_admin")
        private val PROFILE_PATH_KEY = stringPreferencesKey("profile_path")
        private val PROFILE_JSON_KEY = stringPreferencesKey("profile_json")
        private val APP_LANGUAGE_KEY = stringPreferencesKey("app_language")

        private val json = Json { ignoreUnknownKeys = true }
    }

    val isLoggedIn: Flow<Boolean> = context.dataStore.data.map { prefs ->
        prefs[ACCESS_TOKEN_KEY] != null
    }

    val accessToken: Flow<String?> = context.dataStore.data.map { prefs ->
        prefs[ACCESS_TOKEN_KEY]
    }

    val refreshToken: Flow<String?> = context.dataStore.data.map { prefs ->
        prefs[REFRESH_TOKEN_KEY]
    }

    val userId: Flow<Int?> = context.dataStore.data.map { prefs ->
        prefs[USER_ID_KEY]
    }

    val username: Flow<String?> = context.dataStore.data.map { prefs ->
        prefs[USERNAME_KEY]
    }

    val displayName: Flow<String?> = context.dataStore.data.map { prefs ->
        prefs[DISPLAY_NAME_KEY]
    }

    val isAdmin: Flow<Boolean> = context.dataStore.data.map { prefs ->
        prefs[IS_ADMIN_KEY] == "true"
    }

    val profilePath: Flow<String?> = context.dataStore.data.map { prefs ->
        prefs[PROFILE_PATH_KEY]
    }

    val profileImageResource: Flow<ImageResource?> = context.dataStore.data.map { prefs ->
        prefs[PROFILE_JSON_KEY]?.let { jsonStr ->
            try {
                val dto = json.decodeFromString<ProfileImageJson>(jsonStr)
                dto.toImageResource()
            } catch (_: Exception) {
                null
            }
        }
    }

    val appLanguage: Flow<String?> = context.dataStore.data.map { prefs ->
        prefs[APP_LANGUAGE_KEY]
    }

    fun getAccessTokenSync(): String? = runBlocking {
        accessToken.first()
    }

    fun getRefreshTokenSync(): String? = runBlocking {
        refreshToken.first()
    }

    suspend fun saveTokens(
        accessToken: String,
        refreshToken: String,
        expiresIn: Int,
    ) {
        val expiresAt = (System.currentTimeMillis() / 1000) + expiresIn
        context.dataStore.edit { prefs ->
            prefs[ACCESS_TOKEN_KEY] = accessToken
            prefs[REFRESH_TOKEN_KEY] = refreshToken
            prefs[TOKEN_EXPIRES_AT_KEY] = expiresAt.toInt()
        }
    }

    suspend fun saveUser(
        userId: Int,
        username: String,
        displayName: String,
        isAdmin: Boolean,
        profilePath: String?,
        profileImage: ImageResource? = null,
    ) {
        context.dataStore.edit { prefs ->
            prefs[USER_ID_KEY] = userId
            prefs[USERNAME_KEY] = username
            prefs[DISPLAY_NAME_KEY] = displayName
            prefs[IS_ADMIN_KEY] = isAdmin.toString()
            if (profilePath != null) {
                prefs[PROFILE_PATH_KEY] = profilePath
            } else {
                prefs.remove(PROFILE_PATH_KEY)
            }
            if (profileImage != null) {
                prefs[PROFILE_JSON_KEY] = json.encodeToString(
                    ProfileImageJson.fromImageResource(profileImage)
                )
            } else {
                prefs.remove(PROFILE_JSON_KEY)
            }
        }
    }

    suspend fun saveAppLanguage(lang: String) {
        context.dataStore.edit { prefs ->
            prefs[APP_LANGUAGE_KEY] = lang
        }
    }

    suspend fun clearAuth() {
        context.dataStore.edit { prefs ->
            prefs.clear()
        }
    }

    suspend fun isTokenExpired(): Boolean {
        val expiresAt = context.dataStore.data.first()[TOKEN_EXPIRES_AT_KEY] ?: return true
        val now = (System.currentTimeMillis() / 1000).toInt()
        return now >= expiresAt - 60 // 60 second buffer
    }
}

class AuthInterceptor @Inject constructor(
    private val authManager: AuthManager,
) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val originalRequest = chain.request()

        // Skip auth for login and refresh endpoints
        val path = originalRequest.url.encodedPath
        if (path.contains("/auth/login") || path.contains("/auth/refresh")) {
            return chain.proceed(originalRequest)
        }

        val token = authManager.getAccessTokenSync()
        if (token == null) {
            return chain.proceed(originalRequest)
        }

        val authenticatedRequest = originalRequest.newBuilder()
            .header("Authorization", "Bearer $token")
            .build()

        return chain.proceed(authenticatedRequest)
    }
}
