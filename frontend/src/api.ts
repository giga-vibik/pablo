import type { Account, Media, Platform, Post, PostKind, PostStats } from './types'

const TOKEN_KEY = 'pablo_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)

  const token = getToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const res = await fetch(path, { ...init, headers })

  if (res.status === 401) {
    clearToken()
    throw new ApiError(401, 'Сессия истекла, войдите заново')
  }

  if (!res.ok) {
    let message = `Ошибка ${res.status}`
    try {
      const body = await res.json()
      if (body?.message) message = body.message
    } catch {
      // тело может быть пустым — оставляем дефолтное сообщение
    }
    throw new ApiError(res.status, message)
  }

  if (res.status === 204 || res.headers.get('Content-Length') === '0') {
    return undefined as T
  }

  const text = await res.text()
  return text ? (JSON.parse(text) as T) : (undefined as T)
}

function json(body: unknown): RequestInit {
  return {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  }
}

export async function login(loginName: string, password: string): Promise<string> {
  const res = await request<{ access_token: string }>(
    '/v1/login',
    json({ login: loginName, password }),
  )

  setToken(res.access_token)
  return res.access_token
}

export function listPosts(limit = 50, offset = 0): Promise<{ posts: Post[] }> {
  return request(`/v1/posts?limit=${limit}&offset=${offset}`)
}

export function getPost(postId: string): Promise<Post> {
  return request(`/v1/posts/${postId}`)
}

export interface CreatePostBody {
  kind: PostKind
  content?: string
  scheduled_at?: string
  targets: { platform: Platform; caption?: string }[]
}

export function createPost(body: CreatePostBody): Promise<Post> {
  return request('/v1/posts', json(body))
}

export function deletePost(postId: string): Promise<void> {
  return request(`/v1/posts/${postId}`, { method: 'DELETE' })
}

export function publishPost(postId: string): Promise<Post> {
  return request(`/v1/posts/${postId}/publish`, { method: 'POST' })
}

export function getPostStats(postId: string): Promise<PostStats> {
  return request(`/v1/posts/${postId}/stats`)
}

export function uploadVideo(postId: string, file: File): Promise<Media> {
  const form = new FormData()
  form.append('file', file)

  // Content-Type не ставим руками — boundary проставит браузер.
  return request(`/v1/posts/${postId}/media`, { method: 'POST', body: form })
}

export function listAccounts(): Promise<{ accounts: Account[] }> {
  return request('/v1/accounts')
}

export function syncAccounts(): Promise<{ accounts: Account[] }> {
  return request('/v1/accounts/sync', { method: 'POST' })
}

export function getConnectURL(platform: Platform, redirectURL: string): Promise<{ auth_url: string }> {
  const params = new URLSearchParams({ platform, redirect_url: redirectURL })
  return request(`/v1/accounts/connect_url?${params.toString()}`)
}
