import { useState } from 'react'
import {
  View,
  Text,
  Modal,
  TouchableOpacity,
  ScrollView,
  StyleSheet,
  Pressable,
} from 'react-native'
import { Check } from 'lucide-react-native'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'

interface FilterBottomSheetProps {
  visible: boolean
  title: string
  options: string[]
  selectedValue?: string
  onSelect: (value: string | undefined) => void
  onClose: () => void
  allowClear?: boolean
}

export function FilterBottomSheet({
  visible,
  title,
  options,
  selectedValue,
  onSelect,
  onClose,
  allowClear = true,
}: FilterBottomSheetProps) {
  const layout = useResponsiveLayout()
  const isWideScreen = layout.device !== 'phone'

  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={onClose}
    >
      <Pressable style={[styles.overlay, isWideScreen && { justifyContent: 'center', alignItems: 'center' }]} onPress={onClose}>
        <Pressable
          style={[
            styles.container,
            isWideScreen && {
              maxWidth: layout.largeControls ? 600 : 500,
              width: '90%',
              borderRadius: 16,
              borderTopLeftRadius: 16,
              borderTopRightRadius: 16,
            },
          ]}
          onPress={(e) => e.stopPropagation()}
        >
          <Text style={[styles.title, { fontSize: scaledFont(18, layout.fontScale) }]}>{title}</Text>
          <ScrollView style={styles.list} showsVerticalScrollIndicator={false}>
            {allowClear && selectedValue && (
              <TouchableOpacity
                style={[styles.item, !selectedValue && styles.itemSelected, { paddingVertical: layout.largeControls ? 18 : 14 }]}
                onPress={() => {
                  onSelect(undefined)
                  onClose()
                }}
              >
                <Text style={[styles.itemText, !selectedValue && styles.itemTextSelected, { fontSize: scaledFont(16, layout.fontScale) }]}>
                  All
                </Text>
              </TouchableOpacity>
            )}
            {options.map((option) => (
              <TouchableOpacity
                key={option}
                style={[styles.item, selectedValue === option && styles.itemSelected, { paddingVertical: layout.largeControls ? 18 : 14 }]}
                onPress={() => {
                  onSelect(option)
                  onClose()
                }}
              >
                <Text
                  style={[styles.itemText, selectedValue === option && styles.itemTextSelected, { fontSize: scaledFont(16, layout.fontScale) }]}
                  numberOfLines={1}
                >
                  {option}
                </Text>
                {selectedValue === option && (
                  <Check size={scaledFont(16, layout.fontScale)} color="#e50914" />
                )}
              </TouchableOpacity>
            ))}
            {options.length === 0 && (
              <Text style={[styles.emptyText, { fontSize: scaledFont(14, layout.fontScale) }]}>No options available</Text>
            )}
          </ScrollView>
          <TouchableOpacity style={[styles.closeButton, { paddingVertical: layout.largeControls ? 18 : 14 }]} onPress={onClose}>
            <Text style={[styles.closeButtonText, { fontSize: scaledFont(16, layout.fontScale) }]}>Close</Text>
          </TouchableOpacity>
        </Pressable>
      </Pressable>
    </Modal>
  )
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    justifyContent: 'flex-end',
  },
  container: {
    backgroundColor: '#1f1f1f',
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    paddingTop: 20,
    paddingBottom: 34,
    maxHeight: '70%',
  },
  title: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#fff',
    textAlign: 'center',
    marginBottom: 16,
    paddingHorizontal: 20,
  },
  list: {
    maxHeight: 400,
  },
  item: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 14,
    paddingHorizontal: 20,
    borderBottomWidth: 1,
    borderBottomColor: '#2a2a2a',
  },
  itemSelected: {
    backgroundColor: 'rgba(229, 9, 20, 0.15)',
  },
  itemText: {
    fontSize: 16,
    color: '#ccc',
    flex: 1,
  },
  itemTextSelected: {
    color: '#e50914',
    fontWeight: '600',
  },
  emptyText: {
    fontSize: 14,
    color: '#888',
    textAlign: 'center',
    paddingVertical: 30,
  },
  closeButton: {
    marginTop: 16,
    marginHorizontal: 20,
    paddingVertical: 14,
    backgroundColor: '#2a2a2a',
    borderRadius: 8,
    alignItems: 'center',
  },
  closeButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#fff',
  },
})
