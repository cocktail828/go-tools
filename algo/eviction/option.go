package eviction

import "time"

type Option func(*cache)

func WithExpiration(t time.Duration) Option {
	return func(c *cache) { c.expiration = t }
}

func WithOnEviction(f func(string, any)) Option {
	return func(c *cache) { c.onEvict = f }
}

func WithOnPurge(f func(string, any)) Option {
	return func(c *cache) { c.onPurge = f }
}


