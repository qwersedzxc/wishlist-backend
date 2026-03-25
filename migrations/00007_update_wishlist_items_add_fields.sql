-- +goose Up
-- +goose StatementBegin
ALTER TABLE wishlist_items 
    ADD COLUMN image_url TEXT,
    ADD COLUMN category VARCHAR(100);

COMMENT ON COLUMN wishlist_items.image_url IS 'URL загруженного изображения предмета';
COMMENT ON COLUMN wishlist_items.category IS 'Категория предмета';

CREATE INDEX IF NOT EXISTS idx_wishlist_items_category ON wishlist_items(category);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE wishlist_items 
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS category;
-- +goose StatementEnd
