package main

import (
	//"github.com/xeipuuv/gojsonschema"
	"encoding/json"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/wryun/yajsonschema"
	"github.com/xeipuuv/gojsonschema"
)

func main() {
	js.Global.Set("yajsonschema", map[string]interface{}{
		"convert": func(schema string) string {
			jsonschema, err := yajsonschema.Convert(strings.NewReader(schema))
			if err != nil {
				return err.Error()
			} else if output, err := json.MarshalIndent(jsonschema, "", "  "); err != nil {
				return err.Error()
			} else {
				return string(output)
			}
		},
		"validate": func(schema string, input string) string {
			jsonschema, err := yajsonschema.Convert(strings.NewReader(schema))
			if err != nil {
				return err.Error()
			}
			schemaLoader := gojsonschema.NewGoLoader(jsonschema)
			documentLoader := gojsonschema.NewStringLoader(input)
			if result, err := gojsonschema.Validate(schemaLoader, documentLoader); err != nil {
				return err.Error()
			} else if !result.Valid() {
				a := make([]string, len(result.Errors()))
				for i, desc := range result.Errors() {
					a[i] = desc.String()
				}
				return strings.Join(a, "\n")
			} else {
				return "valid"
			}
		},
	})
}
