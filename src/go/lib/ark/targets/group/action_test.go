package group

import (
	"context"
	"os"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark"

	"github.com/stretchr/testify/require"
)

func TestAction(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	target := &Target{
		RawTarget: ark.RawTarget{
			Name:  "group_test",
			Type:  Type,
			File:  "test",
			Realm: cwd,
		},
	}

	err = target.Validate()
	require.NoError(t, err)

	action := &Action{
		Target: target,
	}

	err = action.Execute(context.Background())
	require.NoError(t, err)
}
