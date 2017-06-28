package atool

import (
	"testing"
	"github.com/alecthomas/assert"
	"fmt"
	"os"
)

func TestStructsFile(t *testing.T) {
	data, _, err := StructsFile("test/sample.go")
	assert.Nil(t, err, "Failed open sample file")
	assert.Equal(t, len(data), 2)
	assert.Equal(t, "Fuel", data[0].Name)
	assert.Equal(t, "Rocket", data[1].Name)
}

func TestStructsFile_fail(t *testing.T) {
	data, _, err := StructsFile("test/undefined.go")
	assert.Nil(t, data)
	assert.True(t, os.IsNotExist(err))
}

func TestInterfacesFile(t *testing.T) {
	list, p, err := InterfacesFile("test/sample.go")
	assert.Nil(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Control", list[0].Name)
	var names []string
	for _, m := range list[0].Methods {
		names = append(names, m.Name)
		if len(m.In) > 0 {
			fmt.Printf("Arg type: %v\n", p.ToString(m.In[0].Type))
		}
	}
	assert.EqualValues(t, []string{"Land", "IsLanded", "Aircraft", "Launch"}, names)
	assert.NotNil(t, list[0].Method("IsLanded"))
	assert.Nil(t, list[0].Method("Nothing"))
}

func ExampleStructsFile() {
	// Print structs and fields count in file
	structs, _, err := StructsFile("test/sample.go")
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
