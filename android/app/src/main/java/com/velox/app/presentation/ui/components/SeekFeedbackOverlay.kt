package com.velox.app.presentation.ui.components

import androidx.compose.animation.core.*
import com.velox.app.presentation.ui.components.LucideIcons
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ChevronLeft
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.foundation.shape.GenericShape
import androidx.compose.foundation.background
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.ui.theme.VeloxTheme

/**
 * SeekFeedbackOverlay — YouTube-style double-tap seek feedback
 *
 * Shows directional arrows + seek amount on double-tap left/right edges.
 * Used in VideoPlayer for gesture-based seeking.
 */
@Composable
fun SeekFeedbackOverlay(
    side: String, // "left" or "right"
    amount: Int,  // seconds to seek (positive = forward, negative = backward)
    modifier: Modifier = Modifier,
) {
    // Simple fade animation for arrows
    val infiniteTransition = rememberInfiniteTransition(label = "seek")

    val alpha1 by infiniteTransition.animateFloat(
        initialValue = 0.3f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(300, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse,
        ),
        label = "alpha1",
    )
    val alpha2 by infiniteTransition.animateFloat(
        initialValue = 0.3f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(300, easing = LinearEasing, delayMillis = 100),
            repeatMode = RepeatMode.Reverse,
        ),
        label = "alpha2",
    )
    val alpha3 by infiniteTransition.animateFloat(
        initialValue = 0.3f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(300, easing = LinearEasing, delayMillis = 200),
            repeatMode = RepeatMode.Reverse,
        ),
        label = "alpha3",
    )

    val arrows = if (side == "right") {
        listOf(
            Triple(LucideIcons.ChevronRight, alpha1, 0.dp),
            Triple(LucideIcons.ChevronRight, alpha2, (-10).dp),
            Triple(LucideIcons.ChevronRight, alpha3, (-20).dp),
        )
    } else {
        listOf(
            Triple(LucideIcons.ChevronLeft, alpha3, 0.dp),
            Triple(LucideIcons.ChevronLeft, alpha2, (-10).dp),
            Triple(LucideIcons.ChevronLeft, alpha1, (-20).dp),
        )
    }

    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
        modifier = modifier
    ) {
        Row {
            arrows.forEach { (icon, alpha, offsetValue) ->
                Box(
                    modifier = Modifier.offset(x = offsetValue),
                ) {
                    Icon(
                        imageVector = icon,
                        contentDescription = null,
                        tint = Color.White,
                        modifier = Modifier
                            .size(28.dp)
                            .alpha(alpha),
                    )
                }
            }
        }

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = "${if (side == "left") "-" else "+"}${amount}s",
            color = Color.White,
            fontSize = 16.sp,
            fontWeight = FontWeight.Bold,
        )
    }
}

@Composable
fun SeekFeedbackOverlayContainer(
    visible: Boolean,
    side: String,
    amount: Int,
    modifier: Modifier = Modifier,
) {
    androidx.compose.animation.AnimatedVisibility(
        visible = visible,
        enter = androidx.compose.animation.fadeIn(androidx.compose.animation.core.tween(200)),
        exit = androidx.compose.animation.fadeOut(androidx.compose.animation.core.tween(400)),
        modifier = modifier.fillMaxSize()
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
            val isLeft = side == "left"
            val align = if (isLeft) Alignment.CenterStart else Alignment.CenterEnd
            

            SeekFeedbackOverlay(
                side = side,
                amount = amount,
                modifier = Modifier.align(align).padding(horizontal = 64.dp),
            )
        }
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun SeekFeedbackRightPreview() {
    VeloxTheme {
        SeekFeedbackOverlay(
            side = "right",
            amount = 10,
        )
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun SeekFeedbackLeftPreview() {
    VeloxTheme {
        SeekFeedbackOverlay(
            side = "left",
            amount = 15,
        )
    }
}
