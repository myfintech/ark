package probe

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark"

	"github.com/stretchr/testify/require"
)

func TestTarget(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	target := Target{
		DialAddress:    "https://www.google.com",
		Timeout:        "5s",
		Delay:          "1s",
		MaxRetries:     5,
		ExpectedStatus: 200,
		RawTarget: ark.RawTarget{
			Name:  "probe_test",
			Type:  Type,
			File:  filepath.Join(cwd, "targets_test.go"),
			Realm: cwd,
		},
	}
	err = target.Validate()
	require.NoError(t, err)

	checksum, err := target.Checksum()
	require.NoError(t, err)

	artifact, err := target.Produce(checksum)
	require.NoError(t, err)

	probe := artifact.(*Artifact)
	require.Equal(t, "d8493e5c5d2f0769d16a30d6959a60f57bb605be678df83bb4865d9ea10bd55e", probe.Hash)
}
