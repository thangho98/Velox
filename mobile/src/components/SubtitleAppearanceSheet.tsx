/**
 * SubtitleAppearanceSheet — modal for customizing subtitle appearance.
 *
 * Settings: Size (Small/Medium/Large), Color (white/yellow/green/cyan),
 * Background (None/Semi-transparent/Solid).
 *
 * Reads and writes to the shared player store so settings persist.
 */

import {
  View,
  Text,
  StyleSheet,
  Modal,
  TouchableOpacity,
  ScrollView,
} from 'react-native'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'

// ---------- Types ----------

interface SubtitleAppearanceSheetProps {
  visible: boolean
  onClose: () => void
  // Current values
  size: 'small' | 'medium' | 'large'
  color: string
  background: 'solid' | 'semi' | 'none'
  // Setters
  onSizeChange: (size: 'small' | 'medium' | 'large') => void
  onColorChange: (color: string) => void
  onBackgroundChange: (bg: 'solid' | 'semi' | 'none') => void
}

// ---------- Options ----------

const SIZE_OPTIONS: { label: string; value: 'small' | 'medium' | 'large' }[] = [
  { label: 'Small', value: 'small' },
  { label: 'Medium', value: 'medium' },
  { label: 'Large', value: 'large' },
]

const COLOR_OPTIONS: { label: string; value: string; preview: string }[] = [
  { label: 'White', value: '#ffffff', preview: '#ffffff' },
  { label: 'Yellow', value: '#FFD700', preview: '#FFD700' },
  { label: 'Green', value: '#00FF00', preview: '#00FF00' },
  { label: 'Cyan', value: '#00FFFF', preview: '#00FFFF' },
]

const BACKGROUND_OPTIONS: { label: string; value: 'none' | 'semi' | 'solid' }[] = [
  { label: 'None', value: 'none' },
  { label: 'Semi-transparent', value: 'semi' },
  { label: 'Solid', value: 'solid' },
]

// ---------- Component ----------

