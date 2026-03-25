-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS friendships (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT unique_friendship UNIQUE (user_id, friend_id),
    CONSTRAINT no_self_friendship CHECK (user_id != friend_id)
);

COMMENT ON TABLE friendships IS 'Дружеские связи между пользователями';
COMMENT ON COLUMN friendships.id IS 'Уникальный идентификатор связи';
COMMENT ON COLUMN friendships.user_id IS 'ID пользователя, отправившего запрос';
COMMENT ON COLUMN friendships.friend_id IS 'ID пользователя, получившего запрос';
COMMENT ON COLUMN friendships.status IS 'Статус дружбы (pending, accepted, rejected)';

CREATE INDEX IF NOT EXISTS idx_friendships_user_id ON friendships(user_id);
CREATE INDEX IF NOT EXISTS idx_friendships_friend_id ON friendships(friend_id);
CREATE INDEX IF NOT EXISTS idx_friendships_status ON friendships(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS friendships;
-- +goose StatementEnd
