import PlatformIcon from './PlatformIcon'
import { PLATFORM_COLORS, PLATFORM_LABELS, type Target } from '../types'

interface Props {
  targets: Target[]
}

const STATUS_TEXT: Record<Target['status'], string> = {
  pending: 'в очереди',
  publishing: 'публикуется',
  published: 'опубликован',
  failed: 'ошибка',
}

// Один кружок на площадку: цвет площадки заливается только после публикации,
// поэтому статус поста читается без легенды — по количеству «залитых» значков.
export default function PlatformMarks({ targets }: Props) {
  return (
    <div className="platform-marks">
      {targets.map((target) => (
        <span
          key={target.id}
          className={`mark ${target.status}`}
          style={{ ['--mark-color' as string]: PLATFORM_COLORS[target.platform] }}
          title={
            `${PLATFORM_LABELS[target.platform]} — ${STATUS_TEXT[target.status]}` +
            (target.error_message ? `: ${target.error_message}` : '')
          }
        >
          <PlatformIcon platform={target.platform} />
        </span>
      ))}
    </div>
  )
}
