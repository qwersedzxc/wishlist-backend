-- +goose Up
-- +goose StatementBegin

-- Создание таблицы ролей
CREATE TABLE IF NOT EXISTS roles (
    id          SERIAL       PRIMARY KEY,
    name        VARCHAR(50)  NOT NULL UNIQUE,
    description TEXT,
    permissions JSONB        DEFAULT '[]'::jsonb,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMENT ON TABLE roles IS 'Роли пользователей в системе';
COMMENT ON COLUMN roles.id IS 'Уникальный идентификатор роли';
COMMENT ON COLUMN roles.name IS 'Название роли (user, admin, moderator)';
COMMENT ON COLUMN roles.description IS 'Описание роли';
COMMENT ON COLUMN roles.permissions IS 'JSON массив разрешений для роли';
COMMENT ON COLUMN roles.created_at IS 'Дата создания роли';
COMMENT ON COLUMN roles.updated_at IS 'Дата последнего обновления';

-- Создание таблицы связи пользователей и ролей (many-to-many)
CREATE TABLE IF NOT EXISTS user_roles (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id    INTEGER     NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_by UUID        REFERENCES users(id) ON DELETE SET NULL,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    is_active  BOOLEAN     NOT NULL DEFAULT true
);

COMMENT ON TABLE user_roles IS 'Связь пользователей с ролями';
COMMENT ON COLUMN user_roles.id IS 'Уникальный идентификатор связи';
COMMENT ON COLUMN user_roles.user_id IS 'ID пользователя';
COMMENT ON COLUMN user_roles.role_id IS 'ID роли';
COMMENT ON COLUMN user_roles.granted_by IS 'Кто назначил роль';
COMMENT ON COLUMN user_roles.granted_at IS 'Когда назначена роль';
COMMENT ON COLUMN user_roles.expires_at IS 'Когда истекает роль (NULL = бессрочно)';
COMMENT ON COLUMN user_roles.is_active IS 'Активна ли роль';

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_active ON user_roles(user_id, is_active) WHERE is_active = true;
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_unique ON user_roles(user_id, role_id) WHERE is_active = true;

-- Вставка базовых ролей
INSERT INTO roles (name, description, permissions) VALUES 
('user', 'Обычный пользователь', '["read_own_wishlists", "create_wishlist", "edit_own_wishlist", "delete_own_wishlist", "add_friends", "view_friends_wishlists"]'::jsonb),
('admin', 'Администратор', '["*"]'::jsonb);

-- Добавление роли по умолчанию для существующих пользователей
INSERT INTO user_roles (user_id, role_id, granted_by)
SELECT u.id, r.id, NULL
FROM users u
CROSS JOIN roles r
WHERE r.name = 'user'
AND NOT EXISTS (
    SELECT 1 FROM user_roles ur 
    WHERE ur.user_id = u.id AND ur.role_id = r.id
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
-- +goose StatementEnd