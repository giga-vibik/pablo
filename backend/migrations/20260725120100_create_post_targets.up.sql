-- Таргет — одна площадка внутри поста. Статус ведётся отдельно: публикация в
-- TikTok может пройти, а в Instagram упасть, и пост станет partially_published.
CREATE TABLE post_targets
(
    id               UUID PRIMARY KEY,
    post_id          UUID         NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
    platform         TEXT         NOT NULL CHECK (platform IN ('instagram', 'tiktok', 'youtube', 'threads')),
    caption          TEXT         NOT NULL DEFAULT '',
    status           TEXT         NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'publishing', 'published', 'failed')),
    external_post_id TEXT         NOT NULL DEFAULT '',
    external_url     TEXT         NOT NULL DEFAULT '',
    error_message    TEXT         NOT NULL DEFAULT '',
    published_at     TIMESTAMP(0),
    created_at       TIMESTAMP(0) NOT NULL,
    updated_at       TIMESTAMP(0) NOT NULL,

    -- Одна площадка встречается в посте один раз.
    UNIQUE (post_id, platform)
);

CREATE INDEX idx_post_targets_post_id ON post_targets (post_id);
