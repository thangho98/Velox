import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

type Theme = 'light' | 'dark' | 'system'
type ViewMode = 'grid' | 'list' | 'compact'

export type ToastType = 'success' | 'error' | 'info'

export interface Toast {
  id: string
  message: string
  type: ToastType
}

interface UIState {
  // Theme
  theme: Theme

  // Navigation
  sidebarCollapsed: boolean

  // View preferences
  libraryViewMode: ViewMode
  mediaCardSize: 'small' | 'medium' | 'large'

  // Active modals/drawers
  isSearchOpen: boolean
  activeMobileTab: 'home' | 'libraries' | 'favorites' | 'settings'

  // Toast notifications (not persisted)
  toasts: Toast[]

  // Actions
  setTheme: (theme: Theme) => void
  toggleSidebar: () => void
  setLibraryViewMode: (mode: ViewMode) => void
  setMediaCardSize: (size: 'small' | 'medium' | 'large') => void
  toggleSearch: () => void
  setActiveMobileTab: (tab: 'home' | 'libraries' | 'favorites' | 'settings') => void
  addToast: (message: string, type?: ToastType) => void
  removeToast: (id: string) => void
}

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      // Defaults
      theme: 'system',
      sidebarCollapsed: false,
      libraryViewMode: 'grid',
      mediaCardSize: 'medium',
      isSearchOpen: false,
      activeMobileTab: 'home',
      toasts: [],

      // Actions
      setTheme: (theme) => set({ theme }),

      toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),

      setLibraryViewMode: (mode) => set({ libraryViewMode: mode }),

      setMediaCardSize: (size) => set({ mediaCardSize: size }),

      toggleSearch: () => set((state) => ({ isSearchOpen: !state.isSearchOpen })),

      setActiveMobileTab: (tab) => set({ activeMobileTab: tab }),

      addToast: (message, type = 'info') => {
        const id = `toast-${Date.now()}-${Math.random().toString(36).slice(2)}`
        set((state) => ({ toasts: [...state.toasts, { id, message, type }] }))
        // Auto-remove after 4 seconds
        setTimeout(() => {
          set((state) => ({ toasts: state.toasts.filter((t) => t.id !== id) }))
        }, 4000)
      },

      removeToast: (id) => set((state) => ({ toasts: state.toasts.filter((t) => t.id !== id) })),
    }),
    {
      name: 'velox-ui',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        theme: state.theme,
        sidebarCollapsed: state.sidebarCollapsed,
        libraryViewMode: state.libraryViewMode,
        mediaCardSize: state.mediaCardSize,
      }),
    },
  ),
)
