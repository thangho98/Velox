package com.velox.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.Text
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import com.velox.app.presentation.tv.TvAppNavigation
import com.velox.app.domain.repository.AuthRepository
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class TvMainActivity : ComponentActivity() {

    @Inject
    lateinit var authRepository: AuthRepository

    @OptIn(ExperimentalTvMaterial3Api::class)
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            val isLoggedIn by authRepository.isLoggedIn().collectAsState(initial = null)

            // TODO: Wrap with TvTheme once defined
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(Color.Black),
                contentAlignment = Alignment.Center
            ) {
                if (isLoggedIn != null) {
                    val startDestination = if (isLoggedIn == true) "home" else "login"
                    TvAppNavigation(startDestination = startDestination)
                }
            }
        }
    }
}
