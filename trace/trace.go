package trace

import "github.com/caddyserver/caddy/v2"

type Tracers struct{}

func (ts *Tracers) Provision(caddy.Context) error {
	return nil
}

func (ts *Tracers) Validate() error {
	return nil
}
