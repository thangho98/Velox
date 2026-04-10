package com.velox.app.presentation.ui.components

import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.PathFillType
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.StrokeJoin
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.graphics.vector.PathParser
import androidx.compose.ui.unit.dp

object LucideIcons {
    val VolumeOff: ImageVector
        get() = lucideIcon(
            name = "Lucide.VolumeOff",
            pathData = listOf(
                "M11,5 L6,9 L2,9 L2,15 L6,15 L11,19 L11,5 Z",
                "M22,9 L16,15",
                "M16,9 L22,15",
            ),
        )

    val VolumeUp: ImageVector
        get() = lucideIcon(
            name = "Lucide.VolumeUp",
            pathData = listOf(
                "M11,5 L6,9 L2,9 L2,15 L6,15 L11,19 L11,5 Z",
                "M15.54 8.46a5 5 0 0 1 0 7.07",
                "M19.07 4.93a10 10 0 0 1 0 14.14",
            ),
        )

    val ArrowBack: ImageVector
        get() = lucideIcon(
            name = "Lucide.ArrowBack",
            pathData = listOf(
                "m12 19-7-7 7-7",
                "M19 12H5",
            ),
        )

    val Logout: ImageVector
        get() = lucideIcon(
            name = "Lucide.Logout",
            pathData = listOf(
                "M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4",
                "M16,17 L21,12 L16,7",
                "M21,12 L9,12",
            ),
        )

    val Close: ImageVector
        get() = lucideIcon(
            name = "Lucide.Close",
            pathData = listOf(
                "M18 6 6 18",
                "m6 6 12 12",
            ),
        )

    val Cloud: ImageVector
        get() = lucideIcon(
            name = "Lucide.Cloud",
            pathData = listOf(
                "M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z",
            ),
        )

    val Visibility: ImageVector
        get() = lucideIcon(
            name = "Lucide.Visibility",
            pathData = listOf(
                "M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z",
                "M12.0,9.0 A3.0,3.0 0 1,0 12.0,15.0 A3.0,3.0 0 1,0 12.0,9.0 Z",
            ),
        )

    val VisibilityOff: ImageVector
        get() = lucideIcon(
            name = "Lucide.VisibilityOff",
            pathData = listOf(
                "M9.88 9.88a3 3 0 1 0 4.24 4.24",
                "M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68",
                "M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61",
                "M2,2 L22,22",
            ),
        )

    val Wifi: ImageVector
        get() = lucideIcon(
            name = "Lucide.Wifi",
            pathData = listOf(
                "M12 20h.01",
                "M2 8.82a15 15 0 0 1 20 0",
                "M5 12.859a10 10 0 0 1 14 0",
                "M8.5 16.429a5 5 0 0 1 7 0",
            ),
        )

    val WifiOff: ImageVector
        get() = lucideIcon(
            name = "Lucide.WifiOff",
            pathData = listOf(
                "M12 20h.01",
                "M8.5 16.429a5 5 0 0 1 7 0",
                "M5 12.859a10 10 0 0 1 5.17-2.69",
                "M19 12.859a10 10 0 0 0-2.007-1.523",
                "M2 8.82a15 15 0 0 1 4.177-2.643",
                "M22 8.82a15 15 0 0 0-11.288-3.764",
                "m2 2 20 20",
            ),
        )

    val Refresh: ImageVector
        get() = lucideIcon(
            name = "Lucide.Refresh",
            pathData = listOf(
                "M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8",
                "M21 3v5h-5",
                "M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16",
                "M8 16H3v5",
            ),
        )

    val Folder: ImageVector
        get() = lucideIcon(
            name = "Lucide.Folder",
            pathData = listOf(
                "M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z",
            ),
        )

    val FolderOpen: ImageVector
        get() = lucideIcon(
            name = "Lucide.FolderOpen",
            pathData = listOf(
                "m6 14 1.5-2.9A2 2 0 0 1 9.24 10H20a2 2 0 0 1 1.94 2.5l-1.54 6a2 2 0 0 1-1.95 1.5H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h3.9a2 2 0 0 1 1.69.9l.81 1.2a2 2 0 0 0 1.67.9H18a2 2 0 0 1 2 2v2",
            ),
        )

