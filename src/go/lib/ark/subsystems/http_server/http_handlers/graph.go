package http_handlers

import (
	"bytes"
	osexec "os/exec"

	"github.com/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/myfintech/ark/src/go/lib/ark"
	api_errors2 "github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server/api_errors"
	"github.com/myfintech/ark/src/go/lib/exec"
)

func NewGetGraphEdgesHandler(store ark.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		edges, err := store.GetGraphEdges()
		if err != nil {
			return api_errors2.InternalServerError.
				WithErr(err)
		}

		return c.JSON(edges)
	}
}

func NewGetGraphHandler(store ark.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		graph, err := store.GetGraph()
		if err != nil {
			return api_errors2.InternalServerError.
				WithErr(err)
		}

		return c.JSON(graph)
	}
}

func NewGraphRenderHandler(store ark.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		graph, err := store.GetGraph()
		if err != nil {
			return api_errors2.InternalServerError.
				WithErr(err)
		}

		if targetKey := c.Query("targetKey", ""); targetKey != "" {
			target, err := store.GetTargetByKey(targetKey)
			if err != nil {
				return api_errors2.NotFoundError.
					WithErr(err)
			}
			graph = graph.Isolate(target)
		}

		format := c.Query("format", "svg")

		switch format {
		case "json":
			data, err := graph.MarshalJSON()
			if err != nil {
				return err
			}
			return c.JSON(string(data))
		case "text":
			data := graph.String()
			return c.SendString(data)
		case "dot":
			data := graph.Dot(nil)
			return c.SendString(string(data))
		case "svg":
			dotCMD, err := osexec.LookPath("dot")
			if err != nil {
				return err
			}
			c.Set("Content-Disposition", "inline")
			c.Set("Content-Type", "image/svg+xml")
			return exec.LocalExecutor(exec.LocalExecOptions{
				Command:          []string{dotCMD, "-Tsvg"},
				Stdin:            bytes.NewBuffer(graph.Dot(nil)),
				Stdout:           c, // os.Stdout,
				Stderr:           c,
				InheritParentEnv: true,
			}).Run()
		default:
			return api_errors2.BadRequest.
				WithErr(errors.Errorf("%s is not a valid export format \n", format))
		}
	}
}
