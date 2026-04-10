import re
import os

mapping = {
    'Icons.Default.SkipNext': 'LucideIcons.SkipNext',
    'Icons.Default.BrightnessHigh': 'LucideIcons.BrightnessHigh',
    'Icons.Default.BrightnessLow': 'LucideIcons.BrightnessLow',
    'Icons.Default.Info': 'LucideIcons.Info',
    'Icons.Default.ChevronLeft': 'LucideIcons.ChevronLeft',
    'Icons.Default.Translate': 'LucideIcons.Translate',
    'Icons.Default.Download': 'LucideIcons.Download'
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
            if 'LucideIcons' in content and 'import com.velox.app.presentation.ui.components.LucideIcons' not in content:
                content = re.sub(r'(import [^\n]+\n)', r'\1import com.velox.app.presentation.ui.components.LucideIcons\n', content, count=1)
            
            if content != original:
                with open(filepath, 'w') as f:
                    f.write(content)
