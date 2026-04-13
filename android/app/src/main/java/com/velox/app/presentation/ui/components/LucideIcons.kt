package com.velox.app.presentation.ui.components

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.rounded.ArrowBack
import androidx.compose.material.icons.automirrored.rounded.KeyboardArrowLeft
import androidx.compose.material.icons.automirrored.rounded.KeyboardArrowRight
import androidx.compose.material.icons.automirrored.rounded.List
import androidx.compose.material.icons.automirrored.rounded.Logout
import androidx.compose.material.icons.automirrored.rounded.ShowChart
import androidx.compose.material.icons.automirrored.rounded.VolumeOff
import androidx.compose.material.icons.automirrored.rounded.VolumeUp
import androidx.compose.material.icons.rounded.Add
import androidx.compose.material.icons.rounded.Airplay
import androidx.compose.material.icons.rounded.ArrowDropDown
import androidx.compose.material.icons.rounded.BrightnessHigh
import androidx.compose.material.icons.rounded.Cast
import androidx.compose.material.icons.rounded.Check
import androidx.compose.material.icons.rounded.CheckCircle
import androidx.compose.material.icons.rounded.Close
import androidx.compose.material.icons.rounded.Cloud
import androidx.compose.material.icons.rounded.DarkMode
import androidx.compose.material.icons.rounded.Delete
import androidx.compose.material.icons.rounded.Devices
import androidx.compose.material.icons.rounded.Download
import androidx.compose.material.icons.rounded.Edit
import androidx.compose.material.icons.rounded.Favorite
import androidx.compose.material.icons.rounded.FavoriteBorder
import androidx.compose.material.icons.rounded.FlashOn
import androidx.compose.material.icons.rounded.Folder
import androidx.compose.material.icons.rounded.FolderOpen
import androidx.compose.material.icons.rounded.Forward5
import androidx.compose.material.icons.rounded.Fullscreen
import androidx.compose.material.icons.rounded.FullscreenExit
import androidx.compose.material.icons.rounded.GridView
import androidx.compose.material.icons.rounded.Home
import androidx.compose.material.icons.rounded.Info
import androidx.compose.material.icons.rounded.Link
import androidx.compose.material.icons.rounded.Lock
import androidx.compose.material.icons.rounded.LockOpen
import androidx.compose.material.icons.rounded.Mic
import androidx.compose.material.icons.rounded.MicOff
import androidx.compose.material.icons.rounded.MoreVert
import androidx.compose.material.icons.rounded.Movie
import androidx.compose.material.icons.rounded.MusicNote
import androidx.compose.material.icons.rounded.Notifications
import androidx.compose.material.icons.rounded.Pause
import androidx.compose.material.icons.rounded.Person
import androidx.compose.material.icons.rounded.PlayArrow
import androidx.compose.material.icons.rounded.PlayCircle
import androidx.compose.material.icons.rounded.Refresh
import androidx.compose.material.icons.rounded.Repeat
import androidx.compose.material.icons.rounded.RepeatOne
import androidx.compose.material.icons.rounded.Replay5
import androidx.compose.material.icons.rounded.Schedule
import androidx.compose.material.icons.rounded.Search
import androidx.compose.material.icons.rounded.Security
import androidx.compose.material.icons.rounded.Settings
import androidx.compose.material.icons.rounded.SkipNext
import androidx.compose.material.icons.rounded.Speed
import androidx.compose.material.icons.rounded.Star
import androidx.compose.material.icons.rounded.Subtitles
import androidx.compose.material.icons.rounded.Translate
import androidx.compose.material.icons.rounded.Tv
import androidx.compose.material.icons.rounded.Upload
import androidx.compose.material.icons.rounded.Videocam
import androidx.compose.material.icons.rounded.Visibility
import androidx.compose.material.icons.rounded.VisibilityOff
import androidx.compose.material.icons.rounded.Warning
import androidx.compose.material.icons.rounded.Wifi
import androidx.compose.material.icons.rounded.WifiOff

/**
 * Icon aliases that map to Material Rounded icons.
 * Keeps the LucideIcons.X API so callers don't need changes.
 */
