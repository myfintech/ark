package mhttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/myfintech/ark/src/go/lib/log"
	"github.com/myfintech/ark/src/go/lib/mnet"
	"github.com/myfintech/ark/src/go/lib/utils"
)

var (
	traceContextKey  = "mantl_trace_id"
	loggerContextKey = "mantl_context_logger"

	lengthContextKey = "response_length"
	statusContextKey = "response_status_code"

	// MissingContextLogger indicates that you have not installed the required middleware
	MissingContextLogger = errors.New("there was no context logger on this request, check you installed the http MiddlewareContextLogger")
)

func requestFields(r *http.Request) log.Fields {
	// reqURL := secureURL(r.URL)
	reqURL := r.URL
	return log.Fields{
		"request": map[string]interface{}{
			"method":     r.Method,
			"protocol":   r.Proto,
			"host":       reqURL.Hostname(),
			"port":       reqURL.Port(),
			"path":       reqURL.Path,
			"url":        reqURL.Hostname() + reqURL.Path,
			"referer":    r.Header.Get("Referer"),
			"size":       r.Header.Get("Content-Length"),
			"remote_ip":  mnet.ClientIPFromRequest(r),
			"user_agent": r.Header.Get("User-Agent"),
		},
	}
}

func responseFields(status int, length int, duration time.Duration) log.Fields {
	return log.Fields{
		"response": map[string]interface{}{
			"status":      status,
			"size":        length,
			"duration_s":  duration.Seconds(),
			"duration_ms": duration.Milliseconds(),
			"duration_ns": duration.Nanoseconds(),
		},
	}
}

func traceLogger() (*log.Entry, string) {
	traceID := utils.UUIDV4()
	return log.WithFields(log.Fields{
		"trace_id": traceID,
	}), traceID
}

// MiddlewareContextLogger instruments the request with a context logger and attaches a trace header to every request

// LoggerFromRequest extracts and returns the logger from the request
// if one was not present one will be created to prevent a panic from accessing a nil pointer
// However, it will not be scoped to the request. If you care about this check the error value.
func LoggerFromRequest(r *http.Request) (ctxLogger *log.Entry, traceID string, err error) {
	traceID, _ = r.Context().Value(&traceContextKey).(string)
	ctxLogger, logOk := r.Context().Value(&loggerContextKey).(*log.Entry)

	if !logOk {
		ctxLogger, traceID = traceLogger()
		err = MissingContextLogger
		return
	}

	return
}

func statusFromRequest(r *http.Request) (status int, ok bool) {
	status, ok = r.Context().Value(&statusContextKey).(int)
	return
}

func lengthFromRequest(r *http.Request) (length int, ok bool) {
	length, ok = r.Context().Value(&lengthContextKey).(int)
	return
}

// RequestWithResponseStatusCode does exactly what its name implies

// RequestWithResponseLength does exactly what its name implies

// RequestLoggerMiddleware logs details about the request before and after
