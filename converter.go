package yajsonschema

import (
	"errors"
	"io"
	"sort"
	"strings"

	// go-yaml doesn't have custom tag support yet... grabbing random fork.
	// see: https://github.com/go-yaml/yaml/issues/191
	"github.com/wryun/yaml"
)

const (
	optionalSigil = "?"

	objectType = "object"
	arrayType  = "array"
)

// Convert processes either 1 or 2 yajsonschema documents accessible via io.Reader,
// and outputs a json schema.
func Convert(yamlSchemaReader io.Reader) (map[string]interface{}, error) {
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
	dec.RegisterCustomTagUnmarshaller("!enum", &CustomTagUnmarshaler{
		func(unmarshalYaml func(interface{}) error) (interface{}, error) {
			enum := enumTag{}
			if err := unmarshalYaml(&enum.Contents); err != nil {
				return nil, err
			}

			return enum, nil
		},
	})
	dec.RegisterCustomTagUnmarshaller("!ref", &CustomTagUnmarshaler{
		func(unmarshalYaml func(interface{}) error) (interface{}, error) {
			ref := refTag{}
			if err := unmarshalYaml(&ref.Contents); err != nil {
				return nil, err
			}

			return ref, nil
		},
	})
	dec.RegisterCustomTagUnmarshaller("!type", &CustomTagUnmarshaler{
		func(unmarshalYaml func(interface{}) error) (interface{}, error) {
			t := typeTag{}
			if err := unmarshalYaml(&t.Contents); err != nil {
				return nil, err
			}

			return t, nil
		},
	})

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

type CustomTagUnmarshaler struct {
	doit func(func(interface{}) error) (interface{}, error)
}

func (ctu *CustomTagUnmarshaler) UnmarshalYAML(yamlUnmarshal func(interface{}) error) (interface{}, error) {
	return ctu.doit(yamlUnmarshal)
}

type enumTag struct {
	Contents []interface{}
}
type typeTag struct {
	Contents interface{}
}
type refTag struct {
	Contents string
}

func buildFragment(yamlSchema interface{}) (map[string]interface{}, error) {
	switch val := yamlSchema.(type) {
	default:
		return map[string]interface{}(map[string]interface{}{
			"enum": []interface{}{val}, // draft 04 doesn't support const
		}), nil
	case []interface{}:
		return buildArraySchema(val)
	case map[interface{}]interface{}:
		return buildObjectSchema(val)
	case enumTag:
		return map[string]interface{}(map[string]interface{}{
			"enum": val.Contents,
		}), nil
	case typeTag:
		switch typeContents := val.Contents.(type) {
		case map[interface{}]interface{}:
			result := make(map[string]interface{}, len(typeContents))
			for k, v := range typeContents {
				result[k.(string)] = v
			}
			return result, nil
		case string:
			return map[string]interface{}{
				"type": typeContents,
			}, nil
		default:
			return nil, nil
		}
	case refTag:
		return map[string]interface{}{
			"$ref": "#/definitions/" + val.Contents,
		}, nil
	}
}

func buildArraySchema(arr []interface{}) (map[string]interface{}, error) {
	var err error
	schema := map[string]interface{}(map[string]interface{}{
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

func buildObjectSchema(obj map[interface{}]interface{}) (map[string]interface{}, error) {
	var err error
	var properties map[string]interface{}
	var required []string
	schema := map[string]interface{}(map[string]interface{}{
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
		// So it's stable. Mostly for testing.
		sort.Strings(required)
		schema["required"] = required
	}

	return schema, nil
}