    val GridView: ImageVector
        get() = lucideIcon(
            name = "Lucide.GridView",
            pathData = listOf(
                "M5.0,3.0 H19.0 A2.0,2.0 0 0 1 21.0,5.0 V19.0 A2.0,2.0 0 0 1 19.0,21.0 H5.0 A2.0,2.0 0 0 1 3.0,19.0 V5.0 A2.0,2.0 0 0 1 5.0,3.0 Z",
                "M3 9h18",
                "M3 15h18",
                "M9 3v18",
                "M15 3v18",
            ),
        )

    val ListIcon: ImageVector
        get() = lucideIcon(
            name = "Lucide.ListIcon",
            pathData = listOf(
                "M8,6 L21,6",
                "M8,12 L21,12",
                "M8,18 L21,18",
                "M3,6 L3.01,6",
                "M3,12 L3.01,12",
                "M3,18 L3.01,18",
            ),
        )

    val Movie: ImageVector
        get() = lucideIcon(
            name = "Lucide.Movie",
            pathData = listOf(
                "M5.0,3.0 H19.0 A2.0,2.0 0 0 1 21.0,5.0 V19.0 A2.0,2.0 0 0 1 19.0,21.0 H5.0 A2.0,2.0 0 0 1 3.0,19.0 V5.0 A2.0,2.0 0 0 1 5.0,3.0 Z",
                "M7 3v18",
                "M3 7.5h4",
                "M3 12h18",
                "M3 16.5h4",
                "M17 3v18",
                "M17 7.5h4",
                "M17 16.5h4",
            ),
        )

    val Search: ImageVector
        get() = lucideIcon(
            name = "Lucide.Search",
            pathData = listOf(
                "M11.0,3.0 A8.0,8.0 0 1,0 11.0,19.0 A8.0,8.0 0 1,0 11.0,3.0 Z",
                "m21 21-4.3-4.3",
            ),
        )

    val Settings: ImageVector
        get() = lucideIcon(
            name = "Lucide.Settings",
            pathData = listOf(
                "M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z",
                "M12.0,9.0 A3.0,3.0 0 1,0 12.0,15.0 A3.0,3.0 0 1,0 12.0,9.0 Z",
            ),
        )

    val Tv: ImageVector
        get() = lucideIcon(
            name = "Lucide.Tv",
            pathData = listOf(
                "M4.0,7.0 H20.0 A2.0,2.0 0 0 1 22.0,9.0 V20.0 A2.0,2.0 0 0 1 20.0,22.0 H4.0 A2.0,2.0 0 0 1 2.0,20.0 V9.0 A2.0,2.0 0 0 1 4.0,7.0 Z",
                "M17,2 L12,7 L7,2",
            ),
        )

    val Check: ImageVector
        get() = lucideIcon(
            name = "Lucide.Check",
            pathData = listOf(
                "M20 6 9 17l-5-5",
            ),
        )

    val Favorite: ImageVector
        get() = lucideIcon(
            name = "Lucide.Favorite",
            pathData = listOf(
                "M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z",
            ),
        )

    val PlayArrow: ImageVector
        get() = lucideIcon(
            name = "Lucide.PlayArrow",
            pathData = listOf(
                "M5,3 L19,12 L5,21 L5,3 Z",
            ),
        )

    val PlayCircle: ImageVector
        get() = lucideIcon(
            name = "Lucide.PlayCircle",
            pathData = listOf(
                "M12.0,2.0 A10.0,10.0 0 1,0 12.0,22.0 A10.0,10.0 0 1,0 12.0,2.0 Z",
                "M10,8 L16,12 L10,16 L10,8 Z",
            ),
        )

    val Star: ImageVector
        get() = lucideIcon(
            name = "Lucide.Star",
            pathData = listOf(
                "M12,2 L15.09,8.26 L22,9.27 L17,14.14 L18.18,21.02 L12,17.77 L5.82,21.02 L7,14.14 L2,9.27 L8.91,8.26 L12,2 Z",
            ),
        )

    val Lock: ImageVector
        get() = lucideIcon(
            name = "Lucide.Lock",
            pathData = listOf(
                "M5.0,11.0 H19.0 A2.0,2.0 0 0 1 21.0,13.0 V20.0 A2.0,2.0 0 0 1 19.0,22.0 H5.0 A2.0,2.0 0 0 1 3.0,20.0 V13.0 A2.0,2.0 0 0 1 5.0,11.0 Z",
                "M7 11V7a5 5 0 0 1 10 0v4",
            ),
        )

    val Delete: ImageVector
        get() = lucideIcon(
            name = "Lucide.Delete",
            pathData = listOf(
                "M3 6h18",
                "M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6",
                "M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2",
                "M10,11 L10,17",
                "M14,11 L14,17",
            ),
        )

