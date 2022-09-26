package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/myfintech/ark/src/go/lib/arkplugins/vanity-domain/pkg"
	"github.com/myfintech/ark/src/go/lib/log"
)

func main() {
	opts := new(pkg.Options)

	stdInBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("unable to read from Standard In")
	}
	fmt.Println(string(stdInBytes))

	if err = json.Unmarshal(stdInBytes, &opts); err != nil {
		log.Fatalf("unable to unmarshal bytes into plugin options struct: %v", err)
	}

	manifest, err := pkg.NewManifest(*opts)
	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
	}

	_, _ = os.Stdout.Write([]byte(manifest))
}
