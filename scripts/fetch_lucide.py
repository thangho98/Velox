import urllib.request
import xml.etree.ElementTree as ET
import os

iconsToFetch = [
    'volume-x', 'volume-2', 'arrow-left', 'log-out', 'x', 'cloud', 'eye', 'eye-off',
    'wifi', 'wifi-off', 'refresh-cw', 'folder', 'folder-open', 'grid-3x3', 'list', 'film', 'search',
    'settings', 'tv', 'check', 'heart', 'play', 'play-circle', 'star', 'edit', 'lock',
    'trash-2', 'more-vertical', 'bell', 'user', 'clock', 'plus', 'chevron-down', 'monitor',
    'gauge', 'upload', 'zap', 'chevron-right', 'repeat-1', 'repeat', 'activity', 'rotate-ccw',
    'rotate-cw', 'skip-forward', 'house', 'subtitles', 'music', 'pause', 'info', 'shield',
    'video', 'mic', 'mic-off', 'cast', 'airplay'
]

nameMapping = {
    'volume-x': 'VolumeOff',
    'volume-2': 'VolumeUp',
    'arrow-left': 'ArrowBack',
    'log-out': 'Logout',
    'x': 'Close',
    'cloud': 'Cloud',
    'eye': 'Visibility',
    'eye-off': 'VisibilityOff',
    'wifi': 'Wifi',
    'wifi-off': 'WifiOff',
    'refresh-cw': 'Refresh',
    'folder': 'Folder',
    'folder-open': 'FolderOpen',
    'grid-3x3': 'GridView',
    'list': 'ListIcon',
    'film': 'Movie',
    'search': 'Search',
    'settings': 'Settings',
    'tv': 'Tv',
    'check': 'Check',
    'heart': 'Favorite',
    'play': 'PlayArrow',
    'play-circle': 'PlayCircle',
    'star': 'Star',
    'edit': 'Edit',
    'lock': 'Lock',
    'trash-2': 'Delete',
    'more-vertical': 'MoreVert',
    'bell': 'Notifications',
    'user': 'Person',
    'clock': 'Schedule',
    'plus': 'Add',
    'chevron-down': 'ArrowDropDown',
    'monitor': 'Devices',
    'gauge': 'Speed',
    'upload': 'Upload',
    'zap': 'FlashOn',
    'chevron-right': 'ChevronRight',
    'repeat-1': 'RepeatOne',
    'repeat': 'Repeat',
    'activity': 'ShowChart',
    'rotate-ccw': 'Replay10',
    'rotate-cw': 'Forward10',
    'skip-forward': 'SkipNext',
    'house': 'House',
    'subtitles': 'Subtitles',
    'music': 'MusicTrack',
    'pause': 'Pause',
    'info': 'Info',
    'shield': 'Security',
}

def to_pascal(s):
    return "".join(word.capitalize() for word in s.split('-'))

