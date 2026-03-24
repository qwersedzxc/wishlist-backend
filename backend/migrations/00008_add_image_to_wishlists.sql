-- +goose Up
-- +goose StatementBegin
ALTER TABLE wishlists 
    ADD COLUMN image_url TEXT;

COMMENT ON COLUMN wishlists.image_url IS 'URL изображения вишлиста';

CREATE INDEX IF NOT EXISTS idx_wishlists_image_url ON wishlists(image_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE wishlists 
    DROP COLUMN IF EXISTS image_url;
-- +goose StatementEnd