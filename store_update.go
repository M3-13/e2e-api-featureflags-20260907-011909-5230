package main

// Update replaces the enabled, description and rollout_percent fields of the
// flag stored under key. The key itself is left unchanged. It returns the
// updated flag, or ErrFlagNotFound if no flag exists under key.
func (s *Store) Update(key string, f Flag) (Flag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.flags[key]
	if !ok {
		return Flag{}, ErrFlagNotFound
	}

	existing.Enabled = f.Enabled
	existing.Description = f.Description
	existing.RolloutPercent = f.RolloutPercent

	s.flags[key] = existing
	return existing, nil
}
