package pkg

import (
	"bytes"
	"path/filepath"

	"github.com/myfintech/ark/src/go/lib/kube/statefulapp"
)

func NewVaultManifest(opts statefulapp.Options) (string, error) {
	buff := new(bytes.Buffer)

	opts.Env["VAULT_DATA_VOLUME"] = filepath.Join("/mnt/", opts.Name)

	if opts.DataDir != "" {
		opts.Env["VAULT_DATA_VOLUME"] = filepath.Join("/mnt", opts.DataDir)
	}

	manifest := statefulapp.NewStatefulApp(opts)

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}
	return buff.String(), nil
}
