import { useEffect, useState } from 'react'

import PlatformIcon from './PlatformIcon'
import { getPostStats } from '../api'
import {
  METRIC_LABELS,
  PLATFORM_COLORS,
  PLATFORM_LABELS,
  type Metrics,
  type PlatformStats,
  type PostStats,
} from '../types'

interface Props {
  postId: string
  onClose: () => void
}

export default function StatsModal({ postId, onClose }: Props) {
  const [stats, setStats] = useState<PostStats | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let alive = true

    getPostStats(postId)
      .then((res) => alive && setStats(res))
      .catch((err) => alive && setError(err instanceof Error ? err.message : 'Не удалось загрузить'))
      .finally(() => alive && setLoading(false))

    return () => {
      alive = false
    }
  }, [postId])

  // Esc закрывает — поп-ап без этого ощущается ловушкой.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose()
    window.addEventListener('keydown', onKey)

    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  const hasTotals = stats && countFilled(stats.totals) > 0

  return (
    <div className="overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-head">
          <h3>Статистика</h3>
          <button className="ghost tiny" onClick={onClose}>
            Закрыть
          </button>
        </div>

        <div className="modal-body">
          {loading && <div className="muted">Спрашиваем zernio…</div>}
          {error && <div className="banner">{error}</div>}

          {stats && !error && (
            <>
              {hasTotals && (
                <div className="totals">
                  <div className="totals-head">Всего по площадкам</div>
                  <MetricGrid metrics={stats.totals} big />
                  {typeof stats.totals?.engagement_rate === 'number' &&
                    stats.totals.engagement_rate > 0 && (
                      <div className="hint">
                        Вовлечённость {stats.totals.engagement_rate.toFixed(2)}%
                      </div>
                    )}
                </div>
              )}

              {stats.platforms.map((p) => (
                <PlatformBlock key={p.platform} stats={p} />
              ))}

              {stats.platforms.length === 0 && (
                <div className="muted">У поста нет площадок</div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  )
}

function PlatformBlock({ stats }: { stats: PlatformStats }) {
  const filled = countFilled(stats.metrics)

  return (
    <div className="stat-block">
      <div className="stat-head">
        <span
          className={`mark ${stats.state === 'ready' ? 'published' : ''}`}
          style={{ ['--mark-color' as string]: PLATFORM_COLORS[stats.platform] }}
        >
          <PlatformIcon platform={stats.platform} />
        </span>

        <div style={{ flex: 1, minWidth: 0 }}>
          <div className="stat-name">{PLATFORM_LABELS[stats.platform]}</div>
          {stats.username && <div className="stat-sub">{stats.username}</div>}
        </div>

        {stats.url && (
          <a href={stats.url} target="_blank" rel="noreferrer" className="stat-link">
            открыть ↗
          </a>
        )}
      </div>

      {stats.state === 'ready' && filled > 0 && <MetricGrid metrics={stats.metrics} />}

      {/* Площадка отдала все нули — не ошибка: пост только опубликован. */}
      {stats.state === 'ready' && filled === 0 && (
        <div className="stat-note">Площадка пока сообщает нули</div>
      )}

      {stats.state !== 'ready' && (
        <div className={`stat-note ${stats.state}`}>
          {stats.message || (stats.state === 'pending' ? 'данные ещё собираются' : 'нет данных')}
        </div>
      )}

      {stats.state === 'ready' &&
        typeof stats.metrics?.engagement_rate === 'number' &&
        stats.metrics.engagement_rate > 0 && (
          <div className="stat-note">
            Вовлечённость {stats.metrics.engagement_rate.toFixed(2)}%
          </div>
        )}
    </div>
  )
}

function MetricGrid({ metrics, big }: { metrics?: Metrics; big?: boolean }) {
  if (!metrics) return null

  // Показываем только непустые: у каждой площадки свой набор, и нули там,
  // где метрики просто не существует, читаются как провал.
  const shown = METRIC_LABELS.filter(([key]) => {
    const value = metrics[key]
    return typeof value === 'number' && value > 0
  })

  if (shown.length === 0) return null

  return (
    <div className={`metrics ${big ? 'big' : ''}`}>
      {shown.map(([key, label]) => (
        <div className="metric" key={key}>
          <div className="metric-value">{formatNumber(metrics[key] as number)}</div>
          <div className="metric-label">{label}</div>
        </div>
      ))}
    </div>
  )
}

function countFilled(metrics?: Metrics): number {
  if (!metrics) return 0

  return METRIC_LABELS.filter(([key]) => {
    const value = metrics[key]
    return typeof value === 'number' && value > 0
  }).length
}

function formatNumber(value: number): string {
  if (value >= 1_000_000) return (value / 1_000_000).toFixed(1).replace('.0', '') + 'M'
  if (value >= 10_000) return Math.round(value / 1000) + 'K'

  return Math.round(value).toLocaleString('ru-RU')
}
