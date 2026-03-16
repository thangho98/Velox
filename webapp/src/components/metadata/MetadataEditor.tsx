import { useState } from 'react'
import { LuX, LuSave } from 'react-icons/lu'
import type {
  Media,
  Series,
  MetadataEditRequest,
  SeriesMetadataEditRequest,
  CreditInput,
} from '@/types/api'
import type { CreditWithPerson } from '@/types/api'
import { GenreEditor } from './GenreEditor'
import { CreditEditor } from './CreditEditor'
import { ImageUploader } from './ImageUploader'
import { tmdbImage } from '@/lib/image'

interface MediaEditorProps {
  type: 'media'
  media: Media
  genres: string[]
  credits: CreditWithPerson[]
  onSave: (req: MetadataEditRequest) => void
  onUploadImage: (imageType: string, file: File) => void
  isSaving: boolean
  isUploadingImage: boolean
  onClose: () => void
}

interface SeriesEditorProps {
  type: 'series'
  series: Series
  genres: string[]
  credits: CreditWithPerson[]
  onSave: (req: SeriesMetadataEditRequest) => void
  onUploadImage: (imageType: string, file: File) => void
  isSaving: boolean
  isUploadingImage: boolean
  onClose: () => void
}

type MetadataEditorProps = MediaEditorProps | SeriesEditorProps