    val MoreVert: ImageVector
        get() = lucideIcon(
            name = "Lucide.MoreVert",
            pathData = listOf(
                "M12.0,11.0 A1.0,1.0 0 1,0 12.0,13.0 A1.0,1.0 0 1,0 12.0,11.0 Z",
                "M12.0,4.0 A1.0,1.0 0 1,0 12.0,6.0 A1.0,1.0 0 1,0 12.0,4.0 Z",
                "M12.0,18.0 A1.0,1.0 0 1,0 12.0,20.0 A1.0,1.0 0 1,0 12.0,18.0 Z",
            ),
        )

    val Notifications: ImageVector
        get() = lucideIcon(
            name = "Lucide.Notifications",
            pathData = listOf(
                "M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9",
                "M10.3 21a1.94 1.94 0 0 0 3.4 0",
            ),
        )

    val Person: ImageVector
        get() = lucideIcon(
            name = "Lucide.Person",
            pathData = listOf(
                "M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2",
                "M12.0,3.0 A4.0,4.0 0 1,0 12.0,11.0 A4.0,4.0 0 1,0 12.0,3.0 Z",
            ),
        )

    val Schedule: ImageVector
        get() = lucideIcon(
            name = "Lucide.Schedule",
            pathData = listOf(
                "M12.0,2.0 A10.0,10.0 0 1,0 12.0,22.0 A10.0,10.0 0 1,0 12.0,2.0 Z",
                "M12,6 L12,12 L16,14",
            ),
        )

    val Add: ImageVector
        get() = lucideIcon(
            name = "Lucide.Add",
            pathData = listOf(
                "M5 12h14",
                "M12 5v14",
            ),
        )

    val ArrowDropDown: ImageVector
        get() = lucideIcon(
            name = "Lucide.ArrowDropDown",
            pathData = listOf(
                "m6 9 6 6 6-6",
            ),
        )

    val Devices: ImageVector
        get() = lucideIcon(
            name = "Lucide.Devices",
            pathData = listOf(
                "M4.0,3.0 H20.0 A2.0,2.0 0 0 1 22.0,5.0 V15.0 A2.0,2.0 0 0 1 20.0,17.0 H4.0 A2.0,2.0 0 0 1 2.0,15.0 V5.0 A2.0,2.0 0 0 1 4.0,3.0 Z",
                "M8,21 L16,21",
                "M12,17 L12,21",
            ),
        )

    val Speed: ImageVector
        get() = lucideIcon(
            name = "Lucide.Speed",
            pathData = listOf(
                "m12 14 4-4",
                "M3.34 19a10 10 0 1 1 17.32 0",
            ),
        )

    val Upload: ImageVector
        get() = lucideIcon(
            name = "Lucide.Upload",
            pathData = listOf(
                "M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4",
                "M17,8 L12,3 L7,8",
                "M12,3 L12,15",
            ),
        )

    val FlashOn: ImageVector
        get() = lucideIcon(
            name = "Lucide.FlashOn",
            pathData = listOf(
                "M13,2 L3,14 L12,14 L11,22 L21,10 L12,10 L13,2 Z",
            ),
        )

    val ChevronRight: ImageVector
        get() = lucideIcon(
            name = "Lucide.ChevronRight",
            pathData = listOf(
                "m9 18 6-6-6-6",
            ),
        )

    val RepeatOne: ImageVector
        get() = lucideIcon(
            name = "Lucide.RepeatOne",
            pathData = listOf(
                "m17 2 4 4-4 4",
                "M3 11v-1a4 4 0 0 1 4-4h14",
                "m7 22-4-4 4-4",
                "M21 13v1a4 4 0 0 1-4 4H3",
                "M11 10h1v4",
            ),
        )

    val Repeat: ImageVector
        get() = lucideIcon(
            name = "Lucide.Repeat",
            pathData = listOf(
                "m17 2 4 4-4 4",
                "M3 11v-1a4 4 0 0 1 4-4h14",
                "m7 22-4-4 4-4",
                "M21 13v1a4 4 0 0 1-4 4H3",
            ),
        )

    val ShowChart: ImageVector
        get() = lucideIcon(
            name = "Lucide.ShowChart",
            pathData = listOf(
                "M22 12h-4l-3 9L9 3l-3 9H2",
            ),
        )

    val Replay10: ImageVector
        get() = lucideIcon(
            name = "Lucide.Replay10",
            pathData = listOf(
                "M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8",
                "M3 3v5h5",
            ),
        )

