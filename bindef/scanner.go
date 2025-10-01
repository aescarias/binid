package bindef

type Scanner[T comparable] struct {
	Data     []T
	Position int
}

// NewScanner returns a new Scanner object with data.
func NewScanner[T comparable](data []T) Scanner[T] {
	return Scanner[T]{Data: data, Position: 0}
}

// IsDone reports whether the scanner has consumed all of its data.
func (s *Scanner[T]) IsDone() bool {
	return s.Position >= len(s.Data)
}

// Advance moves the cursor n items forward.
func (s *Scanner[T]) Advance(n int) {
	s.Position += n
}

// Cursor returns the item at the current position.
func (s *Scanner[T]) Cursor() T {
	return s.Data[s.Position]
}

// Peek returns the next n items without advancing the cursor.
func (s *Scanner[T]) Peek(n int) []T {
	return s.Data[s.Position+1 : s.Position+1+n]
}
