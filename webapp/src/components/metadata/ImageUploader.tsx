import { useRef, useState } from 'react'
import { LuUpload, LuTrash2 } from 'react-icons/lu'

interface ImageUploaderProps {
  label: string
  currentUrl?: string
  onUpload: (file: File) => void
  onDelete?: () => void
  isUploading?: boolean
}

export function ImageUploader({
  label,
  currentUrl,
  onUpload,
  onDelete,
  isUploading,
}: ImageUploaderProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const [dragOver, setDragOver] = useState(false)

  function handleFile(file: File) {
    if (!file.type.startsWith('image/')) return
    if (file.size > 10 * 1024 * 1024) {
      alert('File too large (max 10MB)')
      return
    }
    setPreview(URL.createObjectURL(file))
    onUpload(file)
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault()
    setDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file) handleFile(file)
  }

  const displayUrl = preview || currentUrl

  return (
    <div className="space-y-2">
      <label className="text-sm font-medium text-gray-300">{label}</label>
      <div
        className={`relative flex min-h-[120px] items-center justify-center rounded-lg border-2 border-dashed transition-colors ${
          dragOver ? 'border-blue-500 bg-blue-500/10' : 'border-gray-600 bg-[#2a2a2a]'
        }`}
        onDragOver={(e) => {
          e.preventDefault()
          setDragOver(true)
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
        onClick={() => inputRef.current?.click()}
      >
        {displayUrl ? (
          <img src={displayUrl} alt={label} className="max-h-[200px] rounded object-contain" />
        ) : (
          <div className="flex flex-col items-center gap-2 p-4 text-gray-400">
            <LuUpload size={24} />
            <span className="text-sm">Drop image or click to browse</span>
            <span className="text-xs text-gray-500">JPEG, PNG, WebP • Max 10MB</span>
          </div>
        )}
        {isUploading && (
          <div className="absolute inset-0 flex items-center justify-center rounded-lg bg-black/50">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
          </div>
        )}
        <input
          ref={inputRef}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          className="hidden"
          onChange={(e) => {
            const file = e.target.files?.[0]
            if (file) handleFile(file)
          }}
        />
      </div>
      {displayUrl && onDelete && (
        <button
          onClick={(e) => {
            e.stopPropagation()
            setPreview(null)
            onDelete()
          }}
          className="flex items-center gap-1 text-xs text-red-400 hover:text-red-300"
        >
          <LuTrash2 size={12} /> Remove custom image
        </button>
      )}
    </div>
  )
}
