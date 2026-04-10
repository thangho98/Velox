import os
import re

mapping = {
    'Icons.AutoMirrored.Filled.VolumeOff': 'LucideIcons.VolumeOff',
    'Icons.AutoMirrored.Filled.VolumeUp': 'LucideIcons.VolumeUp',
    'Icons.AutoMirrored.Filled.ArrowBack': 'LucideIcons.ArrowBack',
    'Icons.AutoMirrored.Filled.Logout': 'LucideIcons.Logout',
    'Icons.Default.Clear': 'LucideIcons.Close',
    'Icons.Default.Close': 'LucideIcons.Close',
    'Icons.Default.Cloud': 'LucideIcons.Cloud',
    'Icons.Default.Visibility': 'LucideIcons.Visibility',
    'Icons.Default.VisibilityOff': 'LucideIcons.VisibilityOff',
    'Icons.Default.Wifi': 'LucideIcons.Wifi',
    'Icons.Default.WifiOff': 'LucideIcons.WifiOff',
    'Icons.Default.Refresh': 'LucideIcons.Refresh',
    'Icons.Filled.Refresh': 'LucideIcons.Refresh',
    'Icons.Default.ArrowBack': 'LucideIcons.ArrowBack',
    'Icons.Default.Folder': 'LucideIcons.Folder',
    'Icons.Default.LibraryBooks': 'LucideIcons.Folder',
    'Icons.Default.GridView': 'LucideIcons.GridView',
    'Icons.Default.List': 'LucideIcons.ListIcon',
    'Icons.Default.Movie': 'LucideIcons.Movie',
    'Icons.Default.Search': 'LucideIcons.Search',
    'Icons.Default.Settings': 'LucideIcons.Settings',
    'Icons.Default.Tv': 'LucideIcons.Tv',
    'Icons.Default.Check': 'LucideIcons.Check',
    'Icons.Default.Favorite': 'LucideIcons.Favorite',
    'Icons.Default.FavoriteBorder': 'LucideIcons.Favorite',
    'Icons.Default.PlayArrow': 'LucideIcons.PlayArrow',
    'Icons.Default.PlayCircle': 'LucideIcons.PlayCircle',
    'Icons.Default.Star': 'LucideIcons.Star',
    'Icons.Default.Edit': 'LucideIcons.Edit',
    'Icons.Default.Lock': 'LucideIcons.Lock',
    'Icons.Default.Delete': 'LucideIcons.Delete',
    'Icons.Default.MoreVert': 'LucideIcons.MoreVert',
    'Icons.Default.Notifications': 'LucideIcons.Notifications',
    'Icons.Default.Person': 'LucideIcons.Person',
    'Icons.Default.Schedule': 'LucideIcons.Schedule',
    'Icons.Default.Add': 'LucideIcons.Add',
    'Icons.Default.ArrowDropDown': 'LucideIcons.ArrowDropDown',
    'Icons.Default.Devices': 'LucideIcons.Devices',
    'Icons.Default.Speed': 'LucideIcons.Speed',
    'Icons.Default.Upload': 'LucideIcons.Upload',
    'Icons.Default.FlashOn': 'LucideIcons.FlashOn',
    'Icons.Default.ChevronRight': 'LucideIcons.ChevronRight',
    'Icons.Default.RepeatOne': 'LucideIcons.RepeatOne',
    'Icons.Default.Repeat': 'LucideIcons.Repeat',
    'Icons.Outlined.ShowChart': 'LucideIcons.ShowChart',
    'Icons.Outlined.ClosedCaption': 'LucideIcons.Subtitles',
    'Icons.Default.Replay10': 'LucideIcons.Replay10',
    'Icons.Default.Forward10': 'LucideIcons.Forward10',
    'Icons.Outlined.SkipNext': 'LucideIcons.SkipNext',
    
    # Custom restored ones
    'LucideIcons.MusicTrack': 'LucideIcons.MusicTrack',
    'LucideIcons.Subtitles': 'LucideIcons.Subtitles',
    'LucideIcons.House': 'LucideIcons.House'
}

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    original = content
    for k, v in mapping.items():
        content = content.replace(k, v)
        
    # Also add import if LucideIcons is used
    if 'LucideIcons' in content and 'import com.velox.app.presentation.ui.components.LucideIcons' not in content:
        # insert import around other imports
        content = re.sub(r'(import [^\n]+\n)', r'\1import com.velox.app.presentation.ui.components.LucideIcons\n', content, count=1)
        
    if original != content:
        with open(filepath, 'w') as f:
            f.write(content)
        print(f"Updated {filepath}")

for root, dirs, files in os.walk('android-native/app/src/main/java'):
    for file in files:
        if file.endswith('.kt') and file != 'LucideIcons.kt':
            process_file(os.path.join(root, file))
