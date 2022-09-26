package vault_tools

import (
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

// NewVaultFunc returns a function for use in Go's templating language which may be used to retrieve secrets
