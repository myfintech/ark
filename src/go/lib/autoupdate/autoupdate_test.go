package autoupdate

import (
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/gofiber/fiber/v2/middleware/compress"

	"github.com/myfintech/ark/src/go/lib/fs"

	"golang.org/x/sync/errgroup"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

var localVersion = LocalVersionInfo{
	Name:                 "test",
	Version:              "0.0.9",
	VersionCheckEndpoint: "http://0.0.0.0:35999/version",
}

func init() {
	httpServer := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	httpServer.Use(compress.New())

	httpServer.Get("/version", func(c *fiber.Ctx) error {
		return c.JSON(RemoteVersionInfo{
			Version: "0.1.0",
			OSPackages: []OSPackage{
				{
					URL:      "http://0.0.0.0:35999/bin",
					OS:       runtime.GOOS,
					Arch:     runtime.GOARCH,
					Checksum: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
				},
			},
		})
	})

	httpServer.Get("/bin", func(c *fiber.Ctx) error {
		return c.Send([]byte("test"))
	})

	if err := ListenerRoutine(httpServer, ":35999"); err != nil {
		panic(err)
	}
}

func ListenerRoutine(httpServer *fiber.App, addr string) error {
	waitForListen := make(chan bool)

	eg, _ := errgroup.WithContext(context.Background())

	eg.Go(func() error {
		close(waitForListen)
		return httpServer.Listen(addr)
	})

	<-waitForListen
	return nil
}

func TestCheckVersion(t *testing.T) {
	Init(&localVersion) // FIXME: add check for the server before running tests as there is a race condition where the test runs before the server is available; see https://buildkite.com/mantl/mantl-mono/builds/1897#b0b351b6-2349-45da-88c5-ad4ae4d29c19 for example of race failure

	updatable, remoteVersionInfo, err := CheckVersion()
	require.NoError(t, err)
	require.True(t, updatable)
	require.Equal(t, "0.1.0", remoteVersionInfo.Version)
}

func TestUpgrade(t *testing.T) {
	Init(&localVersion) // FIXME: add check for the server before running tests as there is a race condition where the test runs before the server is available

	testbin, err := os.CreateTemp("", "test")
	require.NoError(t, err)

	err = Upgrade(testbin.Name())
	require.NoError(t, err)

	data, err := fs.ReadFileString(testbin.Name())
	require.NoError(t, err)
	require.Equal(t, "test", data)

	err = os.RemoveAll(testbin.Name())
	require.NoError(t, err)
}

func TestSelectVersionByOS(t *testing.T) {
	selected := SelectVersionByOS([]OSPackage{
		{
			Arch: "amd64",
			OS:   "darwin",
		},
		{
			Arch: "amd64",
			OS:   "linux",
		},
	})
	require.NotNil(t, selected)
	require.Equal(t, runtime.GOOS, selected.OS)
	require.Equal(t, runtime.GOARCH, selected.Arch)

	selected = SelectVersionByOS([]OSPackage{
		{
			Arch: "amd64",
			OS:   "linux",
		},
		{
			Arch: "amd64",
			OS:   "darwin",
		},
	})
	require.NotNil(t, selected)
	require.Equal(t, runtime.GOOS, selected.OS)
	require.Equal(t, runtime.GOARCH, selected.Arch)
}
