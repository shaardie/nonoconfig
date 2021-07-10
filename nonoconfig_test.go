package nonoconfig

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoValidConfigurationFile(t *testing.T) {
	nnc := NewNoNoConfig(
		".testdata", // is a directory
		".testdata/does-not-exists.yaml",
	)
	var s interface{}
	assert.Error(t, nnc.Config(&s))
}

func TestUnreadableConfig(t *testing.T) {
	cfgFile := ".testdata/unreadable-config.yaml"
	assert.NoError(t, os.Chmod(cfgFile, 0200))
	defer func() {
		assert.NoError(t, os.Chmod(cfgFile, 0644))
	}()

	nnc := NewNoNoConfig(cfgFile)
	var s interface{}
	assert.Error(t, nnc.Config(&s))

}

func TestInvalidYamlConfig(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/invalid-yaml.yaml")
	var s interface{}
	assert.Error(t, nnc.Config(&s))
}

func TestNilPointerParameter(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")
	var s *interface{} = nil
	assert.Error(t, nnc.Config(s))
}

func TestParameterNotPointer(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")
	var s string = "not a pointer"
	assert.Error(t, nnc.Config(s))
}

func TestNotFound(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")
	var s interface{}
	assert.Error(t, nnc.Config(&s, "invalid"))
	assert.Error(t, nnc.Config(&s, "this", "is", "invalid"))
	assert.Error(t, nnc.Config(&s, "single_string", "no map"))
}

func TestPrimitiveInterface(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")

	var singleInterface interface{}
	assert.NoError(t, nnc.Config(&singleInterface, "single_string"))
	assert.Equal(t, "single_string", singleInterface)
}

func TestPrimitiveString(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")

	var singleString string
	assert.NoError(t, nnc.Config(&singleString, "single_string"))
	assert.Equal(t, "single_string", singleString)
}

func TestPrimitiveInt(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")

	var singleInt int
	assert.NoError(t, nnc.Config(&singleInt, "single_int"))
	assert.Equal(t, 42, singleInt)
}

func TestPrimitiveFloat(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")

	var singleFloat64 float64
	var singleFloat32 float32
	assert.NoError(t, nnc.Config(&singleFloat64, "single_float"))
	assert.Equal(t, 3.141, singleFloat64)
	assert.NoError(t, nnc.Config(&singleFloat32, "single_float"))
	assert.Equal(t, float32(3.141), singleFloat32)
}

func TestPrimitiveBool(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")

	var singleBool bool
	assert.NoError(t, nnc.Config(&singleBool, "single_bool"))
	assert.Equal(t, true, singleBool)
}

func TestMapEmpty(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")

	sInterfaceInterface := make(map[interface{}]interface{})
	assert.NoError(t, nnc.Config(&sInterfaceInterface, "map_null"))
	assert.Equal(t, map[interface{}]interface{}{}, sInterfaceInterface)

	sStringInt := make(map[string]int)
	assert.NoError(t, nnc.Config(&sStringInt, "map_null"))
	assert.Equal(t, map[string]int{}, sStringInt)

}

func TestMapStringToInt(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")

	s := make(map[string]int)
	assert.NoError(t, nnc.Config(&s, "map_string_to_int"))
	assert.Equal(t, map[string]int{
		"first":  1,
		"second": 2,
		"third":  3,
	}, s)
}

func TestMapStringToInterface(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")

	s := make(map[string]interface{})
	assert.NoError(t, nnc.Config(&s))
	assert.IsType(t, map[string]interface{}{}, s)
}

func TestMapInterfaceToInterface(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")

	s := make(map[interface{}]interface{})
	assert.NoError(
		t,
		nnc.Config(&s),
	)
	assert.IsType(t, map[interface{}]interface{}{}, s)
}

func TestMapTypeIntToString(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")

	s := make(map[int]string)
	assert.NoError(
		t,
		nnc.Config(&s, "map_int_to_string"),
	)
	assert.Equal(t, map[int]string{
		1: "first",
		2: "second",
		3: "third",
	}, s)

}

func TestArrayNull(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")

	sInterface := []interface{}{}
	assert.NoError(t, nnc.Config(&sInterface, "array_null"))
	assert.Equal(t, []interface{}{}, sInterface)

	sString := []string{}
	assert.NoError(t, nnc.Config(&sString, "array_null"))
	assert.Equal(t, []string{}, sString)
}

func TestArrayString(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")

	s := []string{}
	assert.NoError(t, nnc.Config(&s, "array_string"))
	assert.Equal(t, []string{"first", "second", "third"}, s)
}

func TestStruct(t *testing.T) {
	nnc := NewNoNoConfig(".testdata/config.yaml")

	type inner struct {
		First  int     `nonoconfig:"first"`
		Second float64 `nonoconfig:"second"`
		Third  bool    `nonoconfig:"third"`
	}

	type outer struct {
		MatchFieldName bool
		NeedATag       bool  `nonoconfig:"need_a_tag"`
		Inner          inner `nonoconfig:"recursive"`
	}

	s := outer{}

	assert.NoError(
		t,
		nnc.Config(&s, "struct"),
	)
	assert.Equal(t, outer{
		MatchFieldName: true,
		NeedATag:       true,
		Inner: inner{
			First:  1,
			Second: 2.0,
			Third:  true,
		},
	}, s)

}

func ExampleNewNoNoConfig() {
	// Creating a new config from a list of possible configuration files.
	// First found configuration file is used.
	nnc := NewNoNoConfig(
		".testdata/does-not-exist.yaml",
		".testdata/config.yaml",
		".testdata/does-not-exist-either.yaml",
	)

	fmt.Printf("%T", nnc)
	// Output: *nonoconfig.NoNoConfig
}

func ExampleNoNoConfig_config_simple() {
	// Create new NoNoConfig
	// File content:
	// ```yaml
	// single_float: 3.141
	// ```
	nnc := NewNoNoConfig(".testdata/config.yaml")

	// Get a float
	var f float64
	err := nnc.Config(&f, "single_float")
	if err != nil {
		fmt.Println("Unable to get float, %w", err)
		os.Exit(1)
	}
	fmt.Println(f)
	// Output: 3.141
}

func ExampleNoNoConfig_Config_complex() {
	// Create new NoNoConfig
	// File content:
	// ```yaml
	// complex_type:
	//   map:
	//   first: 1
	//   second: 2
	//   third: 3
	//   array:
	//   - first
	//   - second
	//   - third
	//   float: 3.141
	// ```
	nnc := NewNoNoConfig(".testdata/config.yaml")

	// Contruct complex type with annotation with name mapping.
	ct := struct {
		Map   map[string]int `nonoconfig:"map"`
		Array []string       `nonoconfig:"array"`
		Float float64        `nonoconfig:"float"`
	}{}

	// Read data
	err := nnc.Config(&ct, "complex_type")
	if err != nil {
		fmt.Println("Unable to get complex type, %w", err)
		os.Exit(1)
	}
	fmt.Println(ct)
	// Output: {map[first:1 second:2 third:3] [first second third] 3.141}
}
