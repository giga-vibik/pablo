import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import PlatformIcon from '../components/PlatformIcon'
import { createPost, publishPost, uploadVideo } from '../api'
import {
  CAPTION_LIMITS,
  PLATFORM_COLORS,
  PLATFORM_LABELS,
  PLATFORM_SHORT,
  VIDEO_PLATFORMS,
  type Platform,
  type PostKind,
} from '../types'

export default function NewPostPage() {
  const navigate = useNavigate()

  const [kind, setKind] = useState<PostKind>('video')
  const [content, setContent] = useState('')
  const [captions, setCaptions] = useState<Partial<Record<Platform, string>>>({})
  const [selected, setSelected] = useState<Platform[]>(VIDEO_PLATFORMS)
  const [file, setFile] = useState<File | null>(null)
  const [dragOver, setDragOver] = useState(false)
  const [scheduledAt, setScheduledAt] = useState('')
  const [previewOf, setPreviewOf] = useState<Platform>('instagram')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  // Threads принимает только текст, остальные площадки — только видео,
  // поэтому набор платформ жёстко зависит от типа поста.
  const platforms: Platform[] = kind === 'video' ? VIDEO_PLATFORMS : ['threads']

  // Локальный URL живёт, пока выбран этот файл; иначе утечёт при каждой замене.
  const fileURL = useMemo(() => (file ? URL.createObjectURL(file) : ''), [file])
  useEffect(() => () => { if (fileURL) URL.revokeObjectURL(fileURL) }, [fileURL])

  useEffect(() => {
    if (!selected.includes(previewOf) && selected.length > 0) {
      setPreviewOf(selected[0])
    }
  }, [selected, previewOf])

  const switchKind = (next: PostKind) => {
    setKind(next)
    setSelected(next === 'video' ? VIDEO_PLATFORMS : ['threads'])
    setPreviewOf(next === 'video' ? 'instagram' : 'threads')
    setFile(null)
  }

  const togglePlatform = (platform: Platform) => {
    setSelected((prev) =>
      prev.includes(platform) ? prev.filter((p) => p !== platform) : [...prev, platform],
    )
  }

  const onDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)

    const dropped = e.dataTransfer.files?.[0]
    if (dropped?.type.startsWith('video/')) {
      setFile(dropped)
    }
  }

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (selected.length === 0) {
      setError('Выберите хотя бы одну площадку')
      return
    }

    if (kind === 'video' && !file) {
      setError('Прикрепите видео')
      return
    }

    setSaving(true)

    try {
      const post = await createPost({
        kind,
        content,
        scheduled_at: scheduledAt ? new Date(scheduledAt).toISOString() : undefined,
        targets: selected.map((platform) => ({
          platform,
          caption: captions[platform] || undefined,
        })),
      })

      if (file) {
        await uploadVideo(post.id, file)
      }

      // Без расписания публикуем сразу — иначе пост подхватит воркер.
      if (!scheduledAt) {
        await publishPost(post.id)
      }

      navigate('/posts')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось создать пост')
    } finally {
      setSaving(false)
    }
  }

  const previewText = captions[previewOf] || content

  return (
    <>
      <div className="page-header">
        <div>
          <h1 className="title">Новый пост</h1>
          <p className="subtitle">
            {kind === 'video'
              ? 'Одно видео — во все выбранные площадки'
              : 'Текстовая заметка в Threads'}
          </p>
        </div>
        <button className="ghost" onClick={() => navigate('/posts')}>
          Отмена
        </button>
      </div>

      {error && <div className="banner">{error}</div>}

      <form className="composer" onSubmit={submit}>
        <div>
          <div className="field">
            <div className="segmented">
              <button
                type="button"
                className={kind === 'video' ? 'on' : ''}
                onClick={() => switchKind('video')}
              >
                Видео
              </button>
              <button
                type="button"
                className={kind === 'text' ? 'on' : ''}
                onClick={() => switchKind('text')}
              >
                Текст
              </button>
            </div>
          </div>

          {kind === 'video' && (
            <div className="field">
              <label className="field-label">Видео</label>
              <div
                className={`dropzone ${dragOver ? 'over' : ''}`}
                onDragOver={(e) => {
                  e.preventDefault()
                  setDragOver(true)
                }}
                onDragLeave={() => setDragOver(false)}
                onDrop={onDrop}
              >
                <input
                  type="file"
                  accept="video/mp4,video/quicktime"
                  onChange={(e) => setFile(e.target.files?.[0] ?? null)}
                />
                {file ? (
                  <>
                    <strong>{file.name}</strong>
                    {formatSize(file.size)} · нажмите, чтобы заменить
                  </>
                ) : (
                  <>
                    <strong>Перетащите видео сюда</strong>
                    MP4 или MOV, вертикальное 9:16
                  </>
                )}
              </div>
            </div>
          )}

          <div className="field">
            <label className="field-label">Общий текст</label>
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Уйдёт туда, где не задан свой текст"
            />
          </div>

          <div className="field">
            <label className="field-label">Площадки</label>
            <div className="picker">
              {platforms.map((platform) => {
                const on = selected.includes(platform)
                const caption = captions[platform] ?? ''
                const limit = CAPTION_LIMITS[platform]
                const used = (caption || content).length

                return (
                  <div
                    className={`pick ${on ? 'on' : ''}`}
                    key={platform}
                    style={{ ['--pick-color' as string]: PLATFORM_COLORS[platform] }}
                  >
                    <div className="pick-head" onClick={() => togglePlatform(platform)}>
                      <span
                        className={`mark ${on ? 'published' : ''}`}
                        style={{ ['--mark-color' as string]: PLATFORM_COLORS[platform] }}
                      >
                        <PlatformIcon platform={platform} />
                      </span>
                      <span className="name">{PLATFORM_LABELS[platform]}</span>
                      <span className="state">{on ? '' : 'выключено'}</span>
                      <span className="switch" />
                    </div>

                    {on && (
                      <div className="pick-body">
                        <textarea
                          value={caption}
                          onChange={(e) =>
                            setCaptions((prev) => ({ ...prev, [platform]: e.target.value }))
                          }
                          onFocus={() => setPreviewOf(platform)}
                          placeholder="Свой текст для этой площадки"
                        />
                        <div className={`counter ${used > limit ? 'over' : ''}`}>
                          {used} / {limit}
                        </div>
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          </div>

          <div className="field">
            <label className="field-label">Когда публиковать</label>
            <input
              type="datetime-local"
              value={scheduledAt}
              onChange={(e) => setScheduledAt(e.target.value)}
            />
            <div className="hint">Пусто — публикуем прямо сейчас</div>
          </div>

          <button className="primary" type="submit" disabled={saving}>
            {saving ? 'Сохраняем…' : scheduledAt ? 'Запланировать' : 'Опубликовать'}
          </button>
        </div>

        <aside className="preview-col">
          <div className="preview-tabs">
            {selected.map((platform) => (
              <button
                type="button"
                key={platform}
                className={`mark ${previewOf === platform ? 'published' : ''}`}
                style={{ ['--mark-color' as string]: PLATFORM_COLORS[platform] }}
                onClick={() => setPreviewOf(platform)}
                title={PLATFORM_SHORT[platform]}
              >
                <PlatformIcon platform={platform} />
              </button>
            ))}
          </div>

          <div className="phone">
            <div className="phone-screen">
              {fileURL ? (
                <video src={fileURL} muted loop autoPlay playsInline />
              ) : (
                <div className="phone-placeholder">
                  {kind === 'video' ? 'Превью появится после загрузки видео' : 'Threads'}
                </div>
              )}

              {previewText && <div className="phone-overlay">{previewText}</div>}
            </div>
          </div>

          <div className="hint" style={{ textAlign: 'center' }}>
            {selected.length > 0
              ? `${PLATFORM_SHORT[previewOf]} · так увидят подписчики`
              : 'Площадки не выбраны'}
          </div>
        </aside>
      </form>
    </>
  )
}

function formatSize(bytes: number): string {
  const mb = bytes / (1024 * 1024)
  return mb >= 1 ? `${mb.toFixed(1)} МБ` : `${Math.round(bytes / 1024)} КБ`
}
