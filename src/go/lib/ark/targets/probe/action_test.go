package probe

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
		DialAddress:    "https://www.google.com",
		Timeout:        "5s",
		Delay:          "1s",
		MaxRetries:     5,
		ExpectedStatus: 200,
		RawTarget: ark.RawTarget{
			Name:  "probe_test",
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
