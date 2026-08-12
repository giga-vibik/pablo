import type { Platform } from '../types'

interface Props {
  platform: Platform
}

// Упрощённые глифы площадок: узнаваемая форма без попытки повторить логотип.
export default function PlatformIcon({ platform }: Props) {
  switch (platform) {
    case 'instagram':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <rect x="3" y="3" width="18" height="18" rx="5.5" />
          <circle cx="12" cy="12" r="4" />
          <circle cx="17.4" cy="6.6" r="1.1" fill="currentColor" stroke="none" />
        </svg>
      )

    case 'tiktok':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M14 3v11.2a3.8 3.8 0 1 1-3.2-3.75" />
          <path d="M14 3c.4 2.6 2 4.2 4.6 4.5" />
        </svg>
      )

    case 'youtube':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinejoin="round">
          <rect x="2.5" y="5" width="19" height="14" rx="4.5" />
          <path d="M10.4 9.3 15 12l-4.6 2.7z" fill="currentColor" stroke="none" />
        </svg>
      )

    case 'threads':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M16.8 12.4c-.15-3.1-2-4.6-4.6-4.6-2 0-3.4.9-4 2.2" />
          <path d="M8.6 15.4c1 1 3.4 1.3 4.7.1 1-1 1-3-1.4-3.2-2-.2-3 .7-3 1.7 0 1.1 1 1.7 2 1.6 1.8-.2 2.7-1.7 2.9-4" />
          <path d="M12 21c-5 0-8-3.2-8-9s3-9 8-9 8 3.2 8 9-3 9-8 9z" />
        </svg>
      )
  }
}
