package middleware

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"runtime/debug"
)

// PanicLoggerMiddleware logs the panic error in addition to the stacktrace that caused the panic
func PanicLoggerMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Panic: ", string(debug.Stack()))
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(struct {
						Message string `json:"message"`
					}{"oops, something bad happened"})
				}
			}()
			next.ServeHTTP(w, req)
		})
	})
}
