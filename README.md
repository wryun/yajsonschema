**not ready yet - need to add custom tag support mentioned below for
it to be useable**

# yajsonschema

yajsonschema (yet another json schema [something]) is a tool which converts
a schema written in yaml into a JSON schema (draft 4). i.e. it's primarily
an easier way to write json schemas, but you can also use it directly.

However, you can only represent a subset of possible json schema
documents.

## Example

    ---
    name: !type {
      type: string,
      minLength: 1
    }
    ---
    fileType: !enum [json, yaml]
    organisation: !type string
    evil?: !type bool
    people:
    - name: !ref name
      age: !type {
        type: int,
        min: 0
      }
      -: false # definition of additional properties

Things you may notice here:

 - in general, we assume that the document should simply match itself
   (i.e. a minimal case would be the same document - it should
   validate itself unless...)
 - we use yaml custom types (e.g. '!ref') to abbreviate things
 - normal json schemas default to non-required. Instead we default to
   required (using '?' to represent optional).
 - ways to break out of 'this should match itself':
   - an array indicates that items should match one of the items
     in the array (i.e. anyOf), not that it should be an exact match
   - !ref myname is shorthand for {"$ref": "#/definitions/myname"}
   - !type (object) = standard type definition
     - this is a useful way to 'break out' into standard json schema.
       You can use definition references to jump back to yajsonschema.
   - !type int = type definition where the type is set to int
     and no other validation is applied
   - !enum [arr] gives you a list of possibilities for this value
   - for objects, we understand '?' at the end of fields to indicate
     that it's an optional field, and '-' to be the 'additionalProperties'
     field in the schema (with special handling for true/false so that
     they're passed through with their semantic meaning rather than
     intepreted as an enum/const).

If only one document is defined, we assume there are no definitions.

## Using as a library

See usage in `cmd/yajsonschema.go`

## Using as a CLI tool

To generate the json schema (on stdout):

  yajsonschema -s schema.yaml

To validate json documents immediately:

  yajsonschema -s schema.yaml myinput.yaml myinput.json

(error code of 2 indicates failure to validate; output on stderr)

## Similar work

There's a bunch other validators and 'json schema generators'.
I don't know of any others that rely on yaml custom types like this,
or have the rough principle of self-validation.
