/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package middleware

import (
	"github.com/gorilla/mux"
	"net/http"
)

// Chain stitches together different middlewares and provides a single middleware.
func Chain(middlewareFuncs ...mux.MiddlewareFunc) mux.MiddlewareFunc {
	return mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		for _, m := range middlewareFuncs {
			next = m(next)
		}
		return next
	})
}
