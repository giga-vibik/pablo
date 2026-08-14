export type Platform = 'instagram' | 'tiktok' | 'youtube' | 'threads'

export type PostKind = 'video' | 'text'

export type PostStatus =
  | 'draft'
  | 'scheduled'
  | 'publishing'
  | 'published'
  | 'partially_published'
  | 'failed'

export type TargetStatus = 'pending' | 'publishing' | 'published' | 'failed'

export interface Target {
  id: string
  platform: Platform
  caption?: string
  status: TargetStatus
  external_post_id?: string
  external_url?: string
  error_message?: string
}

export interface Media {
  id: string
  file_name: string
  public_url: string
  mime_type?: string
  size_bytes?: number
}

export interface Post {
  id: string
  kind: PostKind
  content?: string
  status: PostStatus
  scheduled_at?: string
  published_at?: string
  created_at?: string
  targets: Target[]
  media?: Media[]
}

export interface Account {
  id: string
  platform: Platform
  username?: string
  is_active: boolean
  synced_at?: string
}

// Threads — единственная текстовая площадка, остальные требуют видео.
export const VIDEO_PLATFORMS: Platform[] = ['instagram', 'tiktok', 'youtube']

export const PLATFORM_LABELS: Record<Platform, string> = {
  instagram: 'Instagram Reels',
  tiktok: 'TikTok',
  youtube: 'YouTube Shorts',
  threads: 'Threads',
}

export const PLATFORM_SHORT: Record<Platform, string> = {
  instagram: 'Reels',
  tiktok: 'TikTok',
  youtube: 'Shorts',
  threads: 'Threads',
}

export const PLATFORM_COLORS: Record<Platform, string> = {
  instagram: '#d62976',
  tiktok: '#12100e',
  youtube: '#ff0033',
  threads: '#16150f',
}

export type StatsState = 'ready' | 'pending' | 'unavailable'

export interface Metrics {
  impressions?: number
  reach?: number
  views?: number
  likes?: number
  comments?: number
  shares?: number
  saves?: number
  clicks?: number
  engagement_rate?: number
}

export interface PlatformStats {
  platform: Platform
  state: StatsState
  message?: string
  username?: string
  url?: string
  last_updated?: string
  metrics?: Metrics
}

export interface PostStats {
  post_id: string
  totals?: Metrics
  platforms: PlatformStats[]
}

// Показываем в этом порядке. Площадки отдают разный набор: TikTok не сообщает
// reach, YouTube — impressions, поэтому пустые метрики скрываем, а не рисуем нулями.
export const METRIC_LABELS: [keyof Metrics, string][] = [
  ['views', 'Просмотры'],
  ['impressions', 'Показы'],
  ['reach', 'Охват'],
  ['likes', 'Лайки'],
  ['comments', 'Комментарии'],
  ['shares', 'Репосты'],
  ['saves', 'Сохранения'],
  ['clicks', 'Клики'],
]

// Лимиты подписи у площадок. Считаем, чтобы не отправлять заведомо обрезанное.
export const CAPTION_LIMITS: Record<Platform, number> = {
  instagram: 2200,
  tiktok: 2200,
  youtube: 5000,
  threads: 500,
}
