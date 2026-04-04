import { useEffect, useState } from 'react'
import { StatusBar } from 'expo-status-bar'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { NavigationContainer } from '@react-navigation/native'
import { createNativeStackNavigator } from '@react-navigation/native-stack'
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs'
import {
  View,
  ActivityIndicator,
  Text,
  StyleSheet,
  TouchableOpacity,
  ScrollView,
} from 'react-native'
import { House, Film, Tv, FolderOpen, Heart, Search, User, Settings } from 'lucide-react-native'
// react-native-google-cast hooks work without a provider wrapper
import { isTV, TV_SIDEBAR_WIDTH } from './src/lib/tv'

// Initialize platform adapter FIRST
import { initPlatform } from '@velox/shared/platform'
import { mobilePlatform, loadServerUrl } from './src/platform/mobile-adapter'
initPlatform(mobilePlatform)

// Import stores after platform is initialized
import { useAuthStore } from './src/stores'

// Screens
import { LoginScreen } from './src/screens/LoginScreen'
import { HomeScreen } from './src/screens/HomeScreen'
import { LibraryBrowseScreen } from './src/screens/LibraryBrowseScreen'
import { MediaDetailScreen } from './src/screens/MediaDetailScreen'
import { SeriesDetailScreen } from './src/screens/SeriesDetailScreen'
import { VideoPlayerScreen } from './src/screens/VideoPlayerScreen'
import { SettingsScreen } from './src/screens/SettingsScreen'
import { ProfileScreen } from './src/screens/ProfileScreen'
import { MoviesScreen } from './src/screens/MoviesScreen'
import { SeriesScreen } from './src/screens/SeriesScreen'
import { BrowseScreen } from './src/screens/BrowseScreen'
import { FavoritesScreen } from './src/screens/FavoritesScreen'
import { SearchScreen } from './src/screens/SearchScreen'
import { ToastContainer } from './src/components/Toast'
import { AdminScreen } from './src/screens/AdminScreen'
import { NotificationBell } from './src/components/NotificationBell'

// =============================================================================
// Query Client
// =============================================================================

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 1,
    },
  },
})

// =============================================================================
// Navigation Types
// =============================================================================

// Root stack - main app-level navigation
export type RootStackParamList = {
  HomeTabs: undefined
  LibraryBrowse: { id: number; name: string }
  Media: { id: number }
  SeriesDetail: { id: number }
  Episode: { id: number; seriesId?: number; startFromBeginning?: boolean }
  Settings: undefined
  Profile: undefined
  Search: undefined
  Admin: undefined
  Login: undefined
}

// Home tab stack - Home tab screens
type HomeStackParamList = {
  HomeMain: undefined
  LibraryBrowse: { id: number; name: string }
  Media: { id: number }
  SeriesDetail: { id: number }
}

// Movies tab stack
type MoviesStackParamList = {
  MoviesMain: undefined
  Media: { id: number }
}

// Series tab stack
type SeriesStackParamList = {
  SeriesMain: undefined
  SeriesDetail: { id: number }
}

// Browse tab stack
type BrowseStackParamList = {
  BrowseMain: undefined
  LibraryBrowse: { id: number; name: string }
  Media: { id: number }
  SeriesDetail: { id: number }
}

// Favorites tab stack
type FavoritesStackParamList = {
  FavoritesMain: undefined
  Media: { id: number }
  SeriesDetail: { id: number }
}

// Search stack (modal-like, accessible from header)
type SearchStackParamList = {
  SearchMain: undefined
}

// Tab bar param list
export type HomeTabParamList = {
  Home: undefined
  Movies: undefined
  Series: undefined
  Browse: undefined
  Favorites: undefined
}

const RootStack = createNativeStackNavigator<RootStackParamList>()
const HomeStack = createNativeStackNavigator<HomeStackParamList>()
const MoviesStack = createNativeStackNavigator<MoviesStackParamList>()
const SeriesStack = createNativeStackNavigator<SeriesStackParamList>()
const BrowseStack = createNativeStackNavigator<BrowseStackParamList>()
const FavoritesStack = createNativeStackNavigator<FavoritesStackParamList>()
const SearchStack = createNativeStackNavigator<SearchStackParamList>()
const Tab = createBottomTabNavigator<HomeTabParamList>()

// =============================================================================
// Tab Icons
// =============================================================================

