package com.velox.app.data.api

import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import kotlinx.coroutines.runBlocking
import okhttp3.MediaType.Companion.toMediaType
import retrofit2.Retrofit
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Provides VeloxApi with dynamic base URL.
 * Recreates the API when server URL changes.
 */
@Singleton
class VeloxApiProvider @Inject constructor(
    private val serverPrefsManager: ServerPrefsManager,
    private val okHttpClient: okhttp3.OkHttpClient,
    private val json: kotlinx.serialization.json.Json,
) {
    private var cachedApi: VeloxApi? = null
    private var cachedBaseUrl: String? = null

    fun getServerUrlSync(): String {
        return serverPrefsManager.getServerUrlSync()
    }

    fun getApi(): VeloxApi {
        val currentUrl = serverPrefsManager.getServerUrlSync()

        // Return cached API if URL hasn't changed
        if (cachedApi != null && cachedBaseUrl == currentUrl) {
            return cachedApi!!
        }

        // Create new Retrofit with current base URL
        val retrofit = Retrofit.Builder()
            .baseUrl(currentUrl)
            .client(okHttpClient)
            .addConverterFactory(json.asConverterFactory("application/json".toMediaType()))
            .build()

        cachedApi = retrofit.create(VeloxApi::class.java)
        cachedBaseUrl = currentUrl

        return cachedApi!!
    }
}
