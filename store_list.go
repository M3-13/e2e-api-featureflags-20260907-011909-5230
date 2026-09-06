package main

// List returns a copy of all stored flags as a slice. It returns an empty
// slice (never nil) when the store is empty.
func (s *Store) List() []Flag {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flags := make([]Flag, 0, len(s.flags))
	for _, f := range s.flags {
		flags = append(flags, f)
	}
	return flags
}
