[![Build Status](https://travis-ci.org/wryun/yajsonschema.svg?branch=master)](https://travis-ci.org/wryun/yajsonschema)
[![GoDoc API](https://godoc.org/github.com/wryun/yajsonschema?status.svg)](http://godoc.org/github.com/wryun/yajsonschema)

yajsonschema (Yet Another JSON Schema [something]) is a tool which converts
a schema written in the yaml format defined here into a proper
[JSON Schema](http://json-schema.org/) (draft-04).
i.e. it's primarily an easier way to write json schemas, but you can also use
it directly as a validator.

It tries to be as easy as possible to use by starting from the idea that
a normal document should act as its own schema (i.e. should match itself).

**WARNING: this is largely 'proof-of-concept' and hasn't been used beyond
the test-cases.**

## Example schema

    ---
    name: !type {
      type: string,
      minLength: 1
    }
    ---
    fileType: !enum [json, yaml]
    organisation: !ref name
    evil?: !type boolean
    people:
    - name: !ref name
      age: !type {
        type: integer,
        min: 0
      }
      -: false # definition of JSON Schema additionalProperties
    - funkiness: true

Things you may notice here:

 - there are two YAML documents, the first being a list of definitions
   accessible via ref. If you have only one document, this is the schema.
 - we use yaml custom types (e.g. '!ref') to abbreviate things
 - normal json schemas default to object properties being non-required, but
   yajsonschema defaults to required (using '?' at the end of the
   property name to represent optional)
 - ways to break out of 'this should match itself':
   - an array indicates that items should match any of the items
     in the array (i.e. anyOf), not that it should be an exact match
     to all the items as in json schema
   - !ref myname is shorthand for {"$ref": "#/definitions/myname"}
   - !type (object) = standard type definition
     - this is a useful way to 'break out' into standard json schema.
       You can use definition references to jump back to yajsonschema
       syntax.
   - !type int = type definition where the type is set to int
     and no other validation is applied
   - !enum [arr] gives you a list of possibilities for this value
   - for objects, '?' at the end of properties indicates that it's an optional
     property, and '-' is the 'additionalProperties' property in the schema
     (with special handling for true/false so that they're passed through
     with their semantic meaning rather than intepreted as an enum/const)

## Using as a Go library

See usage in `cmd/yajsonschema.go` and
[API documentation on godoc](https://godoc.org/github.com/wryun/yajsonschema)

## Using as a CLI tool

To generate the json schema (on stdout):

    yajsonschema -s schema.yaml

To validate json documents immediately:

    yajsonschema -s schema.yaml myinput.json myinput2.json

(error code of 2 indicates failure to validate; output on stderr)

## Similar work

There're a bunch other validators and 'json schema generators'.
I don't know of any others that rely on yaml custom types like this,
or are inspired explicitly by the idea of being as close as possible
to the input (i.e. such that one can use input as a
validation schema, and can develop a schema by 'relaxing' an
example input appropriately).

## TODO

- clean up API docs (include examples etc.)
- upstream a better version of custom tag support to go-yaml
- package javascript as npm package
- use a js schema validator rather than the Go one on website
  (bloats JS hugely)
- add some random shields
- 'release' the CLI via github (i.e. binary builds)
