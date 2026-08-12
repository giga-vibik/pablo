-- Подключённые аккаунты соцсетей. Источник истины — zernio; таблица служит
-- кэшем, чтобы фронт показывал список без похода во внешний API на каждый рендер.
CREATE TABLE accounts
(
    id                UUID PRIMARY KEY,
    platform          TEXT         NOT NULL CHECK (platform IN ('instagram', 'tiktok', 'youtube', 'threads')),
    zernio_account_id TEXT         NOT NULL,
    username          TEXT         NOT NULL DEFAULT '',
    is_active         BOOLEAN      NOT NULL DEFAULT TRUE,
    synced_at         TIMESTAMP(0) NOT NULL,
    created_at        TIMESTAMP(0) NOT NULL,

    UNIQUE (zernio_account_id)
);

CREATE INDEX idx_accounts_platform ON accounts (platform) WHERE is_active;
