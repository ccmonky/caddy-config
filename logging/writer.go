package logging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/ccmonky/typemap"
	"github.com/pkg/errors"
)

func init() {
	caddy.RegisterModule(Writers{})
}

// Writers define a list of caddy.logging.writers, which can be referenced later by name
type Writers struct {
	Writers []Writer `json:"writers,omitempty" caddy:"namespace=config.logging.writers"`

	ctx caddy.Context
}

// Writer define a caddy.logging.writer config
type Writer struct {
	Name      string          `json:"name"`
	ConfigRaw json.RawMessage `json:"config" caddy:"namespace=caddy.logging.writers inline_key=output"`
}

// ID 获取模块ID
func (Writers) ID() string {
	return "config.logging.writers"
}

// CaddyModule returns the Caddy module information.
func (w Writers) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  caddy.ModuleID(w.ID()),
		New: func() caddy.Module { return new(Writers) },
	}
}

// Provision 实现Provisioner
func (w *Writers) Provision(ctx caddy.Context) error {
	w.ctx = ctx
	sentinel := map[string]struct{}{}
	for _, conf := range w.Writers {
		name := conf.Name
		if name == "" {
			return errors.Errorf("%s encounter empty name", w.ID())
		}
		if _, ok := sentinel[name]; ok {
			return errors.Errorf("%s %s repeated", w.ID(), name)
		}
		val, err := ctx.LoadModule(&conf, "ConfigRaw")
		if err != nil {
			return fmt.Errorf("load %s %s failed: %v", w.ID(), name, err)
		}
		opener, ok := val.(caddy.WriterOpener)
		if !ok {
			return errors.Errorf("%s: %s is not a caddyhttp.WriterOpener", w.ID(), name)
		}
		err = typemap.Set[caddy.WriterOpener](ctx, name, opener)
		if err != nil {
			return err
		}
		sentinel[name] = struct{}{}
	}
	return nil
}

// Validate 实现Validator
func (w Writers) Validate() error {
	for _, conf := range w.Writers {
		writer, err := typemap.Get[caddy.WriterOpener](context.TODO(), conf.Name)
		if err != nil {
			return err
		}
		if writer == nil {
			return errors.Errorf("%s %s is nil pointer", w.ID(), conf.Name)
		}
	}
	return nil
}

// Interface guard
var (
	_ caddy.Validator   = (*Writers)(nil)
	_ caddy.Provisioner = (*Writers)(nil)
)
