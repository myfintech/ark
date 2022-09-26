package pkg

import (
	"bytes"
	"fmt"

	"github.com/myfintech/ark/src/go/lib/kube/objects"

	"github.com/myfintech/ark/src/go/lib/kube/statefulapp"
)

func NewRedisManifest(opts statefulapp.Options) (string, error) {
	buff := new(bytes.Buffer)

	probeAction := &objects.ExecAction{
		Command: []string{
			"redis-cli", "-p", fmt.Sprintf("%d", opts.Port), "ping",
		},
	}

	// immediately check the database is accepting connections
	opts.ReadinessProbe = &objects.ProbeOptions{
		Handler:             objects.Handler{Exec: probeAction},
		InitialDelaySeconds: 10,
		FailureThreshold:    10,
		TimeoutSeconds:      5,
		PeriodSeconds:       5,
	}

	// every 30s check if the database is still accepting connections
	opts.LivenessProbe = &objects.ProbeOptions{
		Handler:             objects.Handler{Exec: probeAction},
		InitialDelaySeconds: 45,
		TimeoutSeconds:      5,
		FailureThreshold:    3,
		PeriodSeconds:       30,
	}

	manifest := statefulapp.NewStatefulApp(opts)

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}
	return buff.String(), nil
}
