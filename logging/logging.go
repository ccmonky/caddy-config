package logging

import "github.com/caddyserver/caddy/v2"

type Logging struct{}

func (logging *Logging) Provision(caddy.Context) error {
	return nil
}

func (logging *Logging) Validate() error {
	return nil
}
