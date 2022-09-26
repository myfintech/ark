package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/myfintech/ark/src/go/lib/arkplugins/kafka/pkg"
	"github.com/myfintech/ark/src/go/lib/kube/statefulapp"
)

func main() {

	opts := pkg.Options{
		Options: statefulapp.Options{},
	}

	stdInBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("unable to read from Standard In")
	}
	if err = json.Unmarshal(stdInBytes, &opts); err != nil {
		log.Fatal("unable to unmarshal bytes into plugin options interface")
	}

	manifest, err := pkg.NewKafkaManifest(opts)
	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}

	_, _ = os.Stdout.Write([]byte(manifest))
}
