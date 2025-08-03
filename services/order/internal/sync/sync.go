package sync

import "fmt"

const (
	OK   Status = "Success"
	Fail Status = "Fail"
)

type Status string

type Sync struct {
	m map[string]chan Status
}

func New() *Sync {
	return &Sync{
		m: make(map[string]chan Status),
	}
}

func (s *Sync) Push(key string) chan Status {
	s.m[key] = make(chan Status, 1)
	return s.m[key]
}

func (s *Sync) Pull(key string) (chan Status, error) {
	ch, ok := s.m[key]
	if !ok {
		return nil, fmt.Errorf("no channel found for key: %s", key)
	}
	return ch, nil
}

func (s *Sync) Remove(key string) {
	delete(s.m, key)
}
