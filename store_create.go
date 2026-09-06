package main

// Create stores a new flag. It returns ErrFlagExists if the key is already
// present and the stored flag otherwise.
func (s *Store) Create(f Flag) (Flag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.flags[f.Key]; ok {
		return Flag{}, ErrFlagExists
	}

	s.flags[f.Key] = f
	return f, nil
}
