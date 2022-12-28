package dynconf

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2"
)

func init() {
	caddy.RegisterModule(Dynconf{})
}

type Dynconf struct {
	Callbacks `json:"callbacks"`
	Listeners []json.RawMessage `json:"listeners" caddy:"namespace=config.ext.dynconf.listeners inline_key=listener"`
}

func (d Dynconf) ID() string {
	return "config.ext.dynconf"
}

// CaddyModule returns the Caddy module information.
func (d Dynconf) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  caddy.ModuleID(d.ID()),
		New: func() caddy.Module { return new(Dynconf) },
	}
}

// Provision implement caddy.Provisioner, execute callbacks with default config
func (d *Dynconf) Provision(ctx caddy.Context) error {
	err := d.Callbacks.Provision(ctx)
	if err != nil {
		return fmt.Errorf("provision callbacks failed: %v", err)
	}
	_, err = ctx.LoadModule(d, "Listeners")
	if err != nil {
		return fmt.Errorf("%s load listeners failed: %v", d.ID(), err)
	}
	d.Listeners = nil // allow GC to deallocate
	return nil
}

// Interface guard
var (
	_ caddy.Provisioner = (*Dynconf)(nil)
)