export function MetadataEditor(props: MetadataEditorProps) {
  const isMedia = props.type === 'media'
  const entity = isMedia ? props.media : props.series
  const initialGenres = props.genres ?? []
  const initialCredits = props.credits ?? []

  const [title, setTitle] = useState(entity.title)
  const [sortTitle, setSortTitle] = useState(entity.sort_title)
  const [overview, setOverview] = useState(entity.overview)
  const [tagline, setTagline] = useState(isMedia ? props.media.tagline : '')
  const [releaseDate, setReleaseDate] = useState(
    isMedia ? props.media.release_date : props.series.first_air_date,
  )
  const [rating, setRating] = useState(isMedia ? props.media.rating : 0)
  const [status, setStatus] = useState(isMedia ? '' : props.series.status)
  const [network, setNetwork] = useState(isMedia ? '' : props.series.network)
  const [genres, setGenres] = useState<string[]>(initialGenres)
  const [genresDirty, setGenresDirty] = useState(false)
  const [credits, setCredits] = useState<CreditInput[]>(
    initialCredits.map((c) => ({
      person_name: c.person.name,
      character: c.credit.character,
      role: c.credit.role as 'cast' | 'director' | 'writer',
      order: c.credit.display_order,
    })),
  )
  const [creditsDirty, setCreditsDirty] = useState(false)
  const [saveNfo, setSaveNfo] = useState(false)
  const [lockMetadata, setLockMetadata] = useState(true)

  function handleGenresChange(newGenres: string[]) {
    setGenres(newGenres)
    setGenresDirty(true)
  }

  function handleCreditsChange(newCredits: CreditInput[]) {
    setCredits(newCredits)
    setCreditsDirty(true)
  }

  function handleSave() {
    if (isMedia) {
      const req: MetadataEditRequest = {
        title,
        sort_title: sortTitle,
        overview,
        tagline,
        release_date: releaseDate,
        rating,
        genres: genresDirty ? genres : undefined,
        credits: creditsDirty ? credits.filter((c) => c.person_name.trim()) : undefined,
        save_nfo: saveNfo,
        metadata_locked: lockMetadata,
      }
      props.onSave(req)
    } else {
      const req: SeriesMetadataEditRequest = {
        title,
        sort_title: sortTitle,
        overview,
        status,
        network,
        first_air_date: releaseDate,
        genres: genresDirty ? genres : undefined,
        credits: creditsDirty ? credits.filter((c) => c.person_name.trim()) : undefined,
        save_nfo: saveNfo,
        metadata_locked: lockMetadata,
      }
      props.onSave(req)
    }
  }

  const posterUrl = entity.poster_path
    ? entity.poster_path.startsWith('local://')
      ? `/api/images/local/${props.type}/${entity.poster_path.slice(8)}`
      : tmdbImage(entity.poster_path, 'w342')
    : undefined

  const backdropUrl = entity.backdrop_path
    ? entity.backdrop_path.startsWith('local://')
      ? `/api/images/local/${props.type}/${entity.backdrop_path.slice(8)}`
      : tmdbImage(entity.backdrop_path, 'w780')
    : undefined

  return (
    <div className="fixed inset-0 z-50 flex justify-end">
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/60" onClick={props.onClose} />

      {/* Panel */}
      <div className="relative flex h-full w-full max-w-xl flex-col overflow-y-auto bg-[#1a1a1a] shadow-2xl">
        {/* Header */}
        <div className="sticky top-0 z-10 flex items-center justify-between border-b border-gray-700 bg-[#1a1a1a] px-6 py-4">
          <h2 className="text-lg font-semibold text-white">Edit Metadata</h2>
          <button onClick={props.onClose} className="text-gray-400 hover:text-white">
            <LuX size={22} />
          </button>
        </div>

        {/* Body */}
        <div className="flex-1 space-y-6 p-6">
          {/* Basic Info */}
          <section className="space-y-3">
            <h3 className="text-sm font-semibold uppercase tracking-wider text-gray-400">
              Basic Info
            </h3>
            <div>
              <label className="mb-1 block text-sm text-gray-300">Title</label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="w-full rounded-lg bg-[#2a2a2a] px-3 py-2 text-white outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm text-gray-300">Sort Title</label>
              <input
                type="text"
                value={sortTitle}
                onChange={(e) => setSortTitle(e.target.value)}
                className="w-full rounded-lg bg-[#2a2a2a] px-3 py-2 text-white outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            {isMedia && (
              <div>
                <label className="mb-1 block text-sm text-gray-300">Tagline</label>
                <input
                  type="text"
                  value={tagline}
                  onChange={(e) => setTagline(e.target.value)}
                  className="w-full rounded-lg bg-[#2a2a2a] px-3 py-2 text-white outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
            )}
            <div>
              <label className="mb-1 block text-sm text-gray-300">Overview</label>
              <textarea
                value={overview}
                onChange={(e) => setOverview(e.target.value)}
                rows={4}
                className="w-full rounded-lg bg-[#2a2a2a] px-3 py-2 text-white outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </section>

          {/* Dates & Ratings */}
          <section className="space-y-3">
            <h3 className="text-sm font-semibold uppercase tracking-wider text-gray-400">
              {isMedia ? 'Release & Rating' : 'Series Info'}
            </h3>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="mb-1 block text-sm text-gray-300">
                  {isMedia ? 'Release Date' : 'First Air Date'}
                </label>
                <input
                  type="date"
                  value={releaseDate}
                  onChange={(e) => setReleaseDate(e.target.value)}
                  className="w-full rounded-lg bg-[#2a2a2a] px-3 py-2 text-white outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              {isMedia && (
                <div>
                  <label className="mb-1 block text-sm text-gray-300">Rating</label>
                  <input
                    type="number"
                    value={rating}
                    onChange={(e) => setRating(Number(e.target.value))}
                    min={0}
                    max={10}
                    step={0.1}
                    className="w-full rounded-lg bg-[#2a2a2a] px-3 py-2 text-white outline-none focus:ring-1 focus:ring-blue-500"
                  />
                </div>
              )}
              {!isMedia && (
                <>
                  <div>
                    <label className="mb-1 block text-sm text-gray-300">Status</label>
                    <select
                      value={status}
                      onChange={(e) => setStatus(e.target.value)}
                      className="w-full rounded-lg bg-[#2a2a2a] px-3 py-2 text-white outline-none"
                    >
                      <option value="">Unknown</option>
                      <option value="Returning Series">Returning Series</option>
                      <option value="Ended">Ended</option>
                      <option value="Canceled">Canceled</option>
                    </select>
                  </div>
                  <div className="col-span-2">
                    <label className="mb-1 block text-sm text-gray-300">Network</label>
                    <input
                      type="text"
                      value={network}
                      onChange={(e) => setNetwork(e.target.value)}
                      className="w-full rounded-lg bg-[#2a2a2a] px-3 py-2 text-white outline-none focus:ring-1 focus:ring-blue-500"
                    />
                  </div>
                </>
              )}
            </div>
          </section>

          {/* Images */}
          <section className="space-y-3">
            <h3 className="text-sm font-semibold uppercase tracking-wider text-gray-400">Images</h3>
            <div className="grid grid-cols-2 gap-4">
              <ImageUploader
                label="Poster"
                currentUrl={posterUrl}
                onUpload={(file) => props.onUploadImage('poster', file)}
                isUploading={props.isUploadingImage}
              />
              <ImageUploader
                label="Backdrop"
                currentUrl={backdropUrl}
                onUpload={(file) => props.onUploadImage('backdrop', file)}
                isUploading={props.isUploadingImage}
              />
            </div>
          </section>

          {/* Genres */}
          <section>
            <h3 className="mb-3 text-sm font-semibold uppercase tracking-wider text-gray-400">
              Genres
            </h3>
            <GenreEditor genres={genres} onChange={handleGenresChange} />
          </section>

          {/* Credits */}
          <section>
            <h3 className="mb-3 text-sm font-semibold uppercase tracking-wider text-gray-400">
              Cast & Crew
            </h3>
            <CreditEditor credits={credits} onChange={handleCreditsChange} />
          </section>

          {/* Options */}
          <section className="space-y-3">
            <h3 className="text-sm font-semibold uppercase tracking-wider text-gray-400">
              Options
            </h3>
            <label className="flex items-center gap-3 text-sm text-gray-300">
              <input
                type="checkbox"
                checked={lockMetadata}
                onChange={(e) => setLockMetadata(e.target.checked)}
                className="h-4 w-4 rounded accent-blue-500"
              />
              Lock Metadata (skip on rescan)
            </label>
            <label className="flex items-center gap-3 text-sm text-gray-300">
              <input
                type="checkbox"
                checked={saveNfo}
                onChange={(e) => setSaveNfo(e.target.checked)}
                className="h-4 w-4 rounded accent-blue-500"
              />
              Save NFO file
            </label>
          </section>
        </div>

        {/* Footer */}
        <div className="sticky bottom-0 flex items-center justify-end gap-3 border-t border-gray-700 bg-[#1a1a1a] px-6 py-4">
          <button
            onClick={props.onClose}
            className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-white"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={props.isSaving || !title.trim()}
            className="flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2 text-sm font-medium text-white hover:bg-blue-500 disabled:opacity-50"
          >
            <LuSave size={16} />
            {props.isSaving ? 'Saving...' : 'Save'}
          </button>
        </div>
      </div>
    </div>
  )
}
