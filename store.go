package main

import (
	"errors"
	"sync"
)

// Flag is a single feature flag.
type Flag struct {
	Key            string `json:"key"`
	Enabled        bool   `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent int    `json:"rollout_percent"`
}

// Store is a thread-safe in-memory flag store. Access flags only under mu.
type Store struct {
	mu    sync.RWMutex
	flags map[string]Flag
}

// NewStore returns an initialized, empty Store.
func NewStore() *Store {
	return &Store{
		flags: make(map[string]Flag),
	}
}

// Server holds the shared store and serves the HTTP API.
type Server struct {
	store *Store
}

// NewServer returns a Server backed by the given Store.
func NewServer(store *Store) *Server {
	return &Server{store: store}
}

var (
	ErrFlagNotFound = errors.New("flag not found")
	ErrFlagExists   = errors.New("flag already exists")
)