def generate_lucide_kt(output_path="android-native/app/src/main/java/com/velox/app/presentation/ui/components/LucideIcons.kt"):
    output = """package com.velox.app.presentation.ui.components\n
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.PathFillType
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.StrokeJoin
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.graphics.vector.PathParser
import androidx.compose.ui.unit.dp

object LucideIcons {
"""

    fetched_icons = set()

    for icon in iconsToFetch:
        try:
            req = urllib.request.urlopen(f"https://unpkg.com/lucide-static@0.344.0/icons/{icon}.svg")
            svg_content = req.read().decode('utf-8')
        except Exception as e:
            print(f"Skipping {icon}: {e}")
            continue
        
        root = ET.fromstring(svg_content)
        paths = []
        
        # Strip namespace for matching tags
        def get_tag(el):
            return el.tag.split('}')[-1]

        for el in root:
            tag = get_tag(el)
            attributes = el.attrib
            if tag == 'path':
                paths.append(attributes['d'])
            elif tag == 'circle':
                cx, cy, r = float(attributes['cx']), float(attributes['cy']), float(attributes['r'])
                paths.append(f"M{cx},{cy-r} A{r},{r} 0 1,0 {cx},{cy+r} A{r},{r} 0 1,0 {cx},{cy-r} Z")
            elif tag == 'rect':
                x = float(attributes.get('x', '0'))
                y = float(attributes.get('y', '0'))
                w = float(attributes.get('width', '0'))
                h = float(attributes.get('height', '0'))
                r = float(attributes.get('rx', '0'))
                if r > 0:
                    paths.append(f"M{x+r},{y} H{x+w-r} A{r},{r} 0 0 1 {x+w},{y+r} V{y+h-r} A{r},{r} 0 0 1 {x+w-r},{y+h} H{x+r} A{r},{r} 0 0 1 {x},{y+h-r} V{y+r} A{r},{r} 0 0 1 {x+r},{y} Z")
                else:
                    paths.append(f"M{x},{y} H{x+w} V{y+h} H{x} Z")
            elif tag == 'line':
                x1, y1, x2, y2 = attributes['x1'], attributes['y1'], attributes['x2'], attributes['y2']
                paths.append(f"M{x1},{y1} L{x2},{y2}")
            elif tag == 'polyline' or tag == 'polygon':
                pts_str = attributes['points'].replace(',', ' ').split()
                pts = [p for p in pts_str if p]
                if len(pts) >= 2:
                    d = f"M{pts[0]},{pts[1]}"
                    for i in range(2, len(pts), 2):
                        if i+1 < len(pts): d += f" L{pts[i]},{pts[i+1]}"
                    if tag == 'polygon': d += " Z"
                    paths.append(d)

        pascalName = nameMapping.get(icon, to_pascal(icon))
        fetched_icons.add(pascalName)
        
        output += f"    val {pascalName}: ImageVector\n"
        output += f"        get() = lucideIcon(\n            name = \"Lucide.{pascalName}\",\n            pathData = listOf(\n"
        for p in paths:
            output += f"                \"{p}\",\n"
        output += "            ),\n        )\n\n"

    # Add logical aliases needed for compatibility with older mappings
    output += "    val Film get() = Movie\n"
    output += "    val FavoriteBorder get() = Favorite\n"
    output += "    val Heart get() = Favorite\n\n"

    # Include manually generated/fallback icons that might not resolve dynamically
    manual_fallbacks = {
        'CheckCircle': ["M22 11.08V12a10 10 0 1 1-5.93-9.14", "m9 11 3 3L22 4"],
        'Link': ["M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71", "M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"],
        'BrightnessHigh': ["M12 2v2", "M12 20v2", "m4.93 4.93 1.41 1.41", "m17.66 17.66 1.41 1.41", "M2 12h2", "M20 12h2", "m6.34 17.66-1.41 1.41", "m19.07 4.93-1.41 1.41", "M12.0,8.0 A4.0,4.0 0 1,0 12.0,16.0 A4.0,4.0 0 1,0 12.0,8.0 Z"],
        'BrightnessLow': ["M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"],
        'ChevronLeft': ["m15 18-6-6 6-6"],
        'Translate': ["m5 8 6 6", "m4 14 6-6 2-3", "M2 5h12", "M7 2h1", "m22 22-5-10-5 10", "M14 18h6"],
        'Download': ["M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4", "M7,10 L12,15 L17,10"],
        'Error': ["m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z", "M12 9v4", "M12 17h.01"],
        'Fullscreen': ["M8 3H5a2 2 0 0 0-2 2v3", "M21 8V5a2 2 0 0 0-2-2h-3", "M3 16v3a2 2 0 0 0 2 2h3", "M16 21h3a2 2 0 0 0 2-2v-3"],
        'FullscreenExit': ["M8 3v3a2 2 0 0 1-2 2H3", "M21 8h-3a2 2 0 0 1-2-2V3", "M3 16h3a2 2 0 0 1 2 2v3", "M16 21v-3a2 2 0 0 1 2-2h3"],
        'LockOpen': ["M7 11V7a5 5 0 0 1 9.9-1", "M5 11h14v10H5z"],
        'House': ["M15 21v-8a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v8", "M3 10a2 2 0 0 1 .709-1.528l7-5.999a2 2 0 0 1 2.582 0l7 5.999A2 2 0 0 1 21 10v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"],
        'Edit': ["M12 20h9", "M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"],
        'Subtitles': ["M7 13h4", "M15 13h2", "M7 9h2", "M11 9h6", "M5,5 H19 A2,2 0 0 1 21,7 V17 A2,2 0 0 1 19,19 H5 A2,2 0 0 1 3,17 V7 A2,2 0 0 1 5,5 Z"],
        'Pause': ["M14 4h4v16h-4z", "M6 4h4v16H6z"]
    }
    
    for k, paths in manual_fallbacks.items():
        if k not in fetched_icons:
            output += f"    val {k}: ImageVector\n"
            output += f"        get() = lucideIcon(\n            name = \"Lucide.{k}\",\n            pathData = listOf(\n"
            for p in paths:
                output += f"                \"{p}\",\n"
            output += "            ),\n        )\n\n"

    output += "}\n\n"
    output += """private fun lucideIcon(
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
            pathData = PathParser().parsePathString(pathString.replace("\\n", "").trim()).toNodes(),
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
"""
    
    os.makedirs(os.path.dirname(output_path), exist_ok=True)
    with open(output_path, "w") as f:
        f.write(output)

if __name__ == "__main__":
    generate_lucide_kt()
    print("Successfully generated LucideIcons.kt with robust XML-based parsing.")
