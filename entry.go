package cache

import (
	"container/list"
	"time"
)

// An Entry represents a value in a cache
type entry struct {
	data        []byte
	expiration  time.Time
	listElement *list.Element
}

func (e *entry) update(value []byte, expiration time.Time) {
	e.data = value
	e.expiration = expiration
}

func (e *entry) read() []byte {
	return e.data
}
