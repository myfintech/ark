package watchman

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	"github.com/myfintech/ark/src/go/lib/watchman/wexp"
)

func TestFileHasher_CalculateRootHash(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	wm, err := Connect(context.Background(), 30)
	require.NoError(t, err)

	resp, err := wm.WatchProject(WatchProjectOptions{Directory: cwd})
	require.NoError(t, err)
	defer func(wm *Client) {
		_, err := wm.DeleteAll()
		if err != nil {
			require.NoError(t, err)
		}
	}(wm)

	// root := cwd
	// if resp.RelPath != "" {
	// 	root = filepath.Join(resp.Watch, resp.RelPath)
	// }

	queryResp, err := wm.Query(QueryOptions{
		Directory: resp.Watch,
		Filter: &QueryFilter{
			Fields:     BasicFields(),
			Expression: wexp.Match("*"),
			DeferVcs:   true,
		},
	})

	require.NoError(t, err)

	hasher := FileHasher(queryResp.Files)
	hash := hasher.CalculateRootHash()
	require.Len(t, hash, 64, "the calculated hash should have a length of 64")

	// TODO: actually verify that the input files properly calculate their sha256
}
