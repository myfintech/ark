package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/myfintech/ark/src/go/lib/arkplugins/cicd-gha-runner/pkg"

	"github.com/myfintech/ark/src/go/lib/log"
)

func main() {

	opts := pkg.CiCdOptions{}

	stdInBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("unable to read from Standard In")
	}
	if err = json.Unmarshal(stdInBytes, &opts); err != nil {
		log.Fatal("unable to unmarshal bytes into plugin options interface")
	}
	manifest, err := pkg.NewCiCdGhaRunnersManifest(opts)

	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
	}

	_, _ = os.Stdout.Write([]byte(manifest))
}
