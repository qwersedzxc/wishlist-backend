-- +goose Up
ALTER TABLE wishlist_items
ADD COLUMN reserved_by UUID REFERENCES users(id) ON DELETE SET NULL,
ADD COLUMN reserved_at TIMESTAMP,
ADD COLUMN is_incognito_reservation BOOLEAN DEFAULT FALSE;

CREATE INDEX idx_wishlist_items_reserved_by ON wishlist_items(reserved_by);

-- +goose Down
DROP INDEX IF EXISTS idx_wishlist_items_reserved_by;
ALTER TABLE wishlist_items
DROP COLUMN IF EXISTS is_incognito_reservation,
DROP COLUMN IF EXISTS reserved_at,
DROP COLUMN IF EXISTS reserved_by;
