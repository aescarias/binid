package bindef

type Scanner[T any] struct {
	Data    []T
	Current int
}

func (s *Scanner[T]) IsDone() bool {
	return s.Current >= len(s.Data)
}

func (s *Scanner[T]) Advance(n int) {
	s.Current += n
}

func (s *Scanner[T]) Cursor() T {
	return s.Data[s.Current]
}

func (s *Scanner[T]) Peek(n int) []T {
	return s.Data[s.Current+1 : s.Current+1+n]
}
