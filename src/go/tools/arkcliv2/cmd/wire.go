//go:build wireinject
// +build wireinject

package cmd

import (
	"github.com/google/wire"
)

func BuildCLI() (*ArkCLI, error) {
	panic(wire.Build(coreSet))
}
