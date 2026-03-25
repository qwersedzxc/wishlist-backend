-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS articles (
    id           UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    title        VARCHAR(255)    NOT NULL,
    body         TEXT            NOT NULL,
    author_id    UUID            NOT NULL,
    published_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ     NOT NULL DEFAULT now()
);
COMMENT ON TABLE articles IS 'Статьи';
COMMENT ON COLUMN articles.title IS 'Заголовок статьи';
COMMENT ON COLUMN articles.body IS 'Тело статьи';
COMMENT ON COLUMN articles.author_id IS 'Идентификатор автора';
COMMENT ON COLUMN articles.published_at IS 'Дата публикации';
CREATE INDEX IF NOT EXISTS idx_articles_author_id ON articles (author_id);
CREATE INDEX IF NOT EXISTS idx_articles_created_at ON articles (created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS articles;
-- +goose StatementEnd

