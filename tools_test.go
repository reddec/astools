package atool

import (
	"fmt"
	"github.com/alecthomas/assert"
	"os"
	"testing"
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

func TestScan(t *testing.T) {
	file, err := Scan("test/sample.go")
	assert.Nil(t, err, "Failed open sample file")
	assert.Equal(t, "sample", file.Package)
	assert.Equal(t, len(file.Structs), 2)
	assert.Equal(t, "Fuel", file.Structs[0].Name)
	assert.Equal(t, "Rocket", file.Structs[1].Name)
	assert.Len(t, file.Interfaces, 1)
	assert.Equal(t, "Control", file.Interfaces[0].Name)
	var names []string
	for _, m := range file.Interfaces[0].Methods {
		names = append(names, m.Name)
	}
	assert.EqualValues(t, []string{"Land", "IsLanded", "Aircraft", "Launch"}, names)
	assert.NotNil(t, file.Interfaces[0].Method("IsLanded"))
	assert.Nil(t, file.Interfaces[0].Method("Nothing"))
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

	args := list[0].Method("Launch").Out
	assert.False(t, args[0].IsError())
	assert.True(t, args[1].IsError())
}

func TestFile_ExtractType(t *testing.T) {
	f, err := Scan("test/sample.go")
	assert.Nil(t, err)
	tp := f.Interface("Fs").Method("Call").In[0]
	ex, err := f.ExtractTypeString(tp.GolangType())
	assert.Nil(t, err)
	assert.False(t, tp.IsPointer())
	t.Log(ex.printer.ToString(ex.Definition))
	assert.Equal(t, "github.com/shopspring/decimal", ex.File.Import)

	tp = f.Interface("Fs").Method("Call").In[1]
	ex, err = f.ExtractTypeString(tp.GolangType())
	assert.Nil(t, err)
	assert.True(t, tp.IsPointer())
	assert.Equal(t, "bytes", ex.File.Import)
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
	// Rocket 6
}
