import re
import os

with open('android-native/app/src/main/java/com/velox/app/presentation/ui/components/LucideIcons.kt', 'r') as f:
    text = f.read()

# Fix Subtitles definition
sub = """    val Subtitles: ImageVector
        get() = lucideIcon(
            name = "Lucide.Subtitles",
            pathData = listOf(
                "M7 13h4",
                "M15 13h2",
                "M7 9h2",
                "M11 9h6",
                "M5,5 H19 A2,2 0 0 1 21,7 V17 A2,2 0 0 1 19,19 H5 A2,2 0 0 1 3,17 V7 A2,2 0 0 1 5,5 Z"
            ),
        )"""

text = re.sub(r'    val Subtitles: ImageVector.*?\}\.build\(\)', sub, text, flags=re.DOTALL)

if 'val Film: ImageVector' not in text:
    text = text.replace('val Movie: ImageVector', 'val Movie: ImageVector\n    val Film get() = Movie')
if 'val FolderOpen: ImageVector' not in text:
    text = text.replace('val Folder: ImageVector', 'val Folder: ImageVector\n    val FolderOpen get() = Folder')
if 'val Heart: ImageVector' not in text:
    text = text.replace('val Favorite: ImageVector', 'val Favorite: ImageVector\n    val Heart get() = Favorite')
if 'val FavoriteBorder: ImageVector' not in text:
    text = text.replace('val Favorite: ImageVector', 'val Favorite: ImageVector\n    val FavoriteBorder get() = Favorite')

with open('android-native/app/src/main/java/com/velox/app/presentation/ui/components/LucideIcons.kt', 'w') as f:
    f.write(text)

mapping = {
    'Icons.Default.Link': 'LucideIcons.Link',
    'Icons.Default.ChevronLeft': 'LucideIcons.ChevronLeft',
    'Icons.Default.ChevronRight': 'LucideIcons.ChevronRight',
    'Icons.Default.Translate': 'LucideIcons.Translate',
    'Icons.Default.Download': 'LucideIcons.Download',
    'Icons.Filled.Error': 'LucideIcons.Error',
    'Icons.Default.Error': 'LucideIcons.Error',
    'Icons.Filled.LockOpen': 'LucideIcons.LockOpen',
    'Icons.Outlined.LockOpen': 'LucideIcons.LockOpen',
    'Icons.Filled.Fullscreen': 'LucideIcons.Fullscreen',
    'Icons.Outlined.Fullscreen': 'LucideIcons.Fullscreen',
    'Icons.Default.Fullscreen': 'LucideIcons.Fullscreen',
    'Icons.Default.FullscreenExit': 'LucideIcons.FullscreenExit',
    'Icons.Outlined.FullscreenExit': 'LucideIcons.FullscreenExit',
    'Icons.Filled.FullscreenExit': 'LucideIcons.FullscreenExit',
    'Icons.Filled.FavoriteBorder': 'LucideIcons.FavoriteBorder',
    'Icons.Filled.PlayArrow': 'LucideIcons.PlayArrow',
    'Icons.Filled.Info': 'LucideIcons.Info',
    'Icons.Filled.Link': 'LucideIcons.Link',
    'Icons.Filled.Add': 'LucideIcons.Add',
    'Icons.Filled.Refresh': 'LucideIcons.Refresh',
    'Icons.Filled.Check': 'LucideIcons.Check',
    'Icons.Filled.Close': 'LucideIcons.Close',
    'Icons.Filled.Delete': 'LucideIcons.Delete',
    'Icons.Default.BrightnessHigh': 'LucideIcons.BrightnessHigh',
    'Icons.Default.BrightnessLow': 'LucideIcons.BrightnessLow',
    'Icons.Filled.BrightnessHigh': 'LucideIcons.BrightnessHigh',
    'Icons.Filled.BrightnessLow': 'LucideIcons.BrightnessLow',
}

for root, _, files in os.walk('android-native/app/src/main/java'):
    for file in files:
        if file.endswith('.kt') and file != 'LucideIcons.kt':
            filepath = os.path.join(root, file)
            with open(filepath, 'r') as f:
                content = f.read()
            original = content
            for k, v in mapping.items():
                content = content.replace(k, v)
            if content != original:
                with open(filepath, 'w') as f:
                    f.write(content)
