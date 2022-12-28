package logging

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2"
)

func init() {
	caddy.RegisterModule(Logging{})
}

type Logging struct {
	*Writers
	Loggers []json.RawMessage `json:"loggers" caddy:"namespace=config.logging.loggers inline_key=logger"`
}

func (logging *Logging) Validate() error {
	return nil
}

func (d Logging) ID() string {
	return "config.logging"
}

// CaddyModule returns the Caddy module information.
func (d Logging) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  caddy.ModuleID(d.ID()),
		New: func() caddy.Module { return new(Logging) },
	}
}

// Provision implement caddy.Provisioner, execute callbacks with default config
func (d *Logging) Provision(ctx caddy.Context) error {
	if d.Writers != nil {
		err := d.Writers.Provision(ctx)
		if err != nil {
			return err
		}
	}
	_, err := ctx.LoadModule(d, "Loggers")
	if err != nil {
		return fmt.Errorf("%s load loggers failed: %v", d.ID(), err)
	}
	d.Loggers = nil // allow GC to deallocate
	return nil
}

// Interface guard
var (
	_ caddy.Provisioner = (*Logging)(nil)
)
