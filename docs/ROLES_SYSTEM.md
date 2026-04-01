# Система ролей и разрешений

## Обзор

Система ролей позволяет управлять доступом пользователей к различным функциям приложения. Реализована с использованием RBAC (Role-Based Access Control).

## Структура базы данных

### Таблица `roles`
```sql
CREATE TABLE roles (
    id          SERIAL       PRIMARY KEY,
    name        VARCHAR(50)  NOT NULL UNIQUE,
    description TEXT,
    permissions JSONB        DEFAULT '[]'::jsonb,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);
```

### Таблица `user_roles`
```sql
CREATE TABLE user_roles (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id    INTEGER     NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_by UUID        REFERENCES users(id) ON DELETE SET NULL,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    is_active  BOOLEAN     NOT NULL DEFAULT true
);
```

## Базовые роли

### 1. User (Пользователь)
**Разрешения:**
- `read_own_wishlists` - просмотр своих вишлистов
- `create_wishlist` - создание вишлистов
- `edit_own_wishlist` - редактирование своих вишлистов
- `delete_own_wishlist` - удаление своих вишлистов
- `add_friends` - добавление друзей
- `view_friends_wishlists` - просмотр вишлистов друзей

### 2. Admin (Администратор)
**Разрешения:**
- `*` - все разрешения (полный доступ)

## API Endpoints

### Получение ролей
```http
GET /api/v1/roles/my
Authorization: Bearer <token>
```
Возвращает роли текущего пользователя.

```http
GET /api/v1/roles/user/{userId}
Authorization: Bearer <token>
```
Возвращает роли указанного пользователя.

### Управление ролями (только админы)
```http
GET /api/v1/roles/
Authorization: Bearer <token>
```
Получить все роли.

```http
POST /api/v1/roles/
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "editor",
  "description": "Редактор контента",
  "permissions": ["edit_content", "moderate_comments"]
}
```
Создать новую роль.

```http
POST /api/v1/roles/assign
Authorization: Bearer <token>
Content-Type: application/json

{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "role_id": 2
}
```
Назначить роль пользователю.

```http
POST /api/v1/roles/remove
Authorization: Bearer <token>
Content-Type: application/json

{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "role_id": 2
}
```
Удалить роль у пользователя.

## Использование в коде

### Middleware для проверки ролей

```go
// Требует определенную роль
r.With(roleMiddleware.RequireRole("admin")).Get("/admin", handler)

// Требует определенное разрешение
r.With(roleMiddleware.RequirePermission("delete_content")).Delete("/content/{id}", handler)

// Требует роль администратора
r.With(roleMiddleware.RequireAdmin()).Get("/admin", handler)

// Загружает роли в контекст (не блокирует)
r.With(roleMiddleware.LoadUserRoles()).Get("/profile", handler)
```

### Проверка разрешений в хендлерах

```go
func (h *Handler) SomeHandler(w http.ResponseWriter, r *http.Request) {
    userWithRoles := middleware.GetUserWithRoles(r)
    if userWithRoles == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Проверка роли
    if !userWithRoles.HasRole("admin") {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Проверка разрешения
    if !userWithRoles.HasPermission("delete_content") {
        http.Error(w, "Insufficient permissions", http.StatusForbidden)
        return
    }

    // Проверка на админа
    if userWithRoles.IsAdmin() {
        // Логика для админа
    }
}
```

### Работа с репозиторием ролей

```go
// Получить пользователя с ролями
userWithRoles, err := roleRepo.GetUserWithRoles(ctx, userID)

// Назначить роль
err := roleRepo.AssignRoleToUser(ctx, userID, roleID, &adminID)

// Удалить роль
err := roleRepo.RemoveRoleFromUser(ctx, userID, roleID)

// Получить всех пользователей с ролью
admins, err := roleRepo.GetUsersWithRole(ctx, "admin")
```

## Инициализация в приложении

```go
// В main.go или app.go
roleRepo := role.New(db)

// В роутере
router := v1.NewRouter(
    wishlistUC,
    authUC,
    friendshipUC,
    roleRepo, // Добавить репозиторий ролей
    provider,
    providerName,
    s3cfg,
    emailService,
    log,
)
```

## Миграция

Для применения системы ролей выполните миграцию:

```bash
# Применить миграцию
goose -dir migrations postgres "your_connection_string" up

# Или через CLI приложения
./app migrate up
```

## Примеры использования

### 1. Защита админских роутов
```go
r.Route("/admin", func(r chi.Router) {
    r.Use(roleMiddleware.RequireAdmin())
    r.Get("/users", adminHandler.GetUsers)
    r.Delete("/users/{id}", adminHandler.DeleteUser)
})
```

### 2. Условная логика по ролям
```go
func (h *WishlistHandler) GetWishlists(w http.ResponseWriter, r *http.Request) {
    userWithRoles := middleware.GetUserWithRoles(r)
    
    var wishlists []entity.Wishlist
    var err error
    
    if userWithRoles != nil && userWithRoles.IsAdmin() {
        // Админы видят все вишлисты
        wishlists, err = h.repo.GetAllWishlists(r.Context())
    } else {
        // Обычные пользователи видят только публичные
        wishlists, err = h.repo.GetPublicWishlists(r.Context())
    }
    
    // ... обработка результата
}
```

### 3. Создание кастомной роли
```go
customRole := &entity.Role{
    Name:        "content_manager",
    Description: stringPtr("Менеджер контента"),
    Permissions: entity.Permissions{
        "create_content",
        "edit_content",
        "delete_own_content",
        "moderate_comments",
    },
}

err := roleRepo.CreateRole(ctx, customRole)
```

## Безопасность

1. **Принцип минимальных привилегий** - назначайте только необходимые разрешения
2. **Регулярный аудит** - проверяйте назначенные роли
3. **Временные роли** - используйте `expires_at` для временных разрешений
4. **Логирование** - ведите журнал изменений ролей

## Расширение системы

Для добавления новых разрешений:

1. Добавьте константы разрешений
2. Обновите базовые роли в миграции
3. Добавьте проверки в middleware или хендлеры
4. Обновите документацию

Система ролей легко расширяется и позволяет гибко управлять доступом в приложении.