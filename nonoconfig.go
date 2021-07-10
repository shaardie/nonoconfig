package nonoconfig

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

// NoNoConfig is returned from the NewNoNoConfig and is the interface from
// which config entries can be received via the `Config` function.
type NoNoConfig struct {
	fs []string
	c  interface{}
}

// NewNoNoConfig creates a NoNoConfig from a list of possible configuration file.
// First found configuration file will be used.
func NewNoNoConfig(configFiles ...string) *NoNoConfig {
	return &NoNoConfig{
		fs: configFiles,
	}
}

// Config tries to store the config value found under the chain of keys in the variable `value` points to.
// If anything goes wrong, an error should be return.
func (nnc *NoNoConfig) Config(value interface{}, keys ...interface{}) error {

	out := reflect.ValueOf(value)
	if out.Kind() != reflect.Ptr {
		return fmt.Errorf("value is %T, not pointer", value)
	}
	if out.IsNil() {
		return errors.New("value is nil pointer")
	}

	if nnc.c == nil {
		err := nnc.updateConfig()
		if err != nil {
			return fmt.Errorf("unable to update config, %w", err)
		}
	}
	v := reflect.ValueOf(nnc.c)

	// Find entry
	for _, key := range keys {

		if v.Kind() != reflect.Map {
			return fmt.Errorf("key %v is not a map", key)
		}

		k := reflect.ValueOf(key)
		v = v.MapIndex(k)
		if !v.IsValid() {
			return fmt.Errorf("key %v not found", key)
		}
	}
	return recursiveReflection(v, out)
}

// configurationFile iterates over the configuration file canditates
// and returns the first found file.
func (nnc *NoNoConfig) configurationFile() (string, error) {
	for _, f := range nnc.fs {
		stat, err := os.Stat(f)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", fmt.Errorf("unable to stat %v, %w", f, err)
		}
		if stat.IsDir() {
			continue
		}
		return f, nil
	}
	return "", errors.New("no configuration file found")
}

// updateConfig reads and parses the first found configuration file and stores it the struct.
func (ncc *NoNoConfig) updateConfig() error {
	f, err := ncc.configurationFile()
	if err != nil {
		return fmt.Errorf("unable to determine configuration file, %w", err)
	}
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return fmt.Errorf("unable to read configuration file %v, %w", f, err)
	}
	err = yaml.Unmarshal(b, &ncc.c)
	if err != nil {
		return fmt.Errorf("error while unmarshaling configuration file %v, %w", f, err)
	}
	return nil
}

func recursiveReflection(in, out reflect.Value) error {

	// Verify out
	if !out.IsValid() {
		return fmt.Errorf("out %v is not valid", out)
	}
	outKind := out.Kind()
	if outKind != reflect.Ptr && outKind != reflect.Interface {
		return fmt.Errorf("out is of kind %v, neither pointer nor interface", outKind)
	}
	out = out.Elem()
	if !out.CanSet() {
		return fmt.Errorf("out %v not set-able", out)
	}
	if !in.IsValid() {
		return fmt.Errorf("input %v is not valid", in)
	}

	// Verify in
	if in.IsZero() {
		out.Set(reflect.Zero(out.Type()))
		return nil
	}
	if !in.CanInterface() {
		return fmt.Errorf("%v not interface-able", in)
	}

	// This should go better
	in = reflect.ValueOf(in.Interface())

	switch out.Kind() {
	case reflect.Interface:
		out.Set(in)
	case reflect.String:
		out.SetString(in.String())
	case reflect.Float32, reflect.Float64:
		out.SetFloat(in.Float())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		out.SetInt(in.Int())
	case reflect.Bool:
		out.SetBool(in.Bool())
	case reflect.Map:
		if in.Kind() != reflect.Map {
			return fmt.Errorf("%v of type %v, not map", in, in.Kind())
		}
		if out.IsNil() {
			out.Set(reflect.MakeMap(out.Type()))
		}
		iter := in.MapRange()
		for iter.Next() {
			key := reflect.New(out.Type().Key())
			if err := recursiveReflection(iter.Key(), key); err != nil {
				return err
			}

			value := reflect.New(out.Type().Elem())
			if err := recursiveReflection(iter.Value(), value); err != nil {
				return err
			}

			out.SetMapIndex(key.Elem(), value.Elem())
		}
	case reflect.Array, reflect.Slice:
		if in.Kind() != reflect.Array && in.Kind() != reflect.Slice {
			return fmt.Errorf("%v of type %v, not array", in, in.Kind())
		}

		for j := 0; j < in.Len(); j++ {
			value := reflect.New(out.Type().Elem())
			if err := recursiveReflection(in.Index(j), value); err != nil {
				return err
			}
			out.Set(reflect.Append(out, value.Elem()))
		}
	case reflect.Struct:
		if in.Kind() != reflect.Map {
			return fmt.Errorf(
				"%v of type %v, not map, so it can't be converted to struct", in, in.Kind())
		}

		for j := 0; j < out.NumField(); j++ {
			typeField := out.Type().Field(j)
			tag, ok := typeField.Tag.Lookup("nonoconfig")
			if !ok {
				tag = typeField.Name
			}
			r := in.MapIndex(reflect.ValueOf(tag))
			if !r.IsValid() {
				continue
			}
			err := recursiveReflection(r, out.Field(j).Addr())
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unknown kind %v", out.Kind())
	}
	return nil
}
