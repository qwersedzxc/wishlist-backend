-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wishlist_items (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    wishlist_id  UUID         NOT NULL REFERENCES wishlists(id) ON DELETE CASCADE,
    title        VARCHAR(255) NOT NULL,
    description  TEXT,
    url          TEXT,
    price        DECIMAL(10, 2),
    priority     INTEGER      NOT NULL DEFAULT 0 CHECK (priority >= 0 AND priority <= 10),
    is_purchased BOOLEAN      NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMENT ON TABLE wishlist_items IS 'Элементы списков желаний';
COMMENT ON COLUMN wishlist_items.id IS 'Уникальный идентификатор элемента';
COMMENT ON COLUMN wishlist_items.wishlist_id IS 'ID вишлиста';
COMMENT ON COLUMN wishlist_items.title IS 'Название желаемого предмета';
COMMENT ON COLUMN wishlist_items.description IS 'Описание предмета';
COMMENT ON COLUMN wishlist_items.url IS 'Ссылка на предмет';
COMMENT ON COLUMN wishlist_items.price IS 'Цена предмета';
COMMENT ON COLUMN wishlist_items.priority IS 'Приоритет (0-10, где 10 - самый высокий)';
COMMENT ON COLUMN wishlist_items.is_purchased IS 'Куплен ли предмет';
COMMENT ON COLUMN wishlist_items.created_at IS 'Дата создания';
COMMENT ON COLUMN wishlist_items.updated_at IS 'Дата последнего обновления';

CREATE INDEX IF NOT EXISTS idx_wishlist_items_wishlist_id ON wishlist_items(wishlist_id);
CREATE INDEX IF NOT EXISTS idx_wishlist_items_is_purchased ON wishlist_items(is_purchased);
CREATE INDEX IF NOT EXISTS idx_wishlist_items_priority ON wishlist_items(priority);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wishlist_items;
-- +goose StatementEnd
