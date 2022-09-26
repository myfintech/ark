package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/myfintech/ark/src/go/lib/kube/microservice"
	"github.com/myfintech/ark/src/go/lib/log"
)

func main() {
	// adding comments
	opts := microservice.Options{
		Replicas: 1,
	}
	stdInBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("unable to read from Standard In")
	}
	if err = json.Unmarshal(stdInBytes, &opts); err != nil {
		log.Fatal("unable to unmarshal bytes into plugin options interface")
	}

	manifest := microservice.NewMicroService(opts)

	_ = manifest.Serialize(os.Stdout)
}
