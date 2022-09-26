package probe

import (
	"context"
	"net/url"
	"time"

	"github.com/myfintech/ark/src/go/lib/logz"

	intlnet "github.com/myfintech/ark/src/go/lib/internal_net"
	"github.com/myfintech/ark/src/go/lib/log"
)

// Action is th executor for implementing a network probe
type Action struct {
	Artifact *Artifact
	Target   *Target
	Logger   logz.FieldLogger
}

var _ logz.Injector = &Action{}

// UseLogger injects a logger into the target's action
func (a *Action) UseLogger(logger logz.FieldLogger) {
	a.Logger = logger
}

// Execute runs the action and produces a probe.Artifact
func (a Action) Execute(_ context.Context) (err error) {
	if a.Target.Timeout == "" {
		a.Target.Timeout = "5s"
	}
	timeout, err := time.ParseDuration(a.Target.Timeout)
	if err != nil {
		return err
	}

	if a.Target.Delay == "" {
		a.Target.Delay = "1s"
	}
	delay, err := time.ParseDuration(a.Target.Delay)
	if err != nil {
		return err
	}

	if a.Target.MaxRetries == 0 {
		a.Target.MaxRetries = 5
	}
	addr, err := url.Parse(a.Target.DialAddress)
	if err != nil {
		return err
	}

	probe, options, err := intlnet.CreateProbe(intlnet.ProbeOptions{
		Timeout:        timeout,
		Delay:          delay,
		Address:        addr,
		MaxRetries:     a.Target.MaxRetries,
		ExpectedStatus: a.Target.ExpectedStatus,
		OnError: func(err error, remainingAttempts int) {
			log.Warnf("an error occurred while running probe %s(%s), remaining attempts %d, %v",
				a.Target.Key(), a.Target.DialAddress, remainingAttempts, err)
		},
	})

	if err != nil {
		return err
	}

	if err = intlnet.RunProbe(options, probe); err != nil {
		return err
	}

	return nil
}
