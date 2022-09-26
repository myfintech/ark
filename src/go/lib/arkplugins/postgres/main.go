package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/myfintech/ark/src/go/lib/arkplugins/postgres/pkg"
	"github.com/myfintech/ark/src/go/lib/kube/statefulapp"
	"github.com/myfintech/ark/src/go/lib/log"
)

func main() {
	opts := pkg.Options{
		Options: statefulapp.Options{
			Replicas:    1,
			Port:        5432,
			ServicePort: 5432,
			Image:       "postgres:9",
			ServiceType: "ClusterIP",
			Env: map[string]string{
				"POSTGRES_PASSWORD": "mantl",
			},
		},
		DefaultUsername: "",
		DefaultDatabase: "",
	}
	stdInBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("unable to read from Standard In")
	}
	if err = json.Unmarshal(stdInBytes, &opts); err != nil {
		log.Fatal("unable to unmarshal bytes into plugin options interface")
	}

	manifest, err := pkg.NewPostgresManifest(opts)
	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
	}

	_, _ = os.Stdout.Write([]byte(manifest))
}
