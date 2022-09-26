// +build cmd

package cmd

import (
	// because of how go modules work wire cannot resolve its own command without
	// being explicitly imported
	_ "github.com/google/wire/cmd/wire"
)
