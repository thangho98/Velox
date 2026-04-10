package com.velox.app.presentation.ui.components

import androidx.compose.foundation.clickable
import com.velox.app.presentation.ui.components.LucideIcons
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.velox.app.ui.theme.NetflixBlack
import com.velox.app.ui.theme.NetflixLightGray
import com.velox.app.ui.theme.NetflixRed
import com.velox.app.ui.theme.NetflixWhite
import com.velox.app.ui.theme.VeloxTheme

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FilterBottomSheet(
    title: String,
    options: List<String>,
    selectedValue: String?,
    onSelect: (String?) -> Unit,
    onDismiss: () -> Unit,
) {
    ModalBottomSheet(
        onDismissRequest = onDismiss,
        containerColor = NetflixBlack,
        contentColor = NetflixWhite,
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 32.dp),
        ) {
            // Header
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 12.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    text = title,
                    color = NetflixWhite,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.SemiBold,
                )
                if (selectedValue != null) {
                    TextButton(onClick = { onSelect(null) }) {
                        Text(
                            text = "Clear",
                            color = NetflixRed,
                            fontSize = 14.sp,
                        )
                    }
                }
            }

            HorizontalDivider(color = NetflixLightGray.copy(alpha = 0.3f))

            // Options list
            LazyColumn(
                modifier = Modifier.fillMaxWidth(),
            ) {
                items(options) { option ->
                    val isSelected = option == selectedValue
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable { onSelect(option) }
                            .padding(horizontal = 16.dp, vertical = 14.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(
                            text = option,
                            color = if (isSelected) NetflixRed else NetflixWhite,
                            fontSize = 16.sp,
                            fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal,
                        )
                        if (isSelected) {
                            Icon(
                                LucideIcons.Check,
                                contentDescription = "Selected",
                                tint = NetflixRed,
                                modifier = Modifier.size(20.dp),
                            )
                        }
                    }
                }
            }
        }
    }
}

// Note: ModalBottomSheet requires runtime context and cannot render in IDE preview.
// Use the Composable Preview with a device/emulator to test this component.
@Preview(showBackground = true, backgroundColor = 0xFF000000)
@Composable
private fun FilterBottomSheetPreview() {
    VeloxTheme {
        FilterBottomSheet(
            title = "Sort By",
            options = listOf("Title", "Year", "Rating", "Duration", "Recently Added"),
            selectedValue = "Rating",
            onSelect = {},
            onDismiss = {},
        )
    }
}
