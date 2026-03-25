-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wishlists (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID         NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    is_public   BOOLEAN      NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMENT ON TABLE wishlists IS 'Списки желаний пользователей';
COMMENT ON COLUMN wishlists.id IS 'Уникальный идентификатор вишлиста';
COMMENT ON COLUMN wishlists.user_id IS 'ID пользователя-владельца';
COMMENT ON COLUMN wishlists.title IS 'Название вишлиста';
COMMENT ON COLUMN wishlists.description IS 'Описание вишлиста';
COMMENT ON COLUMN wishlists.is_public IS 'Публичный ли вишлист';
COMMENT ON COLUMN wishlists.created_at IS 'Дата создания';
COMMENT ON COLUMN wishlists.updated_at IS 'Дата последнего обновления';

CREATE INDEX IF NOT EXISTS idx_wishlists_user_id ON wishlists(user_id);
CREATE INDEX IF NOT EXISTS idx_wishlists_is_public ON wishlists(is_public);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wishlists;
-- +goose StatementEnd
