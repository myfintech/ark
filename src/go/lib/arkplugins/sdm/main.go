package main

import (
	"os"

	"github.com/myfintech/ark/src/go/lib/arkplugins/sdm/pkg"
)

func main() {

	manifest, err := pkg.NewSDMManifest()

	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
	}

	_, _ = os.Stdout.Write([]byte(manifest))
}
