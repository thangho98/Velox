import re

with open('android-native/app/src/main/java/com/velox/app/presentation/ui/components/VideoPlayer.kt', 'r') as f:
    content = f.read()

mapping = {
    'Icons.Default.Error': 'LucideIcons.Error',
    'Icons.Default.VolumeOff': 'LucideIcons.VolumeOff',
    'Icons.Default.VolumeUp': 'LucideIcons.VolumeUp',
    'Icons.Outlined.Settings': 'LucideIcons.Settings',
    'Icons.Outlined.LockOpen': 'LucideIcons.LockOpen',
    'Icons.Outlined.FullscreenExit': 'LucideIcons.FullscreenExit',
    'Icons.Outlined.Fullscreen': 'LucideIcons.Fullscreen',
    'Icons.Default.Pause': 'LucideIcons.Pause',
    'Icons.Default.Fullscreen': 'LucideIcons.Fullscreen',
    'Icons.Outlined.Subtitles': 'LucideIcons.Subtitles'
}

for k, v in mapping.items():
    content = content.replace(k, v)

with open('android-native/app/src/main/java/com/velox/app/presentation/ui/components/VideoPlayer.kt', 'w') as f:
    f.write(content)
