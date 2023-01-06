package dynconf

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/ccmonky/pkg/jsonschema"
	jsgen "github.com/invopop/jsonschema"
	"github.com/sevlyar/retag"
	"github.com/xeipuuv/gojsonschema"
)

// used for test
var Validate func(any, []byte) error

type DefaultRetager struct{}

func (rt DefaultRetager) Convert(p interface{}, maker jsonschema.TagMaker) interface{} {
	return retag.Convert(p, maker)
}

func (rt DefaultRetager) ConvertAny(p interface{}, maker jsonschema.TagMaker) interface{} {
	return retag.ConvertAny(p, maker)
}

func DeleteJsonOmitemptyMarker() retag.TagMaker {
	return deleteJsonOmitemptyMarker{}
}

type deleteJsonOmitemptyMarker struct{}

func (m deleteJsonOmitemptyMarker) MakeTag(t reflect.Type, fieldIndex int) reflect.StructTag {
	field := t.Field(fieldIndex)
	if strings.HasPrefix(t.Name(), "RegCallback[") { // NOTE: ignore wrapper!
		return field.Tag
	}
	jsonTag, ok := field.Tag.Lookup("json")
	if !ok {
		return ""
	}
	if strings.Contains(jsonTag, ",omitempty") {
		key := strings.Split(jsonTag, ",")[0]
		return reflect.StructTag(fmt.Sprintf(`json:"%s"`, key))
	}
	return reflect.StructTag(fmt.Sprintf(`json:"%s"`, jsonTag))
}

type DefaultGenerator struct{}

func (g DefaultGenerator) Reflect(v interface{}) ([]byte, error) {
	return json.Marshal(jsgen.Reflect(v))
}

func (g DefaultGenerator) ReflectFromType(t reflect.Type) ([]byte, error) {
	return json.Marshal(jsgen.ReflectFromType(t))
}

func DefaultValidate(schema, data []byte) error {
	result, err := gojsonschema.Validate(gojsonschema.NewBytesLoader(schema), gojsonschema.NewBytesLoader(data))
	if err != nil {
		return err
	}
	if result.Valid() {
		return nil
	}
	detail := ""
	for _, desc := range result.Errors() {
		detail += fmt.Sprintf("- %s\n", desc)
	}
	return jsonschema.NewValidateFailedError(detail)
}

var validator *jsonschema.Validator

func init() {
	var err error
	validator, err = jsonschema.NewValidator(
		jsonschema.WithRetag(DefaultRetager{}),
		jsonschema.WithTagMaker(DeleteJsonOmitemptyMarker()),
		jsonschema.WithGenerator(DefaultGenerator{}),
		jsonschema.WithValidateFunc(DefaultValidate),
	)
	if err != nil {
		panic(err)
	}
	Validate = validator.Validate
}
