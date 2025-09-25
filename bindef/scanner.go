package bindef

type Scanner struct {
	Data    string
	Current int
}

func (s *Scanner) IsDone() bool {
	return s.Current >= len(s.Data)
}

func (s *Scanner) Advance(n int) {
	s.Current += n
}

func (s *Scanner) Cursor() byte {
	return s.Data[s.Current]
}

func (s *Scanner) Peek(n int) string {
	return s.Data[s.Current+1 : s.Current+1+n]
}
