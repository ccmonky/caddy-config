package dynconf

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/ccmonky/typemap"
)

func init() {
	caddy.RegisterModule(Callbacks{})
}

// Callbacks used to execute some callbacks with default config
//
// Usage:
//
// 1. register callbacks
//
// func init() {
//     typemap.MustRegister[dynconf.Callback](ctx, "group:data_id", dynconf.NewRegCallback[bool]("degrade"))
//     typemap.MustRegister[dynconf.Callback](ctx, "group:data_id:log", dynconf.CallbackFunc(func(sourceKey, data string) error {
//     	   log.Println(sourceKey, data)
//     	   return nil
//     }))
// }
//
// 2. configure callbacks(`config.ext.dynconf.callbacks`)
//
// {
// 	   "config": {
// 	   	   "ext": {
// 	   	   	   "dynconf": {
// 	   	   	   	   "callbacks": {
// 	   	   	   	   	   "defaults": [
// 	   	   	   	   	   	   {
// 	   	   	   	   	   	   	   "keys": ["group:data_id"],
// 	   	   	   	   	   	   	   "default": {
// 	   	   	   	   	   	   	   	   "name": "degrade",
// 	   	   	   	   	   	   	   	   "value": false
// 	   	   	   	   	   	   	   }
// 	   	   	   	   	   	   }
// 	   	   	   	   	   ]
// 	   	   	   	   }
// 	   	   	   }
// 	   	   }
// 	   }
// }
//
// 3. reference callbacks(e.g. nacos)
//
// {
// 	   "config": {
// 	   	   "ext": {
// 	   	   	   "dynconf": {
// 	   	   	   	   "listeners": [
//	                   {
//	                       "listener": "nacos",
//                         "server_config": "",
//                         "client_config": "",
//                         "datas": [
//                             "group": "group"
//	                           "data_id": "data_id",
//                             "callbacks": [
//                                 	"group:data_id",
//                                  "group:data_id:log"
//                             ]
//                         ]
//                     }
//                 ]
// 	   	   	   }
// 	   	   }
// 	   }
// }
type Callbacks struct {
	Defaults []CallbackDefault `json:"defaults,omitempty"`
}

// CallbackDefault define the default config for specified keys
type CallbackDefault struct {
	Keys    []string        `json:"keys"`
	Default json.RawMessage `json:"default"`
}

// ID caddy module id
func (Callbacks) ID() string {
	return "config.ext.dynconf.callbacks"
}

// CaddyModule returns the Caddy module information.
func (cs Callbacks) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  caddy.ModuleID(cs.ID()),
		New: func() caddy.Module { return new(Callbacks) },
	}
}

// Provision implement caddy.Provisioner, execute callbacks with default config
func (cs *Callbacks) Provision(ctx caddy.Context) error {
	for _, def := range cs.Defaults {
		for _, key := range def.Keys {
			cb, err := typemap.Get[Callback](ctx, key)
			if err != nil {
				return fmt.Errorf("get callback %s failed: %v", key, err)
			}
			err = cb.Callback(key, string(def.Default))
			if err != nil {
				return fmt.Errorf("execute callback %s with default value failed: %v", key, err)
			}
		}
	}
	return nil
}

// Callback used as callback for dynamic config source
type Callback interface {
	Callback(sourceKey, data string) error
}

// CallbackFunc callback function
type CallbackFunc func(sourceKey, data string) error

func (cf CallbackFunc) Callback(sourceKey, data string) error {
	return cf(sourceKey, data)
}

// NewRegCallback create a new RegCallback instance
func NewRegCallback[T any](key string) Callback {
	return &RegCallback[T]{
		Name: key, // NOTE: important, will be used to validate equality of the Reg.Name!
	}
}

// RegCallback typemap.Reg[T] as a callback
type RegCallback[T any] typemap.Reg[T]

func (r RegCallback[T]) Callback(sourceKey, data string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 1. log
	var typeNotFound bool
	value, err := typemap.Get[T](ctx, r.Name)
	if err != nil {
		if typemap.IsNotFound(err) {
			typeNotFound = true
			log.Printf("got data %s:%s first time\n", sourceKey, data) // TODO: inject logger
		} else {
			return err
		}
	} else {
		oldData, err := json.Marshal(value)
		if err != nil {
			return err
		}
		log.Printf("change data %s from %s to %s\n", sourceKey, string(oldData), data) // TODO: inject logger
	}
	// 2. validate
	tmp := make(map[string]any)
	err = json.Unmarshal([]byte(data), &tmp)
	if err != nil {
		return err
	}
	if tmp["name"] != r.Name {
		return fmt.Errorf("name %s not equals to default name %s", tmp["name"], r.Name)
	}
	// 3. parse data and inject into typemap
	reg := new(typemap.Reg[T])
	if typeNotFound {
		reg.Action = typemap.RegisterAction
	}
	err = json.Unmarshal([]byte(data), reg)
	if err != nil {
		return err
	}
	return nil
}

// Interface guard
var (
	_ caddy.Provisioner = (*Callbacks)(nil)
)
