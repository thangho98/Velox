package com.velox.app.di

import android.content.Context
import com.velox.app.BuildConfig
import com.velox.app.data.api.AuthInterceptor
import com.velox.app.data.api.AuthManager
import com.velox.app.data.api.ServerPrefsManager
import com.velox.app.data.api.TokenAuthenticator
import com.velox.app.data.api.VeloxApi
import com.velox.app.data.api.VeloxApiProvider
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import java.util.concurrent.TimeUnit
import javax.inject.Singleton
import kotlinx.serialization.json.Json

@Module
@InstallIn(SingletonComponent::class)
object NetworkModule {

    @Provides
    @Singleton
    fun provideJson(): Json = Json {
        ignoreUnknownKeys = true
        coerceInputValues = true
        isLenient = true
    }

    @Provides
    @Singleton
    fun provideServerPrefsManager(@ApplicationContext context: Context): ServerPrefsManager {
        return ServerPrefsManager(context)
    }

    @Provides
    @Singleton
    fun provideAuthManager(@ApplicationContext context: Context): AuthManager {
        return AuthManager(context)
    }

    @Provides
    @Singleton
    fun provideAuthInterceptor(authManager: AuthManager): AuthInterceptor {
        return AuthInterceptor(authManager)
    }

    @Provides
    @Singleton
    fun provideTokenAuthenticator(
        authManager: AuthManager,
        apiProvider: javax.inject.Provider<VeloxApiProvider>
    ): TokenAuthenticator {
        return TokenAuthenticator(authManager, apiProvider)
    }

    @Provides
    @Singleton
    fun provideOkHttpClient(
        authInterceptor: AuthInterceptor,
        tokenAuthenticator: TokenAuthenticator
    ): OkHttpClient {
        val loggingInterceptor = HttpLoggingInterceptor().apply {
            level = if (BuildConfig.DEBUG) {
                HttpLoggingInterceptor.Level.BODY
            } else {
                HttpLoggingInterceptor.Level.NONE
            }
        }

        return OkHttpClient.Builder()
            .addInterceptor(authInterceptor)
            .authenticator(tokenAuthenticator)
            .addInterceptor(loggingInterceptor)
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .build()
    }

    @Provides
    @Singleton
    fun provideVeloxApiProvider(
        serverPrefsManager: ServerPrefsManager,
        okHttpClient: OkHttpClient,
        json: Json,
    ): VeloxApiProvider {
        return VeloxApiProvider(serverPrefsManager, okHttpClient, json)
    }

    @Provides
    fun provideVeloxApi(apiProvider: VeloxApiProvider): VeloxApi {
        return apiProvider.getApi()
    }
}
