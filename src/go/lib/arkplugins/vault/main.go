package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/myfintech/ark/src/go/lib/arkplugins/vault/pkg"
	"github.com/myfintech/ark/src/go/lib/kube/statefulapp"
	"github.com/myfintech/ark/src/go/lib/log"
)

func main() {
	opts := statefulapp.Options{
		Replicas:    1,
		Port:        8200,
		ServicePort: 8200,
		Image:       os.Getenv("IMAGE_URL"),
		ServiceType: "ClusterIP",
		Env:         make(map[string]string),
	}
	stdInBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("unable to read from Standard In")
	}
	if err = json.Unmarshal(stdInBytes, &opts); err != nil {
		log.Fatal("unable to unmarshal bytes into plugin options interface")
	}

	manifest, err := pkg.NewVaultManifest(opts)
	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
	}

	_, _ = os.Stdout.Write([]byte(manifest))
}
