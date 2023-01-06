package dynconf_test

import (
	"testing"

	"github.com/ccmonky/caddy-config/dynconf"
	"github.com/ccmonky/pkg/jsonschema"
	"github.com/stretchr/testify/assert"
)

// DynType should not use omitempty tag, this is a test!
type DynType struct {
	Int    int `json:"int,omitempty"`
	String string
	Uint   uint `json:"-"`
	Embed  struct {
		A string `json:"a,omitempty"`
		B int
		C uint `json:"-"`
	}
}

func TestDeleteJsonOmitemptyMarker(t *testing.T) {
	data := []byte(`{
		"name": "xxx",
		"value": {
			"int": 0,
			"String": "",
			"Embed": {
				"a": "",
				"B": 1
			}
		}
	}`)
	dt := dynconf.RegCallback[DynType]{}
	err := dynconf.Validate(&dt, data)
	assert.Nilf(t, err, "validate good")
	data = []byte(`{
		"name": "xxx",
		"value": {
			"int": 0,
			"String": "",
			"Embed": {
				"B": 1
			}
		}
	}`)
	err = dynconf.Validate(dt, data)
	assert.Truef(t, jsonschema.IsValidateFailedError(err), "validate error")
	assert.Equalf(t, err.Error(), "jsonschema: - value.Embed: a is required\n", "err detail")
}
