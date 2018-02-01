package yajsonschema

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-test/deep"
)

var update = flag.Bool("update", false, "update .json files")

func TestAllCorrectSchemas(t *testing.T) {
	t.Parallel()

	matches, err := filepath.Glob(filepath.Join("testdata", "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, inputFileName := range matches {
		testName := strings.TrimSuffix(inputFileName, ".yaml")
		outputFileName := testName + ".json"

		t.Run(testName, func(t *testing.T) {
			//t.Parallel()
			testSchema(t, inputFileName, outputFileName)
		})
	}
}

func testSchema(t *testing.T, inputFileName, outputFileName string) {
	yamlSchemaFile, err := os.Open(inputFileName)
	if err != nil {
		t.Fatal(err)
	}

	actualOutput, err := Convert(yamlSchemaFile)
	if err != nil {
		t.Fatal(err)
	}
	// We have to deserialise and reserialise this to ensure
	// that the types match.
	actualBytesOutput, err := json.MarshalIndent(actualOutput, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if *update {
		ioutil.WriteFile(outputFileName, actualBytesOutput, 0644)
	}

	expectedBytesOutput, err := ioutil.ReadFile(outputFileName)
	if err != nil {
		t.Fatal(err)
	}
	var expected interface{}
	if err = json.Unmarshal(expectedBytesOutput, &expected); err != nil {
		t.Fatal(err)
	}
	var actual interface{}
	if err = json.Unmarshal(actualBytesOutput, &actual); err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(actual, expected); diff != nil {
		t.Error(diff)
	}
}
