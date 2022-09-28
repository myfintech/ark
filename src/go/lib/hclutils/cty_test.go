package hclutils

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zclconf/go-cty/cty"
)

func TestMapStringInterfaceToCty(t *testing.T) {
	t.Run("should be able to cast supported types to cty.Values", func(t *testing.T) {
		variables := MapStringInterfaceToCty(map[string]interface{}{
			"string":  "test",
			"int32":   int32(1),
			"int64":   int64(1),
			"float32": float32(1),
			"float64": float64(1),
			"map": map[string]interface{}{
				"deep": "string",
			},
		})

		// TODO: find a way to determine the values of the cty.Value structs
		// require that they equal what we think they should
		require.IsType(t, cty.String, variables["string"].Type())
		require.IsType(t, cty.Number, variables["int32"].Type())
		require.IsType(t, cty.Number, variables["int64"].Type())
		require.IsType(t, cty.Number, variables["float32"].Type())
		require.IsType(t, cty.Number, variables["float64"].Type())
		require.IsType(t, cty.Object(map[string]cty.Type{}), variables["map"].Type())
	})

	t.Run("should panic when casting an unsupported type", func(t *testing.T) {
		require.Panics(t, func() {
			MapStringInterfaceToCty(map[string]interface{}{
				"nil": nil,
			})
		})
	})
}
