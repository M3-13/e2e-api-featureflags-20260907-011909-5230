package main

// Get returns the flag stored under key. The second return value is false
// when no flag is stored under the given key.
func (s *Store) Get(key string) (Flag, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flag, ok := s.flags[key]
	return flag, ok
}
