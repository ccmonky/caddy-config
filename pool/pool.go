package pool

import "github.com/caddyserver/caddy/v2"

type Pool struct{}

func (pool *Pool) Provision(caddy.Context) error {
	return nil
}

func (pool *Pool) Validate() error {
	return nil
}