const TAB_ICONS: Record<string, typeof House> = {
  Home: House,
  Movies: Film,
  Series: Tv,
  Browse: FolderOpen,
  Favorites: Heart,
}

function TabIcon({ name, focused }: { name: string; focused: boolean }) {
  const Icon = TAB_ICONS[name] || House
  return (
    <Icon
      size={focused ? 24 : 22}
      color={focused ? '#e50914' : '#888'}
      fill={name === 'Favorites' && focused ? '#e50914' : 'none'}
    />
  )
}

// =============================================================================
// Stack Navigators for each Tab
// =============================================================================

function HomeStackNavigator() {
  return (
    <HomeStack.Navigator
      screenOptions={{
        headerStyle: { backgroundColor: '#141414' },
        headerTintColor: '#fff',
        headerTitleStyle: { fontWeight: 'bold', color: '#e50914', fontSize: 20 },
      }}
    >
      <HomeStack.Screen
        name="HomeMain"
        component={HomeScreen}
        options={({ navigation }) => ({
          title: 'VELOX',
          headerRight: () => <HeaderButtons navigation={navigation} />,
        })}
      />
      <HomeStack.Screen
        name="LibraryBrowse"
        component={LibraryBrowseScreen}
        options={({ route }) => ({ title: route.params.name })}
      />
      <HomeStack.Screen
        name="Media"
        component={MediaDetailScreen}
        options={{
          title: '',
          headerTransparent: true,
          headerTintColor: '#fff',
          headerStyle: { backgroundColor: 'transparent' },
        }}
      />
      <HomeStack.Screen
        name="SeriesDetail"
        component={SeriesDetailScreen}
        options={{
          title: '',
          headerTransparent: true,
          headerTintColor: '#fff',
          headerStyle: { backgroundColor: 'transparent' },
        }}
      />
    </HomeStack.Navigator>
  )
}

function MoviesStackNavigator() {
  return (
    <MoviesStack.Navigator
      screenOptions={{
        headerStyle: { backgroundColor: '#141414' },
        headerTintColor: '#fff',
        headerTitleStyle: { fontWeight: 'bold', color: '#e50914', fontSize: 20 },
      }}
    >
      <MoviesStack.Screen
        name="MoviesMain"
        component={MoviesScreen}
        options={({ navigation }) => ({
          title: 'Movies',
          headerRight: () => <HeaderButtons navigation={navigation} />,
        })}
      />
      <MoviesStack.Screen
        name="Media"
        component={MediaDetailScreen}
        options={{
          title: '',
          headerTransparent: true,
          headerTintColor: '#fff',
          headerStyle: { backgroundColor: 'transparent' },
        }}
      />
    </MoviesStack.Navigator>
  )
}

function SeriesStackNavigator() {
  return (
    <SeriesStack.Navigator
      screenOptions={{
        headerStyle: { backgroundColor: '#141414' },
        headerTintColor: '#fff',
        headerTitleStyle: { fontWeight: 'bold', color: '#e50914', fontSize: 20 },
      }}
    >
      <SeriesStack.Screen
        name="SeriesMain"
        component={SeriesScreen}
        options={({ navigation }) => ({
          title: 'Series',
          headerRight: () => <HeaderButtons navigation={navigation} />,
        })}
      />
      <SeriesStack.Screen
        name="SeriesDetail"
        component={SeriesDetailScreen}
        options={{
          title: '',
          headerTransparent: true,
          headerTintColor: '#fff',
          headerStyle: { backgroundColor: 'transparent' },
        }}
      />
    </SeriesStack.Navigator>
  )
}

function BrowseStackNavigator() {
  return (
    <BrowseStack.Navigator
      screenOptions={{
        headerStyle: { backgroundColor: '#141414' },
        headerTintColor: '#fff',
        headerTitleStyle: { fontWeight: 'bold', color: '#e50914', fontSize: 20 },
      }}
    >
      <BrowseStack.Screen
        name="BrowseMain"
        component={BrowseScreen}
        options={({ navigation }) => ({
          title: 'Browse',
          headerRight: () => <HeaderButtons navigation={navigation} />,
        })}
      />
      <BrowseStack.Screen
        name="LibraryBrowse"
        component={LibraryBrowseScreen}
        options={({ route }) => ({ title: route.params.name })}
      />
      <BrowseStack.Screen
        name="Media"
        component={MediaDetailScreen}
        options={{
          title: '',
          headerTransparent: true,
          headerTintColor: '#fff',
          headerStyle: { backgroundColor: 'transparent' },
        }}
      />
      <BrowseStack.Screen
        name="SeriesDetail"
        component={SeriesDetailScreen}
        options={{
          title: '',
          headerTransparent: true,
          headerTintColor: '#fff',
          headerStyle: { backgroundColor: 'transparent' },
        }}
      />
    </BrowseStack.Navigator>
  )
}

