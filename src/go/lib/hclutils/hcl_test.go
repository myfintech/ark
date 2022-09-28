package hclutils

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/myfintech/ark/src/go/lib/utils"
)

func TestDecodeExpressions(t *testing.T) {
	rawHCL := struct {
		Server struct {
			Name hcl.Expression `hcl:"name,attr"`
			Env  hcl.Expression `hcl:"env,attr"`
		} `hcl:"server,block"`
	}{}

	hclFile, diag := FileFromString(`
		server {
			name = "${workspace}/thing"
			env = {
				TEST = "HELLO"
			}
		}
	`)

	if diag != nil && diag.HasErrors() {
		require.NoError(t, diag)
	}

	diag = gohcl.DecodeBody(hclFile.Body, nil, &rawHCL)
	if diag != nil && diag.HasErrors() {
		require.NoError(t, diag)
	}

	// TODO: perform some range input testing
	// We need to validate more cases to ensure that the reflection
	// in DecodeExpressions does not panic
	serverComputed := struct {
		Name string             `hcl:"name,attr"`
		Env  *map[string]string `hcl:"env,attr"`
	}{}

	err := DecodeExpressions(&rawHCL.Server, &serverComputed, &hcl.EvalContext{
		Variables: MapStringInterfaceToCty(map[string]interface{}{
			"workspace": "/some/directory",
		}),
	})
	require.NoError(t, err)
	require.Equal(t, "/some/directory/thing", serverComputed.Name)

	t.Log(utils.MarshalJSONSafe(serverComputed, true))
}
