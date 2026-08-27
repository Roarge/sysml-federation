package projection

import "context"

// Subscribe registers a channel that receives every new version until ctx
// ends. The channel holds one value: a subscriber that falls behind sees the
// latest version and refetches once, which is all a version event is for
// (AD-0014).
func (s *Store) Subscribe(ctx context.Context) <-chan int {
	ch := make(chan int, 1)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	go func() {
		<-ctx.Done()
		s.mu.Lock()
		delete(s.subs, ch)
		close(ch)
		s.mu.Unlock()
	}()
	return ch
}

// Subscribers counts the open subscriptions.
func (s *Store) Subscribers() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.subs)
}

// notify replaces whatever a subscriber has not read yet with the version.
// The caller holds the lock, so no channel is closed underneath it.
func (s *Store) notify(version int) {
	for ch := range s.subs {
		select {
		case <-ch:
		default:
		}
		ch <- version
	}
}
