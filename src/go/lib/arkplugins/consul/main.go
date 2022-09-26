package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/myfintech/ark/src/go/lib/arkplugins/consul/pkg"
)

func main() {

	opts := pkg.ConsulOptions{
		Name:     "consul",
		Image:    "hashicorp/consul:1.9.3",
		Replicas: 3,
	}

	stdInBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("unable to read from Standard In")
	}

	if err = json.Unmarshal(stdInBytes, &opts); err != nil {
		log.Fatal("unable to unmarshal bytes into plugin options interface")
	}

	manifest, err := pkg.NewConsulManifest(opts)

	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
	}

	_, _ = os.Stdout.Write([]byte(manifest))
}
