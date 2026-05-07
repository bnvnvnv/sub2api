CREATE TABLE IF NOT EXISTS openai_web_threads (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,

    local_thread_id VARCHAR(64) NOT NULL UNIQUE,
    page_session_id VARCHAR(64) NOT NULL UNIQUE,
    upstream_conversation_id VARCHAR(128),
    upstream_session_id VARCHAR(128),

    provider VARCHAR(32) NOT NULL DEFAULT 'openai_web',
    title VARCHAR(255) NOT NULL DEFAULT '',
    requested_model VARCHAR(100) NOT NULL DEFAULT '',
    capability_mode VARCHAR(32) NOT NULL DEFAULT 'web_chat',
    history_mode VARCHAR(32) NOT NULL DEFAULT 'upstream_only',
    cache_policy VARCHAR(32) NOT NULL DEFAULT 'local_only',
    sync_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    status VARCHAR(32) NOT NULL DEFAULT 'active',

    last_synced_at TIMESTAMPTZ,
    last_error TEXT NOT NULL DEFAULT '',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_openai_web_threads_user_id
    ON openai_web_threads(user_id);

CREATE INDEX IF NOT EXISTS idx_openai_web_threads_group_id
    ON openai_web_threads(group_id);

CREATE INDEX IF NOT EXISTS idx_openai_web_threads_account_id
    ON openai_web_threads(account_id);

CREATE INDEX IF NOT EXISTS idx_openai_web_threads_user_updated
    ON openai_web_threads(user_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_openai_web_threads_upstream_conversation_id
    ON openai_web_threads(upstream_conversation_id);

CREATE INDEX IF NOT EXISTS idx_openai_web_threads_deleted_at
    ON openai_web_threads(deleted_at);