object LucideIcons {
    val VolumeOff get() = Icons.AutoMirrored.Rounded.VolumeOff
    val VolumeUp get() = Icons.AutoMirrored.Rounded.VolumeUp
    val ArrowBack get() = Icons.AutoMirrored.Rounded.ArrowBack
    val Logout get() = Icons.AutoMirrored.Rounded.Logout
    val Close get() = Icons.Rounded.Close
    val Cloud get() = Icons.Rounded.Cloud
    val Visibility get() = Icons.Rounded.Visibility
    val VisibilityOff get() = Icons.Rounded.VisibilityOff
    val Wifi get() = Icons.Rounded.Wifi
    val WifiOff get() = Icons.Rounded.WifiOff
    val Refresh get() = Icons.Rounded.Refresh
    val Folder get() = Icons.Rounded.Folder
    val FolderOpen get() = Icons.Rounded.FolderOpen
    val GridView get() = Icons.Rounded.GridView
    val ListIcon get() = Icons.AutoMirrored.Rounded.List
    val Movie get() = Icons.Rounded.Movie
    val Search get() = Icons.Rounded.Search
    val Settings get() = Icons.Rounded.Settings
    val Tv get() = Icons.Rounded.Tv
    val Check get() = Icons.Rounded.Check
    val Favorite get() = Icons.Rounded.Favorite
    val PlayArrow get() = Icons.Rounded.PlayArrow
    val PlayCircle get() = Icons.Rounded.PlayCircle
    val Star get() = Icons.Rounded.Star
    val Lock get() = Icons.Rounded.Lock
    val Delete get() = Icons.Rounded.Delete
    val MoreVert get() = Icons.Rounded.MoreVert
    val Notifications get() = Icons.Rounded.Notifications
    val Person get() = Icons.Rounded.Person
    val Schedule get() = Icons.Rounded.Schedule
    val Add get() = Icons.Rounded.Add
    val ArrowDropDown get() = Icons.Rounded.ArrowDropDown
    val Devices get() = Icons.Rounded.Devices
    val Speed get() = Icons.Rounded.Speed
    val Upload get() = Icons.Rounded.Upload
    val FlashOn get() = Icons.Rounded.FlashOn
    val ChevronRight get() = Icons.AutoMirrored.Rounded.KeyboardArrowRight
    val RepeatOne get() = Icons.Rounded.RepeatOne
    val Repeat get() = Icons.Rounded.Repeat
    val ShowChart get() = Icons.AutoMirrored.Rounded.ShowChart
    val Replay5 get() = Icons.Rounded.Replay5
    val Forward5 get() = Icons.Rounded.Forward5
    val SkipNext get() = Icons.Rounded.SkipNext
    val MusicTrack get() = Icons.Rounded.MusicNote
    val Pause get() = Icons.Rounded.Pause
    val Info get() = Icons.Rounded.Info
    val Security get() = Icons.Rounded.Security
    val Video get() = Icons.Rounded.Videocam
    val Mic get() = Icons.Rounded.Mic
    val MicOff get() = Icons.Rounded.MicOff
    val Cast get() = Icons.Rounded.Cast
    val Airplay get() = Icons.Rounded.Airplay
    val Film get() = Icons.Rounded.Movie
    val FavoriteBorder get() = Icons.Rounded.FavoriteBorder
    val Heart get() = Icons.Rounded.Favorite
    val CheckCircle get() = Icons.Rounded.CheckCircle
    val Link get() = Icons.Rounded.Link
    val BrightnessHigh get() = Icons.Rounded.BrightnessHigh
    val BrightnessLow get() = Icons.Rounded.DarkMode
    val ChevronLeft get() = Icons.AutoMirrored.Rounded.KeyboardArrowLeft
    val Translate get() = Icons.Rounded.Translate
    val Download get() = Icons.Rounded.Download
    val Error get() = Icons.Rounded.Warning
    val Fullscreen get() = Icons.Rounded.Fullscreen
    val FullscreenExit get() = Icons.Rounded.FullscreenExit
    val LockOpen get() = Icons.Rounded.LockOpen
    val House get() = Icons.Rounded.Home
    val Edit get() = Icons.Rounded.Edit
    val Subtitles get() = Icons.Rounded.Subtitles
    val Replay get() = Icons.Rounded.Refresh
}
