package hclutils

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
)

var mapStringInterfaceKind = reflect.TypeOf(map[string]interface{}{}).String()
var sliceStringKind = reflect.TypeOf([]string{}).String()
var mapStringCtyValueKind = reflect.TypeOf(map[string]cty.Value{}).String()

// MapStringStringToCtyObject returns a cty.Object from map[string]string
func MapStringStringToCtyObject(v map[string]string) cty.Value {
	attrs := map[string]cty.Value{}
	for key, value := range v {
		attrs[key] = cty.StringVal(value)
	}
	return cty.ObjectVal(attrs)
}

// StringSliceToCtyTuple returns a slice of strings as a slice of cty.Value
func StringSliceToCtyTuple(v []string) cty.Value {
	var cv []cty.Value
	for _, value := range v {
		cv = append(cv, cty.StringVal(value))
	}
	return cty.TupleVal(cv)
}

// MapStringInterfaceToCty attempts to case a map[string]interface{} to a map[string]cty.Value
// This will panic if it cannot convert a specified interface
func MapStringInterfaceToCty(v map[string]interface{}) (variables map[string]cty.Value) {
	variables = map[string]cty.Value{}
	for key, val := range v {
		typeOf := reflect.TypeOf(val)
		kindOf := typeOf.Kind()
		switch {
		case kindOf == reflect.String:
			variables[key] = cty.StringVal(val.(string))
			continue
		case kindOf == reflect.Int32:
			variables[key] = cty.NumberIntVal(int64(val.(int32)))
			continue
		case kindOf == reflect.Int64:
			variables[key] = cty.NumberIntVal(val.(int64))
			continue
		case kindOf == reflect.Float32:
			variables[key] = cty.NumberFloatVal(float64(val.(float32)))
			continue
		case kindOf == reflect.Float64:
			variables[key] = cty.NumberFloatVal(val.(float64))
			continue
		case typeOf.String() == mapStringCtyValueKind:
			variables[key] = cty.ObjectVal(val.(map[string]cty.Value))
			continue
		case typeOf.String() == mapStringInterfaceKind:
			variables[key] = cty.ObjectVal(MapStringInterfaceToCty(val.(map[string]interface{})))
			continue
		case typeOf.String() == sliceStringKind:
			variables[key] = StringSliceToCtyTuple(val.([]string))
		case kindOf == reflect.Bool:
			variables[key] = cty.BoolVal(val.(bool))
		default:
			panic(errors.Errorf("cannot convert type %s %s to cty.Value", typeOf, kindOf.String()))
		}
	}
	return
}

// MergeMapStringCtyValue merges n number of map[string]string
func MergeMapStringCtyValue(maps ...map[string]cty.Value) map[string]cty.Value {
	merged := map[string]cty.Value{}
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}
