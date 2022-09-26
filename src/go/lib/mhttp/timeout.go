package mhttp

import (
	"net/http"
	"time"

	"github.com/myfintech/ark/src/go/lib/utils/httpresputils"
)

import (
	"context"
)

// Timeout is a middleware that cancels ctx after a given timeout and return
// a 504 Gateway Timeout error to the client.
//
// It's required that you select the ctx.Done() channel to check for the signal.
// If the context has reached its deadline and return, otherwise the timeout
// signal will be just ignored.
//
// ie. a route/handler may look like:
//
//  r.Get("/long", func(w http.ResponseWriter, r *http.Request) {
// 	 ctx := r.Request()
// 	 processTime := time.Duration(rand.Intn(4)+1) * time.Second
//
// 	 select {
// 	 case <-ctx.Done():
// 	 	return
//
// 	 case <-time.After(processTime):
// 	 	 // The above channel simulates some hard work.
// 	 }
//
// 	 w.Write([]byte("done"))
//  })
//