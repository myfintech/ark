package http_server

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"golang.org/x/sync/errgroup"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server/api_errors"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server/http_handlers"
	"github.com/myfintech/ark/src/go/lib/logz"
)

// New creates a new fiber server and storage interface
func New(store ark.Store, logger logz.FieldLogger, broker cqrs.Broker, logFilePath string) *fiber.App {
	// setup the new fiber server and append its configuration
	server := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		// https://github.com/gofiber/fiber/issues/1146#issuecomment-771318720
		// DisableKeepalive:      true,
		IdleTimeout:  time.Second * 10,
		ReadTimeout:  time.Second * 60,
		BodyLimit:    10 * 1024 * 1024,
		ErrorHandler: api_errors.NewErrorHandler(logger),
	})

	server.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	server.Use(pprof.New())

	// POST /targets/add/:targetType
	// Evaluates the provided target by type and responds with its computed artifact
	server.Post("/targets", http_handlers.NewAddTargetHandler(store))

	// GET /targets
	// returns all the targets
	server.Get("/targets", http_handlers.NewListTargetsHandler(store))

	// GET /targets/:targetKey
	// returns all the targets
	// server.Get("/targets?targetKey=", http_handlers.NewGetTargetHandler(store))

	// POST /targets/connect
	// takes a graphEdge and connects the relevant targets in the graph
	server.Post("/targets/connect", http_handlers.NewConnectTargetHandler(store))

	// GET /graph
	// returns a graph
	server.Get("/graph", http_handlers.NewGetGraphHandler(store))

	// GET /graph/render?format=&targetKey=
	// return a visualization of a graph depending on a given format
	server.Get("/graph/render", http_handlers.NewGraphRenderHandler(store))

	// GET /graph/edges
	// returns all graph edges
	server.Get("/graph/edges", http_handlers.NewGetGraphEdgesHandler(store))

	// POST /run
	// triggers an ark execution of a given target
	server.Post("/run", http_handlers.NewRunHandler(store, logger, broker))

	// GET /server/logs
	// returns a followed log stream of the server logs
	server.Get("/server/logs", http_handlers.NewLogsHandler(logFilePath, logger))

	// GET /server/logs/:log_key
	// returns a followed log stream of the logs matching the target's key
	server.Get("/server/logs/:log_key", http_handlers.NewLogsHandler(logFilePath, logger))

	// GET /health
	// returns a general 200 status code to state the server is listening
	server.Get("/health", http_handlers.NewHealthListener())

	return server
}

// NewSubsystem stands up an http_server as a subsystem
func NewSubsystem(addr, logFile string, store ark.Store, logger logz.FieldLogger, broker cqrs.Broker) *subsystems.Process {
	logger = logger.Child(logz.WithFields(logz.Fields{
		"system": topics.HTTPServer.String(),
	}))
	return &subsystems.Process{
		Name: topics.HTTPServer.String(),
		Factory: func(wg *sync.WaitGroup, ctx context.Context) func() error {
			server := New(store, logger, broker, logFile)
			eg, egCtx := errgroup.WithContext(ctx)

			eg.Go(func() error {
				wg.Done()
				logger.Infof("listening on http://%s", addr)
				return server.Listen(addr)
			})

			return func() error {
				select {
				case <-ctx.Done():
					logger.Info("recieved shutdown signal")
					return server.Shutdown()
				case <-egCtx.Done():
					return eg.Wait()
				}
			}
		},
	}
}
