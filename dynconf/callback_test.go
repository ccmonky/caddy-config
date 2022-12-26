package dynconf_test

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/ccmonky/caddy-config/dynconf"
	"github.com/ccmonky/typemap"
)

func TestCallback(t *testing.T) {
	// 1. register callbacks
	ctx := context.Background()
	typemap.MustRegister[dynconf.Callback](ctx, "group:data_id", dynconf.NewRegCallback[bool]("degrade"))
	typemap.MustRegister[dynconf.Callback](ctx, "group:data_id:log", dynconf.CallbackFunc(func(sourceKey, data string) error {
		log.Println(sourceKey, data)
		return nil
	}))
	// 2. configure callbacks
	data := []byte(`{
		"defaults": [
			{
				"keys": ["group:data_id"],
				"default": {
					"name": "degrade",
					"value": false
				}
			}
		]
	}`)
	callbacks := &dynconf.Callbacks{}
	err := json.Unmarshal(data, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	caddyCtx, _ := caddy.NewContext(caddy.Context{Context: context.Background()})
	err = callbacks.Provision(caddyCtx)
	if err != nil {
		t.Fatal(err)
	}
	b, err := typemap.Get[bool](ctx, "degrade")
	if err != nil {
		t.Fatal(err)
	}
	if b != false {
		t.Fatal("should == false")
	}
	// 3. test callbacks for new data
	var cbs []dynconf.Callback
	for _, name := range []string{"group:data_id", "group:data_id:log"} {
		cb, err := typemap.Get[dynconf.Callback](ctx, name)
		if err != nil {
			t.Fatal(err)
		}
		cbs = append(cbs, cb)
	}
	for _, cb := range cbs {
		err := cb.Callback("group:data_id", `{
			"name":  "degrade",
			"value": true
		}`)
		if err != nil {
			t.Fatal(err)
		}
	}
	b, err = typemap.Get[bool](ctx, "degrade")
	if err != nil {
		t.Fatal(err)
	}
	if b != true {
		t.Fatal("should == false")
	}
}
