package ocbootstrap

import (
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

// EnableStackdriverExporter instruments opencensus trace to use stackdriver

// WithStackdriverPropagation instruments an opencensus transport with stackdriver propagation