function FavoritesStackNavigator() {
  return (
    <FavoritesStack.Navigator
      screenOptions={{
        headerStyle: { backgroundColor: '#141414' },
        headerTintColor: '#fff',
        headerTitleStyle: { fontWeight: 'bold', color: '#e50914', fontSize: 20 },
      }}
    >
      <FavoritesStack.Screen
        name="FavoritesMain"
        component={FavoritesScreen}
        options={({ navigation }) => ({
          title: 'Favorites',
          headerRight: () => <HeaderButtons navigation={navigation} />,
        })}
      />
      <FavoritesStack.Screen
        name="Media"
        component={MediaDetailScreen}
        options={{
          title: '',
          headerTransparent: true,
          headerTintColor: '#fff',
          headerStyle: { backgroundColor: 'transparent' },
        }}
      />
      <FavoritesStack.Screen
        name="SeriesDetail"
        component={SeriesDetailScreen}
        options={{
          title: '',
          headerTransparent: true,
          headerTintColor: '#fff',
          headerStyle: { backgroundColor: 'transparent' },
        }}
      />
    </FavoritesStack.Navigator>
  )
}

function SearchStackNavigator() {
  return (
    <SearchStack.Navigator
      screenOptions={{
        headerStyle: { backgroundColor: '#141414' },
        headerTintColor: '#fff',
        headerTitleStyle: { fontWeight: 'bold', color: '#e50914', fontSize: 20 },
      }}
    >
      <SearchStack.Screen
        name="SearchMain"
        component={SearchScreen}
        options={{
          title: 'Search',
          headerShown: false,
        }}
      />
    </SearchStack.Navigator>
  )
}

// =============================================================================
// TV Sidebar Navigation
// =============================================================================

/** Tab definitions shared by the sidebar and the tab bar. */
const TV_TABS = [
  { key: 'Home' as const, label: 'Home', Icon: House },
  { key: 'Movies' as const, label: 'Movies', Icon: Film },
  { key: 'Series' as const, label: 'Series', Icon: Tv },
  { key: 'Browse' as const, label: 'Browse', Icon: FolderOpen },
  { key: 'Favorites' as const, label: 'Favorites', Icon: Heart },
] as const

/** Bottom-section items (Settings / Profile) rendered separately in the sidebar. */
const TV_SIDEBAR_BOTTOM = [
  { key: 'Settings', label: 'Settings', Icon: Settings },
  { key: 'Search', label: 'Search', Icon: Search },
  { key: 'Profile', label: 'Profile', Icon: User },
] as const

type TabKey = (typeof TV_TABS)[number]['key']

/** Map tab keys to their corresponding stack navigators. */
const TAB_SCREENS: Record<TabKey, React.ComponentType<any>> = {
  Home: HomeStackNavigator,
  Movies: MoviesStackNavigator,
  Series: SeriesStackNavigator,
  Browse: BrowseStackNavigator,
  Favorites: FavoritesStackNavigator,
}

function TVSidebarButton({
  label,
  Icon,
  active,
  onPress,
  isFirst,
}: {
  label: string
  Icon: typeof House
  active: boolean
  onPress: () => void
  isFirst?: boolean
}) {
  const [focused, setFocused] = useState(false)

  return (
    <TouchableOpacity
      activeOpacity={0.7}
      onPress={onPress}
      onFocus={() => setFocused(true)}
      onBlur={() => setFocused(false)}
      {...(isFirst ? { hasTVPreferredFocus: true } as any : {})}
      style={[
        tvSidebarStyles.button,
        active && tvSidebarStyles.buttonActive,
        focused && tvSidebarStyles.buttonFocused,
      ]}
    >
      <Icon
        size={22}
        color={active || focused ? '#e50914' : '#888'}
        fill={label === 'Favorites' && active ? '#e50914' : 'none'}
      />
      <Text
        style={[
          tvSidebarStyles.buttonLabel,
          (active || focused) && tvSidebarStyles.buttonLabelActive,
        ]}
      >
        {label}
      </Text>
    </TouchableOpacity>
  )
}

