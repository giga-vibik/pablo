import { useState } from 'react'

import { login } from '../api'

interface Props {
  onLogin: () => void
}

export default function LoginPage({ onLogin }: Props) {
  const [loginName, setLoginName] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      await login(loginName, password)
      onLogin()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось войти')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-screen">
      <form className="login-card" onSubmit={submit}>
        <div className="wordmark">
          Pablo
          <span className="dot" />
        </div>
        <div className="tagline">Reels, TikTok, Shorts и Threads из одного места</div>

        {error && <div className="banner">{error}</div>}

        <div className="field">
          <label className="field-label">Логин</label>
          <input value={loginName} onChange={(e) => setLoginName(e.target.value)} autoFocus />
        </div>

        <div className="field">
          <label className="field-label">Пароль</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>

        <button className="primary" type="submit" disabled={loading}>
          {loading ? 'Входим…' : 'Войти'}
        </button>
      </form>
    </div>
  )
}
