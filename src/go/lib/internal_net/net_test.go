package internal_net

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestProbe(t *testing.T) {
	t.Run("should successfully probe ephemeral server", func(t *testing.T) {
		listener, err := net.Listen("tcp", "0.0.0.0:")
		require.NoError(t, err)
		defer func() { _ = listener.Close() }()

		probe, options, err := CreateProbe(ProbeOptions{
			Timeout:    10 * time.Second,
			Delay:      1 * time.Second,
			MaxRetries: 2,
			Address: &url.URL{
				Scheme: "tcp",
				Host:   listener.Addr().String(),
			},
		})
		require.NoError(t, err)

		err = RunProbe(options, probe)
		require.NoError(t, err)
	})

	t.Run("should handle http status codes", func(t *testing.T) {
		t.Parallel()
		tcpListener, err := net.Listen("tcp", "0.0.0.0:")
		require.NoError(t, err)
		defer func() { _ = tcpListener.Close() }()

		go func() {
			_ = http.Serve(tcpListener, http.NotFoundHandler())
		}()

		t.Run("should fail when unexpected status code is returned", func(t *testing.T) {
			probe, options, pErr := CreateProbe(ProbeOptions{
				Timeout:        10 * time.Second,
				Delay:          1 * time.Second,
				MaxRetries:     0,
				ExpectedStatus: http.StatusOK,
				Address: &url.URL{
					Scheme: "http",
					Host:   tcpListener.Addr().String(),
				},
			})
			require.NoError(t, pErr)
			require.Error(t, RunProbe(options, probe))
		})
		t.Run("should succeed when expected status code is returned", func(t *testing.T) {
			probe, options, pErr := CreateProbe(ProbeOptions{
				Timeout:        10 * time.Second,
				Delay:          1 * time.Second,
				MaxRetries:     0,
				ExpectedStatus: http.StatusNotFound,
				Address: &url.URL{
					Scheme: "http",
					Host:   tcpListener.Addr().String(),
				},
			})
			require.NoError(t, pErr)
			require.NoError(t, RunProbe(options, probe))
		})
	})

	t.Run("should fail to probe unbound port", func(t *testing.T) {
		probe, options, err := CreateProbe(ProbeOptions{
			Timeout:    2 * time.Second,
			Delay:      1 * time.Second,
			MaxRetries: 0,
			Address: &url.URL{
				Scheme: "tcp",
				Host:   "0.0.0.0:30000",
			},
		})
		require.NoError(t, err)
		require.Error(t, RunProbe(options, probe))
	})

	t.Run("should fail to probe unbound port", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		eg, _ := errgroup.WithContext(ctx)
		eg.Go(func() error {
			probe, options, err := CreateProbe(ProbeOptions{
				Timeout:    5 * time.Second,
				Delay:      2 * time.Second,
				MaxRetries: 10,
				Address: &url.URL{
					Scheme: "tcp",
					Host:   "0.0.0.0:30000",
				},
			})
			if err != nil {
				return err
			}

			return RunProbe(options, probe)
		})

		time.Sleep(3 * time.Second)
		listener, err := net.Listen("tcp", "0.0.0.0:30000")
		require.NoError(t, err)
		defer func() { _ = listener.Close() }()

		require.NoError(t, eg.Wait())
	})
}
