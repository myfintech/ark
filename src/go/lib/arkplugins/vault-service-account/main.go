package main

import (
	"os"

	"github.com/myfintech/ark/src/go/lib/arkplugins/vault-service-account/pkg"
)

func main() {

	manifest, err := pkg.NewVaultServiceAccountManifest()

	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
	}

	_, _ = os.Stdout.Write([]byte(manifest))
}
