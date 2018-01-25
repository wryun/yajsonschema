package main

import (
  "fmt"
  "log"
  "flag"
  "os"
  "encoding/json"
  "io/ioutil"

  "github.com/wryun/yajsonschema"
  "github.com/xeipuuv/gojsonschema"
)

func main() {
  log.SetFlags(0)

  var schemaFilename string
  flag.StringVar(&schemaFilename, "schema", "", "yaml schema file to use")
  flag.StringVar(&schemaFilename, "s", "", "yaml schema file to use")
  flag.Parse()

  if schemaFilename == "" {
    log.Fatal("must specify --schema argument")
  }

  schemaReader, err := os.Open(schemaFilename)
  if err != nil {
    log.Fatal(err)
  }

  jsonschema, err := yajsonschema.Convert(schemaReader)
  if err != nil {
    log.Fatal(err)
  }

  if flag.NArg() == 0 {
    if output, err := json.MarshalIndent(jsonschema, "", "  "); err != nil {
      log.Fatal(err)
    } else {
      fmt.Println(string(output))
    }
  } else {
    exitCode := 0
    // validate input files against schema
    schemaLoader := gojsonschema.NewGoLoader(jsonschema)

    for _, filename := range flag.Args() {
      log.SetPrefix(filename + ": ")
      document, err := ioutil.ReadFile(filename)
      if err != nil {
        log.Fatal(err)
      }
      documentLoader := gojsonschema.NewStringLoader(string(document))
      if result, err := gojsonschema.Validate(schemaLoader, documentLoader); err != nil {
        log.Fatal(err)
      } else if !result.Valid() {
        for _, desc := range result.Errors() {
          log.Print(desc)
        }

        exitCode = 2
      }
    }

    os.Exit(exitCode)
  }
}
