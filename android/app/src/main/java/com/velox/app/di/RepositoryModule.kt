package com.velox.app.di

import com.velox.app.data.repository.AuthRepositoryImpl
import com.velox.app.data.repository.MediaRepositoryImpl
import com.velox.app.domain.repository.AuthRepository
import com.velox.app.domain.repository.MediaRepository
import dagger.Binds
import dagger.Module
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
abstract class RepositoryModule {

    @Binds
    @Singleton
    abstract fun bindAuthRepository(
        authRepositoryImpl: AuthRepositoryImpl,
    ): AuthRepository

    @Binds
    @Singleton
    abstract fun bindMediaRepository(
        mediaRepositoryImpl: MediaRepositoryImpl,
    ): MediaRepository

    @Binds
    @Singleton
    abstract fun bindSettingsRepository(
        settingsRepositoryImpl: com.velox.app.data.repository.SettingsRepositoryImpl,
    ): com.velox.app.domain.repository.SettingsRepository

    @Binds
    @Singleton
    abstract fun bindLiveTvRepository(
        liveTvRepositoryImpl: com.velox.app.data.repository.livetv.LiveTvRepositoryImpl,
    ): com.velox.app.domain.repository.livetv.LiveTvRepository
}