    val Forward10: ImageVector
        get() = lucideIcon(
            name = "Lucide.Forward10",
            pathData = listOf(
                "M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8",
                "M21 3v5h-5",
            ),
        )

    val SkipNext: ImageVector
        get() = lucideIcon(
            name = "Lucide.SkipNext",
            pathData = listOf(
                "M5,4 L15,12 L5,20 L5,4 Z",
                "M19,5 L19,19",
            ),
        )

    val MusicTrack: ImageVector
        get() = lucideIcon(
            name = "Lucide.MusicTrack",
            pathData = listOf(
                "M9 18V5l12-2v13",
                "M6.0,15.0 A3.0,3.0 0 1,0 6.0,21.0 A3.0,3.0 0 1,0 6.0,15.0 Z",
                "M18.0,13.0 A3.0,3.0 0 1,0 18.0,19.0 A3.0,3.0 0 1,0 18.0,13.0 Z",
            ),
        )

    val Pause: ImageVector
        get() = lucideIcon(
            name = "Lucide.Pause",
            pathData = listOf(
                "M6.0,4.0 H10.0 V20.0 H6.0 Z",
                "M14.0,4.0 H18.0 V20.0 H14.0 Z",
            ),
        )

    val Info: ImageVector
        get() = lucideIcon(
            name = "Lucide.Info",
            pathData = listOf(
                "M12.0,2.0 A10.0,10.0 0 1,0 12.0,22.0 A10.0,10.0 0 1,0 12.0,2.0 Z",
                "M12 16v-4",
                "M12 8h.01",
            ),
        )

    val Security: ImageVector
        get() = lucideIcon(
            name = "Lucide.Security",
            pathData = listOf(
                "M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z",
            ),
        )

    val Video: ImageVector
        get() = lucideIcon(
            name = "Lucide.Video",
            pathData = listOf(
                "m22 8-6 4 6 4V8Z",
                "M4.0,6.0 H14.0 A2.0,2.0 0 0 1 16.0,8.0 V16.0 A2.0,2.0 0 0 1 14.0,18.0 H4.0 A2.0,2.0 0 0 1 2.0,16.0 V8.0 A2.0,2.0 0 0 1 4.0,6.0 Z",
            ),
        )

    val Mic: ImageVector
        get() = lucideIcon(
            name = "Lucide.Mic",
            pathData = listOf(
                "M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z",
                "M19 10v2a7 7 0 0 1-14 0v-2",
                "M12,19 L12,22",
            ),
        )

    val MicOff: ImageVector
        get() = lucideIcon(
            name = "Lucide.MicOff",
            pathData = listOf(
                "M2,2 L22,22",
                "M18.89 13.23A7.12 7.12 0 0 0 19 12v-2",
                "M5 10v2a7 7 0 0 0 12 5",
                "M15 9.34V5a3 3 0 0 0-5.68-1.33",
                "M9 9v3a3 3 0 0 0 5.12 2.12",
                "M12,19 L12,22",
            ),
        )

    val Cast: ImageVector
        get() = lucideIcon(
            name = "Lucide.Cast",
            pathData = listOf(
                "M2 8V6a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-6",
                "M2 12a9 9 0 0 1 8 8",
                "M2 16a5 5 0 0 1 4 4",
                "M2,20 L2.01,20",
            ),
        )

    val Airplay: ImageVector
        get() = lucideIcon(
            name = "Lucide.Airplay",
            pathData = listOf(
                "M5 17H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2h-1",
                "M12,15 L17,21 L7,21 L12,15 Z",
            ),
        )

    val Film get() = Movie
    val FavoriteBorder get() = Favorite
    val Heart get() = Favorite

    val CheckCircle: ImageVector
        get() = lucideIcon(
            name = "Lucide.CheckCircle",
            pathData = listOf(
                "M22 11.08V12a10 10 0 1 1-5.93-9.14",
                "m9 11 3 3L22 4",
            ),
        )

    val Link: ImageVector
        get() = lucideIcon(
            name = "Lucide.Link",
            pathData = listOf(
                "M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71",
                "M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71",
            ),
        )

