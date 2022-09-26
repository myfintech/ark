package nix

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestNix(t *testing.T) {
	ctx := context.Background()
	if nixPath, err := exec.LookPath("nix-env"); err != nil || nixPath == "" {
		t.Skip(errors.Wrap(err, "skipping nix tests"))
	}

	cwd, err := os.Getwd()
	require.NoError(t, err)

	target := &Target{
		RawTarget: ark.RawTarget{
			Name:  "example",
			Type:  Type,
			File:  filepath.Join(cwd, "targets_test.go"),
			Realm: cwd,
		},
		Packages: []string{
			"nixpkgs.watchman",
			"nixpkgs.k9s",
		},
	}

	action := &Action{
		Target: target,
	}

	require.Implements(t, (*ark.Action)(nil), action)

	require.NoError(t, target.Validate())

	checksum, err := target.Checksum()
	require.NoError(t, err)

	artifact, err := target.Produce(checksum)
	require.NoError(t, err)

	err = action.Execute(ctx)
	require.NoError(t, err)

	result := artifact.(*Artifact)
	require.True(t, result.Cacheable())

	cached, err := result.LocallyCached(ctx)
	require.True(t, cached)
	require.NoError(t, err)

	// cleanup after installations
	for _, p := range target.Packages {
		var name string
		packageName := strings.Split(p, ".")
		if len(packageName) > 2 {
			name = packageName[len(packageName)-1]
		} else {
			name = packageName[1]
		}
		cmd := exec.Command("nix-env", "-e", name)
		require.NoError(t, cmd.Run())
	}
}