/**
 * TV-specific home layout: left sidebar + content area on the right.
 * The sidebar replaces the bottom tab bar which is unsuitable for D-pad navigation.
 */
function TVHomeTabs({ navigation }: { navigation: any }) {
  const [activeTab, setActiveTab] = useState<TabKey>('Home')
  const ActiveScreen = TAB_SCREENS[activeTab]

  return (
    <View style={tvSidebarStyles.container}>
      {/* ── Left sidebar ────────────────────────────────────────────── */}
      <View style={tvSidebarStyles.sidebar}>
        {/* Brand */}
        <Text style={tvSidebarStyles.brand}>VELOX</Text>

        {/* Main tabs */}
        <ScrollView style={tvSidebarStyles.tabList} contentContainerStyle={{ gap: 4 }}>
          {TV_TABS.map((tab, index) => (
            <TVSidebarButton
              key={tab.key}
              label={tab.label}
              Icon={tab.Icon}
              active={activeTab === tab.key}
              onPress={() => setActiveTab(tab.key)}
              isFirst={index === 0}
            />
          ))}
        </ScrollView>

        {/* Bottom section (Settings, Search, Profile) */}
        <View style={tvSidebarStyles.bottomSection}>
          {TV_SIDEBAR_BOTTOM.map((item) => (
            <TVSidebarButton
              key={item.key}
              label={item.label}
              Icon={item.Icon}
              active={false}
              onPress={() => navigation.navigate(item.key)}
            />
          ))}
        </View>
      </View>

      {/* ── Content area ────────────────────────────────────────────── */}
      <View style={tvSidebarStyles.content}>
        <ActiveScreen />
      </View>
    </View>
  )
}

const tvSidebarStyles = StyleSheet.create({
  container: {
    flex: 1,
    flexDirection: 'row',
    backgroundColor: '#141414',
  },
  sidebar: {
    width: TV_SIDEBAR_WIDTH,
    backgroundColor: '#1a1a1a',
    borderRightWidth: 1,
    borderRightColor: '#2a2a2a',
    paddingTop: 48,
    paddingBottom: 24,
    paddingHorizontal: 12,
  },
  brand: {
    color: '#e50914',
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 32,
    paddingHorizontal: 12,
  },
  tabList: {
    flex: 1,
  },
  bottomSection: {
    borderTopWidth: 1,
    borderTopColor: '#2a2a2a',
    paddingTop: 12,
    gap: 4,
  },
  button: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 14,
    paddingVertical: 14,
    paddingHorizontal: 12,
    borderRadius: 8,
  },
  buttonActive: {
    backgroundColor: 'rgba(229, 9, 20, 0.15)',
  },
  buttonFocused: {
    backgroundColor: 'rgba(229, 9, 20, 0.25)',
    borderWidth: 2,
    borderColor: '#e50914',
  },
  buttonLabel: {
    color: '#888',
    fontSize: 16,
    fontWeight: '500',
  },
  buttonLabelActive: {
    color: '#fff',
    fontWeight: '600',
  },
  content: {
    flex: 1,
  },
})

// =============================================================================
// Home Tabs Navigator (phone / tablet -- bottom tab bar)
// =============================================================================

function HomeTabs() {
  if (isTV) {
    // TV layout is handled by TVHomeTabs which is registered as a screen
    // that receives navigation.  This branch should not be reached because
    // AppNavigator renders TVHomeTabs directly, but guard defensively.
    return null
  }

  return (
    <Tab.Navigator
      screenOptions={({ route }: { route: { name: string } }) => ({
        headerShown: false,
        tabBarStyle: styles.tabBar,
        tabBarActiveTintColor: '#e50914',
        tabBarInactiveTintColor: '#888',
        tabBarLabelStyle: styles.tabLabel,
        tabBarIcon: ({ focused }: { focused: boolean }) => <TabIcon name={route.name} focused={focused} />,
      })}
    >
      <Tab.Screen name="Home" component={HomeStackNavigator} />
      <Tab.Screen name="Movies" component={MoviesStackNavigator} />
      <Tab.Screen name="Series" component={SeriesStackNavigator} />
      <Tab.Screen name="Browse" component={BrowseStackNavigator} />
      <Tab.Screen name="Favorites" component={FavoritesStackNavigator} />
    </Tab.Navigator>
  )
}