    val BrightnessHigh: ImageVector
        get() = lucideIcon(
            name = "Lucide.BrightnessHigh",
            pathData = listOf(
                "M12 2v2",
                "M12 20v2",
                "m4.93 4.93 1.41 1.41",
                "m17.66 17.66 1.41 1.41",
                "M2 12h2",
                "M20 12h2",
                "m6.34 17.66-1.41 1.41",
                "m19.07 4.93-1.41 1.41",
                "M12.0,8.0 A4.0,4.0 0 1,0 12.0,16.0 A4.0,4.0 0 1,0 12.0,8.0 Z",
            ),
        )

    val BrightnessLow: ImageVector
        get() = lucideIcon(
            name = "Lucide.BrightnessLow",
            pathData = listOf(
                "M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z",
            ),
        )

    val ChevronLeft: ImageVector
        get() = lucideIcon(
            name = "Lucide.ChevronLeft",
            pathData = listOf(
                "m15 18-6-6 6-6",
            ),
        )

    val Translate: ImageVector
        get() = lucideIcon(
            name = "Lucide.Translate",
            pathData = listOf(
                "m5 8 6 6",
                "m4 14 6-6 2-3",
                "M2 5h12",
                "M7 2h1",
                "m22 22-5-10-5 10",
                "M14 18h6",
            ),
        )

    val Download: ImageVector
        get() = lucideIcon(
            name = "Lucide.Download",
            pathData = listOf(
                "M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4",
                "M7,10 L12,15 L17,10",
            ),
        )

    val Error: ImageVector
        get() = lucideIcon(
            name = "Lucide.Error",
            pathData = listOf(
                "m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z",
                "M12 9v4",
                "M12 17h.01",
            ),
        )

    val Fullscreen: ImageVector
        get() = lucideIcon(
            name = "Lucide.Fullscreen",
            pathData = listOf(
                "M8 3H5a2 2 0 0 0-2 2v3",
                "M21 8V5a2 2 0 0 0-2-2h-3",
                "M3 16v3a2 2 0 0 0 2 2h3",
                "M16 21h3a2 2 0 0 0 2-2v-3",
            ),
        )

    val FullscreenExit: ImageVector
        get() = lucideIcon(
            name = "Lucide.FullscreenExit",
            pathData = listOf(
                "M8 3v3a2 2 0 0 1-2 2H3",
                "M21 8h-3a2 2 0 0 1-2-2V3",
                "M3 16h3a2 2 0 0 1 2 2v3",
                "M16 21v-3a2 2 0 0 1 2-2h3",
            ),
        )

    val LockOpen: ImageVector
        get() = lucideIcon(
            name = "Lucide.LockOpen",
            pathData = listOf(
                "M7 11V7a5 5 0 0 1 9.9-1",
                "M5 11h14v10H5z",
            ),
        )

    val House: ImageVector
        get() = lucideIcon(
            name = "Lucide.House",
            pathData = listOf(
                "M15 21v-8a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v8",
                "M3 10a2 2 0 0 1 .709-1.528l7-5.999a2 2 0 0 1 2.582 0l7 5.999A2 2 0 0 1 21 10v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z",
            ),
        )

    val Edit: ImageVector
        get() = lucideIcon(
            name = "Lucide.Edit",
            pathData = listOf(
                "M12 20h9",
                "M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z",
            ),
        )

    val Subtitles: ImageVector
        get() = lucideIcon(
            name = "Lucide.Subtitles",
            pathData = listOf(
                "M7 13h4",
                "M15 13h2",
                "M7 9h2",
                "M11 9h6",
                "M5,5 H19 A2,2 0 0 1 21,7 V17 A2,2 0 0 1 19,19 H5 A2,2 0 0 1 3,17 V7 A2,2 0 0 1 5,5 Z",
            ),
        )

}

private fun lucideIcon(
    name: String,
    pathData: List<String>,
) = ImageVector.Builder(
    name = name,
    defaultWidth = 24.dp,
    defaultHeight = 24.dp,
    viewportWidth = 24f,
    viewportHeight = 24f,
).apply {
    pathData.forEach { pathString ->
        addPath(
            pathData = PathParser().parsePathString(pathString.replace("\n", "").trim()).toNodes(),
            pathFillType = PathFillType.NonZero,
            name = "",
            fill = null,
            fillAlpha = 1f,
            stroke = SolidColor(Color.White),
            strokeAlpha = 1f,
            strokeLineWidth = 2f,
            strokeLineCap = StrokeCap.Round,
            strokeLineJoin = StrokeJoin.Round,
            strokeLineMiter = 4f,
            trimPathStart = 0f,
            trimPathEnd = 1f,
            trimPathOffset = 0f,
        )
    }
}.build()
