package http_handlers

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server/api_errors"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/gofiber/fiber/v2"
)

// NewLogsHandler streams logs back to the HTTP client
func NewLogsHandler(logFilePath string, logger logz.FieldLogger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Params("log_key") != "" {
			var err error
			logFilePath, err = logz.SuggestedFilePath("ark/graph", fmt.Sprintf("%s/run.log", c.Params("log_key")))
			if err != nil {
				return api_errors.InternalServerError.
					WithErr(err)
			}
		}
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return api_errors.InternalServerError.
				WithErr(err)
		}
		if err = watcher.Add(logFilePath); err != nil {
			return api_errors.InternalServerError.
				WithErr(err)
		}

		c.Response().SetBodyStreamWriter(func(w *bufio.Writer) {
			defer func() {
				_ = watcher.Close()
			}()

			file, openErr := os.OpenFile(logFilePath, os.O_RDONLY, 0644)
			if openErr != nil {
				logger.Error(openErr)
				return
			}
			defer func() {
				_ = file.Close()
			}()

			reader := bufio.NewReader(file)
			for {
				line, _, readErr := reader.ReadLine()
				if readErr == io.EOF {
					<-watcher.Events
					continue
				}
				if readErr != nil {
					logger.Error(readErr)
					return
				}

				if _, err = w.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
					logger.Error(err)
					return
				}

				if err = w.Flush(); err != nil {
					logger.Error(err)
					return
				}
			}
		})

		return nil
	}
}
