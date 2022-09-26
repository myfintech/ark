package internal_net

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/pkg/errors"
)

var netRegExp = regexp.MustCompile("^(tcp|udp)")
var httpRegExp = regexp.MustCompile("^(http|https)")

// ProbeOptions configuration for the RunProbe function
type ProbeOptions struct {
	Timeout        time.Duration
	Delay          time.Duration
	MaxRetries     int
	ExpectedStatus int
	Address        *url.URL
	// TLSVerify      bool
	OnError func(err error, remainingAttempts int)
}

// ProbeFunc a function which returns an error if a target cannot be reached or fails its probe chekc
type ProbeFunc func() error

// ProbeFuncFactory a generic connection interface to wrap dialer's
type ProbeFuncFactory func(ctx context.Context, options ProbeOptions) (ProbeOptions, ProbeFunc)

// CreateNetProbe uses default net.DialTimeout to test connectivity to a TCP or UDP endpoint
func CreateNetProbe(options ProbeOptions) (ProbeOptions, ProbeFunc) {
	return options, func() error {
		conn, err := net.DialTimeout(options.Address.Scheme, options.Address.Host, options.Timeout)
		defer func() {
			if conn != nil {
				_ = conn.Close()
			}
		}()
		if err != nil {
			return err
		}
		return nil
	}
}

// CreateHTTPProbe uses default http.Get to test connectivity to an HTTP|HTTPS endpoint
func CreateHTTPProbe(options ProbeOptions) (ProbeOptions, ProbeFunc) {
	return options, func() error {
		// TODO: implement TLSVerify
		// http.Client{
		// 	Transport:     nil,
		// 	CheckRedirect: nil,
		// 	Jar:           nil,
		// 	Timeout:       0,
		// }
		req, err := http.Get(options.Address.String())
		if err != nil {
			return err
		}

		defer func() {
			_ = req.Body.Close()
		}()

		if options.ExpectedStatus == 0 {
			options.ExpectedStatus = http.StatusOK
		}

		if req.StatusCode != options.ExpectedStatus {
			return errors.Errorf("http probe failed %s %s, expected %d got %d",
				req.Request.Method, req.Request.URL.String(), options.ExpectedStatus, req.StatusCode)
		}

		return nil
	}
}

func CreateProbe(options ProbeOptions) (ProbeFunc, ProbeOptions, error) {
	var probeFunc ProbeFunc

	if options.Address == nil {
		return probeFunc, options, errors.New("Address cannot be nil")
	}

	switch {
	case httpRegExp.MatchString(options.Address.Scheme):
		_, probeFunc = CreateHTTPProbe(options)
		break
	case netRegExp.MatchString(options.Address.Scheme):
		_, probeFunc = CreateNetProbe(options)
		break
	default:
		return probeFunc, options, errors.Errorf("no supported probe for scheme %s", options.Address.Scheme)
	}

	return probeFunc, options, nil
}

// RunProbe attempts to validate connectivity to an arbitrary network and port
func RunProbe(options ProbeOptions, probeFunc ProbeFunc) error {
	options.MaxRetries--

	err := probeFunc()
	if err != nil && options.OnError != nil {
		options.OnError(err, options.MaxRetries)
	}

	if err != nil && options.MaxRetries <= 0 {
		return err
	}

	if err != nil {
		time.Sleep(options.Delay)
		return RunProbe(options, probeFunc)
	}

	return nil
}
