package mock

import (
	"github.com/caddyserver/caddy/v2"
)

func init() {
	caddy.RegisterModule(Mock{})
}

// Mock define a list of http.handlers configration, which can be referenced later by name
type Mock struct {
	*Matchers
}

// CaddyModule returns the Caddy module information.
func (Mock) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "config.mock",
		New: func() caddy.Module { return new(Mock) },
	}
}

// Provision 实现Provisioner
func (c *Mock) Provision(ctx caddy.Context) error {
	if c.Matchers != nil {
		err := c.Matchers.Provision(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Validate 实现Validator
func (c Mock) Validate() error {
	if c.Matchers != nil {
		err := c.Matchers.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

// Interface guard
var (
	_ caddy.Validator   = (*Mock)(nil)
	_ caddy.Provisioner = (*Mock)(nil)
)
