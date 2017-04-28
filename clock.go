package cache

import (
	"time"
)

type clock struct {
	instant time.Time
}

func (c *clock) now() time.Time {
	if c == nil {
		return time.Now()
	}
	return c.instant
}
