package pkg

import (
	"bytes"
	"path/filepath"

	"github.com/myfintech/ark/src/go/lib/kube/objects"

	"github.com/myfintech/ark/src/go/lib/kube/statefulapp"
)

type Options struct {
	statefulapp.Options
	DefaultUsername string `json:"defaultUsername"`
	DefaultDatabase string `json:"defaultDatabase"`
	DefaultPassword string `json:"defaultPassword"`
}

func NewPostgresManifest(opts Options) (string, error) {
	buff := new(bytes.Buffer)

	opts.Env["PGDATA"] = filepath.Join("/mnt/", opts.Name, "data")

	if opts.DataDir != "" {
		opts.Env["PGDATA"] = filepath.Join("/mnt", opts.DataDir, "data")
	}

	if password, ok := opts.Env["POSTGRES_PASSWORD"]; ok {
		opts.DefaultPassword = password
	}

	if username, ok := opts.Env["POSTGRES_USER"]; ok {
		opts.DefaultUsername = username
	}

	if database, ok := opts.Env["POSTGRES_DB"]; ok {
		opts.DefaultDatabase = database
	}

	if opts.DefaultDatabase == "" {
		opts.DefaultDatabase = "postgres"
	}

	probeAction := &objects.ExecAction{
		Command: []string{
			"psql", "-U", opts.DefaultUsername, "-d", opts.DefaultDatabase, "-c", "SELECT 1",
		},
	}

	// immediately check the database is accepting connections
	opts.ReadinessProbe = &objects.ProbeOptions{
		Handler: objects.Handler{
			Exec: probeAction,
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       5,
		FailureThreshold:    10,
	}

	// every 30s check if the database is still accepting connections
	opts.LivenessProbe = &objects.ProbeOptions{
		Handler: objects.Handler{
			Exec: probeAction,
		},
		InitialDelaySeconds: 45,
		TimeoutSeconds:      5,
		PeriodSeconds:       30,
		FailureThreshold:    3,
	}

	manifest := statefulapp.NewStatefulApp(opts.Options)

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}
	return buff.String(), nil
}
