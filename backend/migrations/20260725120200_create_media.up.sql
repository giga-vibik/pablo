-- Видео поста. zernio забирает медиа по public_url, файл ему не передаётся,
-- поэтому бакет должен отдавать объект публично.
CREATE TABLE media
(
    id          UUID PRIMARY KEY,
    post_id     UUID         NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
    file_name   TEXT         NOT NULL,
    storage_key TEXT         NOT NULL,
    public_url  TEXT         NOT NULL,
    mime_type   TEXT         NOT NULL,
    size_bytes  BIGINT       NOT NULL DEFAULT 0,
    created_at  TIMESTAMP(0) NOT NULL
);

CREATE INDEX idx_media_post_id ON media (post_id);