// =============================================================================
// Header Buttons Component
// =============================================================================

function HeaderButtons({ navigation }: { navigation: any }) {
  return (
    <View style={styles.headerRight}>
      <TouchableOpacity
        style={styles.headerButton}
        onPress={() => navigation.navigate('Search')}
      >
        <Search size={18} color="#fff" />
      </TouchableOpacity>
      <NotificationBell />
      <TouchableOpacity
        style={styles.headerButton}
        onPress={() => navigation.navigate('Profile')}
      >
        <User size={18} color="#fff" />
      </TouchableOpacity>
      <TouchableOpacity
        style={styles.headerButton}
        onPress={() => navigation.navigate('Settings')}
      >
        <Settings size={18} color="#fff" />
      </TouchableOpacity>
    </View>
  )
}

// =============================================================================
// Auth Gate Component
// =============================================================================

function AuthGate({ children }: { children: React.ReactNode }) {
  const [isReady, setIsReady] = useState(false)
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)

  useEffect(() => {
    // Load saved server URL, then check auth state
    // Use setInterval to wait for Zustand persist to fully rehydrate
    loadServerUrl().then(() => {
      const timer = setInterval(() => {
        const state = useAuthStore.getState()
        // Check if persist has rehydrated by looking for any persisted field
        if (state.accessToken !== undefined || state.refreshToken !== undefined) {
          setIsReady(true)
          clearInterval(timer)
        }
      }, 100)

      // Timeout after 5 seconds - persist should have rehydrated by now
      setTimeout(() => {
        setIsReady(true)
        clearInterval(timer)
      }, 5000)
    })
  }, [])

  if (!isReady) {
    return (
      <View style={styles.loading}>
        <ActivityIndicator size="large" color="#e50914" />
      </View>
    )
  }

  if (!isAuthenticated) {
    return (
      <LoginScreen
        onLoginSuccess={() => {}}
      />
    )
  }

  return <>{children}</>
}

// =============================================================================
// Main App Component
// =============================================================================

function AppNavigator() {
  return (
    <RootStack.Navigator
      screenOptions={{
        headerStyle: { backgroundColor: '#141414' },
        headerTintColor: '#fff',
      }}
    >
      <RootStack.Screen
        name="HomeTabs"
        component={isTV ? TVHomeTabs : HomeTabs}
        options={{ headerShown: false }}
      />
      <RootStack.Screen
        name="Settings"
        component={SettingsScreen}
        options={{ title: 'Settings' }}
      />
      <RootStack.Screen
        name="Profile"
        component={ProfileScreen}
        options={{ title: 'Profile' }}
      />
      <RootStack.Screen
        name="Login"
        component={LoginScreen}
        options={{ headerShown: false }}
      />
      <RootStack.Screen
        name="Search"
        component={SearchStackNavigator}
        options={{ headerShown: false }}
      />
      <RootStack.Screen
        name="Admin"
        component={AdminScreen}
        options={{ title: 'Admin' }}
      />
      <RootStack.Screen
        name="Media"
        component={MediaDetailScreen}
        options={{
          title: '',
          headerTransparent: true,
          headerTintColor: '#fff',
          headerStyle: { backgroundColor: 'transparent' },
        }}
      />
      <RootStack.Screen
        name="SeriesDetail"
        component={SeriesDetailScreen}
        options={{
          title: '',
          headerTransparent: true,
          headerTintColor: '#fff',
          headerStyle: { backgroundColor: 'transparent' },
        }}
      />
      <RootStack.Screen
        name="Episode"
        component={VideoPlayerScreen}
        options={{ headerShown: false, animation: 'fade' }}
      />
    </RootStack.Navigator>
  )
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthGate>
        <NavigationContainer>
          <AppNavigator />
        </NavigationContainer>
        <ToastContainer />
      </AuthGate>
      <StatusBar style="light" />
    </QueryClientProvider>
  )
}

// =============================================================================
// Styles
// =============================================================================

const styles = StyleSheet.create({
  loading: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#141414',
  },
  headerRight: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
  },
  headerButton: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  // Tab bar styles
  tabBar: {
    backgroundColor: '#0a0a0a',
    borderTopColor: '#1f1f1f',
    borderTopWidth: 1,
    paddingTop: 8,
    paddingBottom: 8,
    height: 65,
  },
  tabLabel: {
    fontSize: 11,
    fontWeight: '500',
  },
})
