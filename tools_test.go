package atool

import (
	"testing"
	"github.com/alecthomas/assert"
	"fmt"
	"os"
)

func TestStructMapFile(t *testing.T) {
	data, err := StructMapFile("test/sample.go")
	assert.Nil(t, err, "Failed open sample file")
	assert.Equal(t, len(data), 2)
	assert.Equal(t, "Fuel", data[0].Name)
	assert.Equal(t, "Rocket", data[1].Name)
}

func TestStructMapFile_fail(t *testing.T) {
	data, err := StructMapFile("test/undefined.go")
	assert.Nil(t, data)
	assert.True(t, os.IsNotExist(err))
}

func ExampleStructMapFile() {
	// Print structs and fields count in file
	structs, err := StructMapFile("test/sample.go")
	if err != nil {
		panic(err)
	}
	for _, st := range structs {
		fmt.Println(st.Name, st.Definition.Fields.NumFields())
	}
	// Output:
	// Fuel 2
	// Rocket 5
}
