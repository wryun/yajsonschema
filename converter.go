package yajsonschema

import (
	"errors"
	"io"
	"strings"

	// go-yaml doesn't have custom tag support yet... grabbing random fork.
	// see: https://github.com/go-yaml/yaml/issues/191

	"github.com/go-yaml/yaml"
)

// JSONSchema is a json schema represented as Go types.
// If using gojsonschema, this can be used with NewGoLoader, or
// can be output using json.Marshal.
type JSONSchema map[string]interface{}
type jsonType string

const (
	optionalSigil = "?"

	objectType = jsonType("object")
	arrayType  = jsonType("array")
)

// Convert processes either 1 or 2 yajsonschema documents accessible via io.Reader,
// and outputs a json schema.
func Convert(yamlSchemaReader io.Reader) (JSONSchema, error) {
	yamlDefinitions, yamlSchema, err := unmarshal(yamlSchemaReader)
	if err != nil {
		return nil, err
	}

	jsonSchema, err := buildFragment(yamlSchema)
	if err != nil {
		return nil, err
	}

	if yamlDefinitions != nil {
		typedYamlDefinitions, ok := yamlDefinitions.(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("first document - definitions - is not an object")
		}
		jsonSchemaDefinitions := make(map[string]interface{}, len(typedYamlDefinitions))
		for untypedKey, value := range typedYamlDefinitions {
			key, ok := untypedKey.(string)
			if !ok {
				return nil, errors.New("first document - definitions - has non-string keys")
			}
			jsonSchemaDefinitions[key], err = buildFragment(value)
			if err != nil {
				return nil, err
			}
		}
		jsonSchema["definitions"] = jsonSchemaDefinitions
	}

	jsonSchema["$schema"] = "http://json-schema.org/draft-04/schema#"
	return jsonSchema, nil
}

func unmarshal(yamlSchema io.Reader) (interface{}, interface{}, error) {
	dec := yaml.NewDecoder(yamlSchema)
	documents := []interface{}{}
	for {
		var document interface{}
		err := dec.Decode(&document)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		documents = append(documents, document)
	}

	var definitions, schema interface{}
	switch len(documents) {
	case 0:
		return nil, nil, errors.New("no yaml documents in schema input")
	default:
		return nil, nil, errors.New("more than two yaml documents in schema input")
	case 1:
		schema = documents[0]
	case 2:
		definitions = documents[0]
		schema = documents[1]
	}

	return definitions, schema, nil
}

func buildFragment(yamlSchema interface{}) (JSONSchema, error) {
	switch val := yamlSchema.(type) {
	default:
		return JSONSchema(map[string]interface{}{
			"enum": []interface{}{val}, // draft 04 doesn't support const
		}), nil
	case []interface{}:
		return buildArraySchema(val)
	case map[interface{}]interface{}:
		return buildObjectSchema(val)
	}
}

func buildArraySchema(arr []interface{}) (JSONSchema, error) {
	var err error
	schema := JSONSchema(map[string]interface{}{
		"type": arrayType,
	})

	switch len(arr) {
	case 0:
		break
	case 1:
		schema["items"], err = buildFragment(arr[0])
		if err != nil {
			return nil, err
		}
	default:
		items := map[string][]interface{}{
			"anyOf": []interface{}{},
		}
		schema["items"] = items
		for _, yamlSchema := range arr {
			anyOfSchema, err := buildFragment(yamlSchema)
			if err != nil {
				return nil, err
			}
			items["anyOf"] = append(items["anyOf"], anyOfSchema)
		}
	}

	return schema, nil
}

func buildObjectSchema(obj map[interface{}]interface{}) (JSONSchema, error) {
	var err error
	var properties map[string]interface{}
	var required []string
	schema := JSONSchema(map[string]interface{}{
		"type": objectType,
	})

	for untypedField, yamlSchema := range obj {
		field, ok := untypedField.(string)
		if !ok {
			return nil, errors.New("object has non-string key")
		}
		if field != "-" {
			if strings.HasSuffix(field, optionalSigil) {
				field = strings.TrimSuffix(field, optionalSigil)
			} else {
				required = append(required, field)
			}

			if properties == nil {
				properties = map[string]interface{}{}
				schema["properties"] = properties
			}
			properties[field], err = buildFragment(yamlSchema)
			if err != nil {
				return nil, err
			}
		} else {
			switch yamlSchema.(type) {
			case bool:
				// i.e. we can't actually handle the case when the value of the additional
				// property must be a boolean constant. Oh well.
				schema["additionalProperties"] = yamlSchema
			default:
				schema["additionalProperties"], err = buildFragment(yamlSchema)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if required != nil {
		schema["required"] = required
	}

	return schema, nil
}
