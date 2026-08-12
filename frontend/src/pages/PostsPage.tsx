import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'

import PlatformMarks from '../components/PlatformMarks'
import { deletePost, listPosts, publishPost } from '../api'
import type { Post } from '../types'

const STATUS_LABEL: Record<Post['status'], string> = {
  draft: 'Черновик',
  scheduled: 'Запланирован',
  publishing: 'Публикуется',
  published: 'Опубликован',
  partially_published: 'Частично',
  failed: 'Ошибка',
}

export default function PostsPage() {
  const [posts, setPosts] = useState<Post[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [busyID, setBusyID] = useState('')

  const load = async () => {
    setError('')
    try {
      const res = await listPosts()
      setPosts(res.posts ?? [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось загрузить посты')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const onPublish = async (postID: string) => {
    setBusyID(postID)
    setError('')
    try {
      await publishPost(postID)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось опубликовать')
    } finally {
      setBusyID('')
    }
  }

  const onDelete = async (postID: string) => {
    if (!confirm('Удалить пост?')) return

    setBusyID(postID)
    try {
      await deletePost(postID)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось удалить')
    } finally {
      setBusyID('')
    }
  }

  return (
    <>
      <div className="page-header">
        <div>
          <h1 className="title">Лента</h1>
          <p className="subtitle">
            {loading ? 'Загружаем…' : `${posts.length} ${plural(posts.length)}`}
          </p>
        </div>

        <Link to="/posts/new">
          <button className="primary">Новый пост</button>
        </Link>
      </div>

      {error && <div className="banner">{error}</div>}

      {loading && (
        <div className="grid">
          {[0, 1, 2, 3].map((i) => (
            <div className="skeleton" key={i} style={{ aspectRatio: '9 / 16' }} />
          ))}
        </div>
      )}

      {!loading && posts.length === 0 && (
        <div className="empty">
          <h3>Пока пусто</h3>
          <div>Загрузите видео — и оно уедет в Reels, TikTok и Shorts одним движением</div>
          <Link to="/posts/new">
            <button className="primary" style={{ marginTop: 18 }}>
              Создать первый пост
            </button>
          </Link>
        </div>
      )}

      {!loading && posts.length > 0 && (
        <div className="grid">
          {posts.map((post) => (
            <PostTile
              key={post.id}
              post={post}
              busy={busyID === post.id}
              onPublish={() => onPublish(post.id)}
              onDelete={() => onDelete(post.id)}
            />
          ))}
        </div>
      )}
    </>
  )
}

interface TileProps {
  post: Post
  busy: boolean
  onPublish: () => void
  onDelete: () => void
}

function PostTile({ post, busy, onPublish, onDelete }: TileProps) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const video = post.media?.[0]

  // Превью оживает при наведении — так лента листается глазами, а не кликами.
  const play = () => videoRef.current?.play().catch(() => undefined)
  const stop = () => {
    const el = videoRef.current
    if (!el) return
    el.pause()
    el.currentTime = 0
  }

  const firstUrl = post.targets.find((t) => t.external_url)?.external_url

  return (
    <div className="tile" onMouseEnter={play} onMouseLeave={stop}>
      <div className="tile-media">
        {video ? (
          <video ref={videoRef} src={video.public_url} muted loop playsInline preload="metadata" />
        ) : post.content ? (
          <div className="tile-text">{post.content}</div>
        ) : (
          <div className="tile-empty">{failureReason(post) || 'без видео'}</div>
        )}

        <div className="tile-scrim" />

        <span className="tile-status">
          <span className={`status-dot ${post.status}`} />
          {STATUS_LABEL[post.status]}
        </span>

        <div className="tile-actions">
          {post.status !== 'published' && (
            <button className="icon-btn" disabled={busy} onClick={onPublish} title="Опубликовать сейчас">
              {busy ? '…' : '↑'}
            </button>
          )}
          <button className="icon-btn" disabled={busy} onClick={onDelete} title="Удалить">
            ✕
          </button>
        </div>

        {video && post.content && (
          <div className="tile-foot">
            <div className="tile-caption">{post.content}</div>
          </div>
        )}
      </div>

      <div className="tile-meta">
        <PlatformMarks targets={post.targets} />
        <span className="when">
          {formatWhen(post)}
          {firstUrl && (
            <a href={firstUrl} target="_blank" rel="noreferrer" title="Открыть публикацию">
              ↗
            </a>
          )}
        </span>
      </div>
    </div>
  )
}

function formatWhen(post: Post): string {
  const value = post.published_at ?? post.scheduled_at ?? post.created_at
  if (!value) return ''

  return new Date(value).toLocaleString('ru-RU', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function failureReason(post: Post): string {
  return post.targets.find((t) => t.error_message)?.error_message ?? ''
}

function plural(n: number): string {
  const mod10 = n % 10
  const mod100 = n % 100

  if (mod10 === 1 && mod100 !== 11) return 'пост'
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return 'поста'
  return 'постов'
}
