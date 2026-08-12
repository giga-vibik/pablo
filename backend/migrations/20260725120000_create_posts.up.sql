-- Пост — единица контента. kind определяет, нужно ли к нему видео:
--   video → instagram reels / tiktok / youtube shorts
--   text  → threads (пишем руками)
CREATE TABLE posts
(
    id           UUID PRIMARY KEY,
    kind         TEXT         NOT NULL CHECK (kind IN ('video', 'text')),
    content      TEXT         NOT NULL DEFAULT '',
    status       TEXT         NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'scheduled', 'publishing', 'published', 'partially_published', 'failed')),
    scheduled_at TIMESTAMP(0),
    published_at TIMESTAMP(0),
    created_at   TIMESTAMP(0) NOT NULL,
    updated_at   TIMESTAMP(0) NOT NULL
);

-- Выборка воркером: посты, которым подошло время публикации.
CREATE INDEX idx_posts_due ON posts (scheduled_at) WHERE status = 'scheduled';

CREATE INDEX idx_posts_created_at ON posts (created_at DESC);
