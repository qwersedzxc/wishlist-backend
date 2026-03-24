package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// StateStore хранит CSRF state в памяти с TTL 10 минут.
// Каждый state одноразовый удаляется сразу после проверки.
type StateStore struct {
	mu    sync.Mutex
	store map[string]time.Time
	ttl   time.Duration
}

// NewStateStore создаёт новый StateStore с TTL 10 минут.
func NewStateStore() *StateStore {
	return &StateStore{
		store: make(map[string]time.Time),
		ttl:   10 * time.Minute,
	}
}

// Generate генерирует случайный state, сохраняет его и возвращает строку.
func (s *StateStore) Generate() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand не должен падать в нормальных условиях ОС
		panic("oauth: StateStore.Generate: " + err.Error())
	}
	state := hex.EncodeToString(b)

	s.mu.Lock()
	s.store[state] = time.Now()
	s.mu.Unlock()

	return state
}

// Validate проверяет state: должен существовать и не быть просроченным.
// После проверки запись удаляется state одноразовый.
// Попутно очищает все устаревшие записи из хранилища.
func (s *StateStore) Validate(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Удаляем все устаревшие записи
	for k, created := range s.store {
		if time.Since(created) >= s.ttl {
			delete(s.store, k)
		}
	}

	created, ok := s.store[state]
	// Удаляем в любом случае state одноразовый
	delete(s.store, state)

	if !ok {
		return false
	}

	return time.Since(created) < s.ttl
}
