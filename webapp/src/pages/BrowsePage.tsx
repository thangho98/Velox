import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { useFolderBrowse } from '@/hooks/stores/useMedia'
import { useTranslation } from '@/hooks/useTranslation'
import { Breadcrumb } from '@/components/Breadcrumb'
import { FolderCard } from '@/components/FolderCard'
import { MediaCard } from '@/components/MediaCard'
import { LuFolder, LuChevronLeft } from 'react-icons/lu'

export function BrowsePage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const { t } = useTranslation('media')

  // Parse URL: ?library_id=1&path=Action/Marvel
  // Root (no params) → shows all libraries
  // lib:N in path → resolves to library_id + path inside it
  const libraryId = Number(searchParams.get('library_id')) || undefined
  const currentPath = searchParams.get('path') || ''

  const { data: result, isLoading } = useFolderBrowse(libraryId, currentPath)

  // Build breadcrumb
  const breadcrumbs = useMemo(() => {
    const crumbs = [{ name: t('browse.title'), path: '', isRoot: true }]

    if (!libraryId && !currentPath) return crumbs

    // When inside a library, add library name as first crumb after root
    // (we get it from the folder name when user clicked from root)
    if (currentPath) {
      const parts = currentPath.split('/').filter(Boolean)
      let accumulated = ''
      parts.forEach((part) => {
        accumulated = accumulated ? accumulated + '/' + part : part
        crumbs.push({
          name: part.startsWith('root:') ? part : part,
          path: accumulated,
          isRoot: false,
        })
      })
    }

    return crumbs
  }, [libraryId, currentPath])

  const handleNavigate = (path: string) => {
    // Clicking a library folder (lib:N) → set library_id, clear path
    if (path.startsWith('lib:')) {
      const libId = path.slice(4)
      setSearchParams({ library_id: libId }, { replace: true })
      return
    }

    // Back to root (all libraries)
    if (!path && !libraryId) {
      setSearchParams({}, { replace: true })
      return
    }

    // Navigate within library
    const newParams = new URLSearchParams()
    if (libraryId) newParams.set('library_id', String(libraryId))
    if (path) newParams.set('path', path)
    setSearchParams(newParams, { replace: true })
  }

  const handleBack = () => {
    if (result?.parent != null && result.parent !== '') {
      handleNavigate(result.parent)
    } else if (currentPath) {
      // Back to library root
      const newParams = new URLSearchParams()
      if (libraryId) newParams.set('library_id', String(libraryId))
      setSearchParams(newParams, { replace: true })
    } else {
      // Back to all libraries
      setSearchParams({}, { replace: true })
    }
  }

  const folders = result?.folders ?? []
  const media = result?.media ?? []
  const isEmpty = folders.length === 0 && media.length === 0
  const showBack = !!(libraryId || currentPath)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4">
        <h1 className="text-3xl font-bold text-white">{t('browse.title')}</h1>

        {showBack && (
          <div className="flex items-center gap-4">
            <button
              onClick={handleBack}
              className="flex items-center gap-1 text-sm text-gray-400 hover:text-white transition-colors"
            >
              <LuChevronLeft className="w-4 h-4" />
              {t('browse.back')}
            </button>
            <Breadcrumb items={breadcrumbs} onNavigate={handleNavigate} />
          </div>
        )}
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-[#e50914] border-t-transparent" />
        </div>
      ) : isEmpty ? (
        <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-[#1a1a1a]">
          <LuFolder className="h-12 w-12 text-gray-600 mb-4" />
          <p className="text-gray-400">{t('browse.empty')}</p>
        </div>
      ) : (
        <div className="space-y-8">
          {/* Folders */}
          {folders.length > 0 && (
            <section>
              {(media.length > 0 || currentPath) && (
                <h2 className="mb-4 text-lg font-semibold text-white">
                  {t('browse.folders', { count: folders.length })}
                </h2>
              )}
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
                {folders.map((folder) => (
                  <FolderCard
                    key={folder.path}
                    name={folder.name}
                    path={folder.path}
                    poster={folder.poster}
                    mediaCount={folder.media_count}
                    onClick={() => handleNavigate(folder.path)}
                  />
                ))}
              </div>
            </section>
          )}

          {/* Media in this folder */}
          {media.length > 0 && (
            <section>
              <h2 className="mb-4 text-lg font-semibold text-white">Media ({media.length})</h2>
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
                {media.map((item) => (
                  <MediaCard
                    key={item.id}
                    id={item.id}
                    title={item.title}
                    posterPath={item.poster_path}
                    type={item.media_type === 'episode' ? 'series' : 'movie'}
                    seriesId={item.series_id || undefined}
                    year={item.release_date ? new Date(item.release_date).getFullYear() : undefined}
                    rating={item.rating}
                  />
                ))}
              </div>
            </section>
          )}
        </div>
      )}
    </div>
  )
}
