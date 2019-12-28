package tzstats

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
)

func parseStructs(dec *json.Decoder, v reflect.Value, elemType reflect.Type) error {
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if d, ok := t.(json.Delim); ok && d == '[' {
			curr := reflect.New(elemType).Elem()
			for i := 0; i < curr.Type().NumField(); i++ {
				tag := curr.Type().Field(i).Tag.Get(tagName)
				if tag == "" || tag == "-" {
					continue
				}

				t, err := dec.Token()
				if err != nil {
					return err
				}

				f := curr.Field(i)
				if f.CanSet() {
					switch val := t.(type) {
					case float64:
						switch f.Kind() {
						case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
							f.SetInt(int64(val))
						case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint8, reflect.Uint64:
							f.SetUint(uint64(val))
						case reflect.String:
							f.SetString(fmt.Sprintf("%f", val))
						default:
							f.SetFloat(val)
						}
					case string:
						f.SetString(val)
					case nil:
					default:
						log.Printf("Unknown field: %#v %T", val, val)
					}
				}
			}

			v.Set(reflect.Append(v, curr))

			if _, err := dec.Token(); err != nil { // read close item delimiter
				return err
			}
		}
	}
	return nil
}

func parseSimpleTypes(dec *json.Decoder, v reflect.Value) error {
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if d, ok := t.(json.Delim); ok && d == '[' {
			curr := make([]interface{}, 0)

			for dec.More() {
				t, err := dec.Token()
				if err != nil {
					return err
				}
				switch val := t.(type) {
				default:
					curr = append(curr, val)
				}

			}

			v.Set(reflect.Append(v, reflect.ValueOf(curr)))
		}
	}
	return nil
}

func parseCount(dec *json.Decoder, v reflect.Value) error {
	var count int64
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if d, ok := t.(json.Delim); ok && d == '[' {
			count++
		}
	}
	v.SetInt(count)
	return nil
}

func parseJSON(r io.Reader, s interface{}) error {
	v := reflect.ValueOf(s)
	elemType := getParentType(s)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v = reflect.New(v.Type())
		}
		v = reflect.Indirect(v)
	}
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Int {
		return fmt.Errorf("Invalid type of response: %T", v)
	}

	dec := json.NewDecoder(r)

	_, err := dec.Token()
	if err != nil {
		return err
	}

	switch elemType.Kind() {
	case reflect.Struct:
		return parseStructs(dec, v, elemType)
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		return parseCount(dec, v)
	default:
		return parseSimpleTypes(dec, v)
	}
}
