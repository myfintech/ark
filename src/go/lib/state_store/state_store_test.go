package state_store

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/pkg/errors"
)

func TestStateStore(t *testing.T) {
	stateStore := NewStateStore()
	reflectedStateStorePrev := reflect.TypeOf(stateStore.Previous)
	reflectedStateStorePrevVal := reflect.ValueOf(stateStore.Previous)
	now := time.Now()
	second := time.Second
	stringMap := map[string]interface{}{}
	stringMapString := map[string]string{}
	stringMapStringSlice := map[string][]string{}

	testCases := []struct {
		key      string
		value    interface{}
		expected interface{}
		message  string
		method   string
	}{
		{"string", "string", "string", "Should be able to cast a string", "GetString"},
		{"int", 1, 1, "Should be able to cast an int", "GetInt"},
		{"int32", int32(1), int32(1), "Should be able to cast an int32", "GetInt32"},
		{"int64", int64(1), int64(1), "Should be able to cast an int64", "GetInt64"},
		{"float64", float64(1), float64(1), "Should be able to cast an float64", "GetFloat64"},
		{"float64", float64(1), float64(1), "Should be able to cast an float64", "GetFloat64"},
		{"bool", true, true, "should be able to cast bool", "GetBool"},
		{"uint", uint(1), uint(1), "should be able to cast uint", "GetUint"},
		{"uint32", uint32(1), uint32(1), "should be able to cast uint 32", "GetUint32"},
		{"uint64", uint64(1), uint64(1), "should be able to cast uint 64", "GetUint64"},
		{"time", now, now, "should be able to cast time", "GetTime"},
		{"duration", second, second, "should be able to cast duration", "GetDuration"},
		{"intSlice", []int{1}, []int{1}, "should be able to cast int slice", "GetIntSlice"},
		{"stringSlice", []string{"string"}, []string{"string"}, "should be able to cast string slice", "GetStringSlice"},
		{"stringMap", stringMap, stringMap, "should be able to cast string map", "GetStringMap"},
		{"stringMapString", stringMapString, stringMapString, "should be able to cast string map string", "GetStringMapString"},
		{"stringMapStringSlice", stringMapStringSlice, stringMapStringSlice, "should be able to cast string map string slice", "GetStringMapStringSlice"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.message, func(t *testing.T) {
			stateStore.Previous.Set(testCase.key, testCase.value)
			method, exists := reflectedStateStorePrev.MethodByName(testCase.method)
			if !exists {
				require.NoError(t, errors.Errorf("no method %s in state store", testCase.method))
			}
			returnValues := method.Func.Call([]reflect.Value{
				reflectedStateStorePrevVal,
				reflect.ValueOf(testCase.key),
			})
			require.Equal(t, testCase.expected, returnValues[0].Interface())
		})
	}

	t.Run("should return a value with a boolean check", func(t *testing.T) {
		v, exists := stateStore.Previous.GetCheck("fake")
		require.Equal(t, nil, v)
		require.Equal(t, false, exists)
	})

	t.Run("should be able to map an arbitrary value to a struct", func(t *testing.T) {
		person := struct {
			Name string `mapstructure:"name"`
			Age  int    `mapstructure:"age"`
		}{}

		stateStore.Current.Set("mapstructure", map[string]interface{}{
			"name": "test",
			"age":  10,
		})

		require.NoError(t, stateStore.Current.MapStructure("mapstructure", &person))
		require.Equal(t, "test", person.Name)
		require.Equal(t, 10, person.Age)
	})

	t.Run("should be able to unmarshal JSON bytes and store it", func(t *testing.T) {
		people := []struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}{
			{"test", 10},
			{},
		}

		data, err := json.Marshal(people[0])
		require.NoError(t, err)

		require.NoError(t, stateStore.Current.SetFromJSON("json", data))

		require.NoError(t, stateStore.Current.MapStructure("json", &people[1]))

		require.Equal(t, "test", people[1].Name)
		require.Equal(t, 10, people[1].Age)
	})
}
