import { useEffect, useState } from 'react'

import PlatformIcon from '../components/PlatformIcon'
import { getConnectURL, listAccounts, syncAccounts } from '../api'
import { PLATFORM_COLORS, PLATFORM_LABELS, type Account, type Platform } from '../types'

const ALL_PLATFORMS: Platform[] = ['instagram', 'tiktok', 'youtube', 'threads']

export default function AccountsPage() {
  const [accounts, setAccounts] = useState<Account[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)

  const load = async () => {
    setError('')
    try {
      const res = await listAccounts()
      setAccounts(res.accounts ?? [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось загрузить аккаунты')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const onSync = async () => {
    setBusy(true)
    setError('')
    try {
      const res = await syncAccounts()
      setAccounts(res.accounts ?? [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось синхронизировать')
    } finally {
      setBusy(false)
    }
  }

  const onConnect = async (platform: Platform) => {
    setBusy(true)
    setError('')
    try {
      const res = await getConnectURL(platform, window.location.origin + '/accounts')
      // Авторизация идёт на стороне zernio, возвращаемся сюда по redirect.
      window.location.href = res.auth_url
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось получить ссылку')
      setBusy(false)
    }
  }

  const byPlatform = new Map(accounts.map((a) => [a.platform, a]))
  const connected = accounts.filter((a) => a.is_active).length

  return (
    <>
      <div className="page-header">
        <div>
          <h1 className="title">Аккаунты</h1>
          <p className="subtitle">
            {loading ? 'Загружаем…' : `${connected} из ${ALL_PLATFORMS.length} подключено`}
          </p>
        </div>

        <button onClick={onSync} disabled={busy}>
          {busy ? 'Обновляем…' : 'Синхронизировать'}
        </button>
      </div>

      {error && <div className="banner">{error}</div>}

      <div className="accounts">
        {ALL_PLATFORMS.map((platform) => {
          const account = byPlatform.get(platform)
          const active = Boolean(account?.is_active)

          return (
            <div
              className="account-card"
              key={platform}
              style={{ ['--brand' as string]: PLATFORM_COLORS[platform] }}
            >
              <div className="glyph">
                <PlatformIcon platform={platform} />
              </div>

              <h3>{PLATFORM_LABELS[platform]}</h3>
              <div className="who">
                {account ? (
                  <>
                    <span className={`status-dot ${active ? 'published' : 'failed'}`} />{' '}
                    {account.username || 'без имени'}
                  </>
                ) : (
                  'не подключён'
                )}
              </div>

              <button onClick={() => onConnect(platform)} disabled={busy}>
                {active ? 'Переподключить' : 'Подключить'}
              </button>
            </div>
          )
        })}
      </div>
    </>
  )
}
