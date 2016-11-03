package httpx

import (
	"context"
	"net/http"
	"time"
)

// CloseHandler cancels the context if the client closes the connection.
func CloseHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if cn, ok := w.(http.CloseNotifier); ok {
			ctx, cancel := context.WithCancel(req.Context())
			req = req.WithContext(ctx)
			defer cancel()

			closed := cn.CloseNotify()
			go func() {
				select {
				case <-ctx.Done():
					// do nothing
				case <-closed:
					cancel()
				}
			}()
		}

		h.ServeHTTP(w, req)
	})
}

// DeadlineHandler returns a Handler which adds a deadline to the context.
//
// Child handlers are responsible for obeying the context deadline and returning
// an appropriate error (or not) response in case of timeout.
func DeadlineHandler(deadline time.Time) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, cancel := context.WithDeadline(req.Context(), deadline)
			req = req.WithContext(ctx)
			defer cancel()

			h.ServeHTTP(w, req)
		})
	}
}

// TimeoutHandler returns a Handler which adds a timeout to the context.
//
// Child handlers are responsible for obeying the context deadline and returning
// an appropriate error (or not) response in case of timeout.
func TimeoutHandler(timeout time.Duration) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, cancel := context.WithTimeout(req.Context(), timeout)
			req = req.WithContext(ctx)
			defer cancel()

			h.ServeHTTP(w, req)
		})
	}
}
