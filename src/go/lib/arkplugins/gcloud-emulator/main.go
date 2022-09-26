package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/myfintech/ark/src/go/lib/arkplugins/gcloud-emulator/pkg"
)

func main() {
	opts := pkg.EmulatorOptions{}

	stdInBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("unable to read from standard In")
	}
	if err = json.Unmarshal(stdInBytes, &opts); err != nil {
		log.Fatal("unable to unmarshal bytes into plugin options interface")
	}

	manifest, err := pkg.NewGcloudEmulatorManifest(opts)
	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
	}

	_, _ = os.Stdout.Write([]byte(manifest))
}
