// TanStack Query Hooks - Auth
export {
  useMe,
  useLogin,
  useLogout,
  useRefreshToken,
  useChangePassword,
  useSessions,
  useRevokeSession,
  useSetupStatus,
  useSetup,
  useProfile,
  useUpdateProfile,
  usePreferences,
  useUpdatePreferences,
  authKeys,
  profileKeys,
} from './stores/useAuth'

// TanStack Query Hooks - Media & Libraries
export {
  useLibraries,
  useCreateLibrary,
  useDeleteLibrary,
  useScanLibrary,
  useMediaList,
  useMedia,
  useMediaWithFiles,
  useProgress,
  useUpdateProgress,
  useFavorites,
  useToggleFavorite,
  useRecentlyWatched,
  libraryKeys,
  mediaKeys,
  userDataKeys,
} from './stores/useMedia'

// TanStack Query Hooks - User Management (Admin)
export {
  useUsers,
  useCreateUser,
  useUpdateUser,
  useDeleteUser,
  useSetLibraryAccess,
  userKeys,
} from './stores/useUsers'
