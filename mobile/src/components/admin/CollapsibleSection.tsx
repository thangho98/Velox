/**
 * CollapsibleSection - Reusable collapsible card for admin sections
 */

import { useState } from 'react'
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native'
import { ChevronDown } from 'lucide-react-native'

interface CollapsibleSectionProps {
  title: string
  subtitle?: string
  children: React.ReactNode
  defaultExpanded?: boolean
}

export function CollapsibleSection({
  title,
  subtitle,
  children,
  defaultExpanded = false,
}: CollapsibleSectionProps) {
  const [expanded, setExpanded] = useState(defaultExpanded)

  return (
    <View style={styles.container}>
      <TouchableOpacity
        onPress={() => setExpanded(!expanded)}
        style={styles.header}
        activeOpacity={0.7}
      >
        <View style={styles.headerText}>
          <Text style={styles.title}>{title}</Text>
          {subtitle ? <Text style={styles.subtitle}>{subtitle}</Text> : null}
        </View>
        <ChevronDown
          size={20}
          color="#888"
          style={{ transform: [{ rotate: expanded ? '180deg' : '0deg' }] }}
        />
      </TouchableOpacity>
      {expanded && <View style={styles.content}>{children}</View>}
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    marginBottom: 12,
    overflow: 'hidden',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
  },
  headerText: {
    flex: 1,
    marginRight: 12,
  },
  title: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  subtitle: {
    color: '#888',
    fontSize: 12,
    marginTop: 2,
  },
  content: {
    padding: 16,
    paddingTop: 0,
  },
})
