package com.velox.app.presentation.ui.components

import androidx.compose.foundation.layout.*
import com.velox.app.presentation.ui.components.LucideIcons
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

sealed class ActionMenuItem {
    abstract val onClick: () -> Unit

    data class Play(override val onClick: () -> Unit) : ActionMenuItem()
    data class Favorite(val isFavorite: Boolean, override val onClick: () -> Unit) : ActionMenuItem()
    data class AddToList(override val onClick: () -> Unit) : ActionMenuItem()
    data class MarkWatched(val isWatched: Boolean, override val onClick: () -> Unit) : ActionMenuItem()
    data class StreamUrl(override val onClick: () -> Unit) : ActionMenuItem()
    data class RefreshMetadata(override val onClick: () -> Unit) : ActionMenuItem()
    data class Info(override val onClick: () -> Unit) : ActionMenuItem()
    data class Edit(override val onClick: () -> Unit) : ActionMenuItem()
    data object Separator : ActionMenuItem() {
        override val onClick: () -> Unit = {}
    }
    data class Custom(
        val label: String,
        val icon: ImageVector,
        override val onClick: () -> Unit,
        val danger: Boolean = false,
        val isLoading: Boolean = false,
        val autoDismiss: Boolean = true,
    ) : ActionMenuItem()
}

@Composable
fun ActionMenu(
    expanded: Boolean,
    onDismiss: () -> Unit,
    items: List<ActionMenuItem>,
    modifier: Modifier = Modifier,
) {
    DropdownMenu(
        expanded = expanded,
        onDismissRequest = onDismiss,
        modifier = modifier,
    ) {
        items.forEach { item ->
            when (item) {
                is ActionMenuItem.Separator -> {
                    HorizontalDivider(
                        color = Color(0xFF333333),
                        modifier = Modifier.padding(vertical = 4.dp),
                    )
                }
                is ActionMenuItem.Play -> ActionDropdownItem(LucideIcons.PlayArrow, "Play", item.onClick, onDismiss)
                is ActionMenuItem.Favorite -> ActionDropdownItem(
                    LucideIcons.Favorite,
                    if (item.isFavorite) "Remove from Favorites" else "Add to Favorites",
                    item.onClick, onDismiss,
                    iconTint = if (item.isFavorite) NetflixRed else Color(0xFFE5E5E5),
                )
                is ActionMenuItem.AddToList -> ActionDropdownItem(LucideIcons.Add, "Add to List", item.onClick, onDismiss)
                is ActionMenuItem.MarkWatched -> ActionDropdownItem(
                    LucideIcons.Check,
                    if (item.isWatched) "Mark as Unwatched" else "Mark as Watched",
                    item.onClick, onDismiss,
                )
                is ActionMenuItem.StreamUrl -> ActionDropdownItem(LucideIcons.Link, "Copy stream URL", item.onClick, onDismiss)
                is ActionMenuItem.RefreshMetadata -> ActionDropdownItem(LucideIcons.Refresh, "Refresh metadata", item.onClick, onDismiss)
                is ActionMenuItem.Info -> ActionDropdownItem(LucideIcons.Info, "Info", item.onClick, onDismiss)
                is ActionMenuItem.Edit -> ActionDropdownItem(LucideIcons.Edit, "Edit metadata", item.onClick, onDismiss)
                is ActionMenuItem.Custom -> ActionDropdownItem(
                    item.icon, item.label, item.onClick, onDismiss,
                    danger = item.danger, isLoading = item.isLoading, autoDismiss = item.autoDismiss
                )
            }
        }
    }
}

@Composable
private fun ActionDropdownItem(
    icon: ImageVector,
    label: String,
    onClick: () -> Unit,
    onDismiss: () -> Unit,
    danger: Boolean = false,
    isLoading: Boolean = false,
    autoDismiss: Boolean = true,
    iconTint: Color = Color(0xFFE5E5E5),
) {
    DropdownMenuItem(
        text = {
            Text(
                text = label,
                color = if (danger) NetflixRed else Color(0xFFE5E5E5),
                fontSize = 15.sp,
            )
        },
        leadingIcon = {
            if (isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.size(20.dp),
                    color = if (danger) NetflixRed else iconTint,
                    strokeWidth = 2.dp
                )
            } else {
                Icon(
                    imageVector = icon,
                    contentDescription = null,
                    tint = if (danger) NetflixRed else iconTint,
                    modifier = Modifier.size(20.dp),
                )
            }
        },
        onClick = {
            onClick()
            if (autoDismiss) onDismiss()
        },
    )
}

@Composable
fun ActionMenuButton(
    expanded: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    IconButton(
        onClick = onClick,
        modifier = modifier,
    ) {
        Icon(
            imageVector = if (expanded) LucideIcons.Close else LucideIcons.MoreVert,
            contentDescription = "More options",
            tint = NetflixWhite,
        )
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF1A1A1A)
@Composable
private fun ActionMenuButtonCollapsedPreview() {
    VeloxTheme {
        ActionMenuButton(
            expanded = false,
            onClick = {},
        )
    }
}

@Preview(showBackground = true, backgroundColor = 0xFF1A1A1A)
@Composable
private fun ActionMenuButtonExpandedPreview() {
    VeloxTheme {
        ActionMenuButton(
            expanded = true,
            onClick = {},
        )
    }
}
