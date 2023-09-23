package session

import (
	"strconv"
	"time"
)

type Session struct {
	id string
}

func (s *Session) ID() string {
	return s.id
}

func NewSession() *Session {
	return &Session{
		id: strconv.FormatInt(time.Now().UnixNano(), 10),
	}
}
