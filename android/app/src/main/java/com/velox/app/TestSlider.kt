package com.velox.app
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.*
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.background
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun Test() {
    Slider(
        value = 0.5f,
        onValueChange = {},
        thumb = { sliderState -> 
            Box(Modifier.size(12.dp).background(Color.White))
        },
        track = { sliderState -> 
            SliderDefaults.Track(
                colors = SliderDefaults.colors(),
                sliderState = sliderState,
                modifier = Modifier.height(4.dp)
            )
        }
    )
}