export function SubtitleAppearanceSheet({
  visible,
  onClose,
  size,
  color,
  background,
  onSizeChange,
  onColorChange,
  onBackgroundChange,
}: SubtitleAppearanceSheetProps) {
  const layout = useResponsiveLayout()
  const isWideScreen = layout.device !== 'phone'

  return (
    <Modal
      visible={visible}
      transparent
      animationType={isWideScreen ? 'fade' : 'slide'}
      onRequestClose={onClose}
    >
      <TouchableOpacity
        style={[styles.overlay, isWideScreen && { justifyContent: 'center', alignItems: 'center' }]}
        activeOpacity={1}
        onPress={onClose}
      >
        <View
          style={[
            styles.sheet,
            { maxHeight: layout.height * 0.65 },
            isWideScreen && {
              maxWidth: layout.largeControls ? 600 : 500,
              width: '90%',
              borderRadius: 16,
              borderTopLeftRadius: 16,
              borderTopRightRadius: 16,
            },
          ]}
          onStartShouldSetResponder={() => true}
        >
          {/* Handle bar */}
          {!isWideScreen && <View style={styles.handleBar} />}

          <ScrollView showsVerticalScrollIndicator={false}>
            <Text style={[styles.title, { fontSize: scaledFont(20, layout.fontScale) }]}>Subtitle Appearance</Text>

            {/* Preview */}
            <View style={styles.previewContainer}>
              <View
                style={[
                  styles.previewBox,
                  background === 'solid'
                    ? { backgroundColor: 'rgba(0,0,0,0.8)' }
                    : background === 'semi'
                      ? { backgroundColor: 'rgba(0,0,0,0.5)' }
                      : undefined,
                ]}
              >
                <Text
                  style={[
                    styles.previewText,
                    {
                      color,
                      fontSize: size === 'small' ? 16 : size === 'medium' ? 20 : 26,
                      textShadowColor: 'rgba(0,0,0,0.9)',
                      textShadowOffset: { width: 1, height: 1 },
                      textShadowRadius: background === 'none' ? 4 : 2,
                    },
                  ]}
                >
                  Sample Subtitle Text
                </Text>
              </View>
            </View>

            {/* Size */}
            <Text style={[styles.sectionTitle, { fontSize: scaledFont(13, layout.fontScale) }]}>Size</Text>
            <View style={styles.optionRow}>
              {SIZE_OPTIONS.map((opt) => (
                <TouchableOpacity
                  key={opt.value}
                  style={[
                    styles.optionButton,
                    size === opt.value && styles.optionButtonSelected,
                    { paddingVertical: layout.largeControls ? 16 : 12 },
                  ]}
                  onPress={() => onSizeChange(opt.value)}
                >
                  <Text
                    style={[
                      styles.optionText,
                      size === opt.value && styles.optionTextSelected,
                      { fontSize: scaledFont(14, layout.fontScale) },
                    ]}
                  >
                    {opt.label}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>

            {/* Color */}
            <Text style={[styles.sectionTitle, { fontSize: scaledFont(13, layout.fontScale) }]}>Color</Text>
            <View style={styles.optionRow}>
              {COLOR_OPTIONS.map((opt) => (
                <TouchableOpacity
                  key={opt.value}
                  style={[
                    styles.colorButton,
                    color === opt.value && styles.colorButtonSelected,
                    { paddingVertical: layout.largeControls ? 14 : 10 },
                  ]}
                  onPress={() => onColorChange(opt.value)}
                >
                  <View
                    style={[styles.colorSwatch, { backgroundColor: opt.preview, width: scaledFont(24, layout.fontScale), height: scaledFont(24, layout.fontScale), borderRadius: scaledFont(12, layout.fontScale) }]}
                  />
                  <Text
                    style={[
                      styles.colorLabel,
                      color === opt.value && styles.optionTextSelected,
                      { fontSize: scaledFont(12, layout.fontScale) },
                    ]}
                  >
                    {opt.label}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>

            {/* Background */}
            <Text style={[styles.sectionTitle, { fontSize: scaledFont(13, layout.fontScale) }]}>Background</Text>
            <View style={styles.optionRow}>
              {BACKGROUND_OPTIONS.map((opt) => (
                <TouchableOpacity
                  key={opt.value}
                  style={[
                    styles.optionButton,
                    background === opt.value && styles.optionButtonSelected,
                    { paddingVertical: layout.largeControls ? 16 : 12 },
                  ]}
                  onPress={() => onBackgroundChange(opt.value)}
                >
                  <Text
                    style={[
                      styles.optionText,
                      background === opt.value && styles.optionTextSelected,
                      { fontSize: scaledFont(14, layout.fontScale) },
                    ]}
                  >
                    {opt.label}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>

            {/* Close button */}
            <TouchableOpacity style={[styles.closeButton, { paddingVertical: layout.largeControls ? 18 : 14 }]} onPress={onClose}>
              <Text style={[styles.closeButtonText, { fontSize: scaledFont(16, layout.fontScale) }]}>Done</Text>
            </TouchableOpacity>
          </ScrollView>
        </View>
      </TouchableOpacity>
    </Modal>
  )
}

// ---------- Styles ----------

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'flex-end',
  },
  sheet: {
    backgroundColor: '#1a1a1a',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    paddingHorizontal: 20,
    paddingBottom: 40,
  },
  handleBar: {
    width: 36,
    height: 4,
    backgroundColor: '#555',
    borderRadius: 2,
    alignSelf: 'center',
    marginTop: 12,
    marginBottom: 8,
  },
  title: {
    color: '#fff',
    fontSize: 20,
    fontWeight: '700',
    textAlign: 'center',
    marginBottom: 16,
  },
  previewContainer: {
    backgroundColor: '#000',
    borderRadius: 8,
    padding: 16,
    marginBottom: 20,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 60,
  },
  previewBox: {
    borderRadius: 4,
    paddingHorizontal: 12,
    paddingVertical: 6,
  },
  previewText: {
    fontWeight: '700',
    textAlign: 'center',
  },
  sectionTitle: {
    color: '#aaa',
    fontSize: 13,
    fontWeight: '600',
    textTransform: 'uppercase',
    letterSpacing: 1,
    marginBottom: 10,
  },
  optionRow: {
    flexDirection: 'row',
    gap: 10,
    marginBottom: 20,
    flexWrap: 'wrap',
  },
  optionButton: {
    flex: 1,
    minWidth: 80,
    paddingVertical: 12,
    paddingHorizontal: 8,
    borderRadius: 8,
    backgroundColor: '#2a2a2a',
    alignItems: 'center',
    borderWidth: 2,
    borderColor: 'transparent',
  },
  optionButtonSelected: {
    borderColor: '#e50914',
    backgroundColor: 'rgba(229, 9, 20, 0.15)',
  },
  optionText: {
    color: '#ccc',
    fontSize: 14,
    fontWeight: '500',
  },
  optionTextSelected: {
    color: '#fff',
    fontWeight: '600',
  },
  colorButton: {
    flex: 1,
    minWidth: 70,
    paddingVertical: 10,
    paddingHorizontal: 6,
    borderRadius: 8,
    backgroundColor: '#2a2a2a',
    alignItems: 'center',
    borderWidth: 2,
    borderColor: 'transparent',
    gap: 6,
  },
  colorButtonSelected: {
    borderColor: '#e50914',
    backgroundColor: 'rgba(229, 9, 20, 0.15)',
  },
  colorSwatch: {
    width: 24,
    height: 24,
    borderRadius: 12,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.3)',
  },
  colorLabel: {
    color: '#ccc',
    fontSize: 12,
    fontWeight: '500',
  },
  closeButton: {
    marginTop: 8,
    backgroundColor: '#e50914',
    borderRadius: 10,
    paddingVertical: 14,
    alignItems: 'center',
  },
  closeButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '700',
  },
})
