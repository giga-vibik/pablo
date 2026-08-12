import { useState } from 'react'
import { NavLink, Navigate, Route, Routes, useNavigate } from 'react-router-dom'

import { clearToken, getToken } from './api'
import AccountsPage from './pages/AccountsPage'
import LoginPage from './pages/LoginPage'
import NewPostPage from './pages/NewPostPage'
import PostsPage from './pages/PostsPage'

export default function App() {
  const [authorized, setAuthorized] = useState(Boolean(getToken()))
  const navigate = useNavigate()

  if (!authorized) {
    return <LoginPage onLogin={() => setAuthorized(true)} />
  }

  const logout = () => {
    clearToken()
    setAuthorized(false)
    navigate('/')
  }

  return (
    <div className="shell">
      <header className="topbar">
        <NavLink to="/posts" className="wordmark">
          Pablo
          <span className="dot" />
        </NavLink>

        <nav className="topnav">
          <NavLink to="/posts">Лента</NavLink>
          <NavLink to="/accounts">Аккаунты</NavLink>
        </nav>

        <button className="ghost tiny" onClick={logout}>
          Выйти
        </button>
      </header>

      <div className="page">
        <Routes>
          <Route path="/" element={<Navigate to="/posts" replace />} />
          <Route path="/posts" element={<PostsPage />} />
          <Route path="/posts/new" element={<NewPostPage />} />
          <Route path="/accounts" element={<AccountsPage />} />
          <Route path="*" element={<Navigate to="/posts" replace />} />
        </Routes>
      </div>
    </div>
  )
}
