package main

import "hash/fnv"

// Evaluate deterministically decides whether the flag identified by key is
// active for the given user. The decision is a stable function of the flag's
// rollout_percent and a FNV-1a hash of key + ":" + user, so the same key and
// user always produce the same result with no randomness and no map iteration.
func (s *Store) Evaluate(key, user string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flag, ok := s.flags[key]
	if !ok {
		return false, ErrFlagNotFound
	}

	h := fnv.New32a()
	h.Write([]byte(key + ":" + user))
	return int(h.Sum32()%100) < flag.RolloutPercent, nil
}
