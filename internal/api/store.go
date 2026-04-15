package api

import (
	"errors"
	"sync"

	"github.com/jamespacheco-dev/pico-api/internal/game"
)

// ErrNotFound is returned when a game ID does not exist in the store.
var ErrNotFound = errors.New("game not found")

// Store is the persistence interface for game sessions.
type Store interface {
	Create(g *game.Game) error
	Get(id string) (*game.Game, error)
	Save(g *game.Game) error
}

// MemoryStore is an in-memory Store implementation for development.
// It stores pointers directly, so mutations to a retrieved Game are reflected
// without calling Save. Save is still required in handler code so that
// switching to a real store later requires no handler changes.
type MemoryStore struct {
	mu    sync.RWMutex
	games map[string]*game.Game
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{games: make(map[string]*game.Game)}
}

func (s *MemoryStore) Create(g *game.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.games[g.ID] = g
	return nil
}

func (s *MemoryStore) Get(id string) (*game.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	g, ok := s.games[id]
	if !ok {
		return nil, ErrNotFound
	}
	return g, nil
}

func (s *MemoryStore) Save(g *game.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.games[g.ID]; !ok {
		return ErrNotFound
	}
	s.games[g.ID] = g
	return nil
}
