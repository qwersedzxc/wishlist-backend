-- +goose Up
-- +goose StatementBegin
ALTER TABLE wishlists 
    ADD COLUMN event_name VARCHAR(255),
    ADD COLUMN event_date DATE,
    ADD COLUMN privacy_level VARCHAR(20) NOT NULL DEFAULT 'friends_only' CHECK (privacy_level IN ('public', 'friends_only', 'link_only')),
    ADD COLUMN share_token VARCHAR(64) UNIQUE;

COMMENT ON COLUMN wishlists.event_name IS 'Название события/праздника';
COMMENT ON COLUMN wishlists.event_date IS 'Дата события';
COMMENT ON COLUMN wishlists.privacy_level IS 'Уровень приватности (public, friends_only, link_only)';
COMMENT ON COLUMN wishlists.share_token IS 'Уникальный токен для доступа по ссылке';

-- Обновляем существующие записи
UPDATE wishlists SET privacy_level = CASE WHEN is_public THEN 'public' ELSE 'friends_only' END;

CREATE INDEX IF NOT EXISTS idx_wishlists_privacy_level ON wishlists(privacy_level);
CREATE INDEX IF NOT EXISTS idx_wishlists_share_token ON wishlists(share_token);
CREATE INDEX IF NOT EXISTS idx_wishlists_event_date ON wishlists(event_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE wishlists 
    DROP COLUMN IF EXISTS event_name,
    DROP COLUMN IF EXISTS event_date,
    DROP COLUMN IF EXISTS privacy_level,
    DROP COLUMN IF EXISTS share_token;
-- +goose StatementEnd
