package main

// Delete removes the flag with the given key from the store. It returns
// ErrFlagNotFound if no flag with that key exists.
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.flags[key]; !ok {
		return ErrFlagNotFound
	}
	delete(s.flags, key)
	return nil
}
