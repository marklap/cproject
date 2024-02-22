// Ping handler for health/liveness (and early dev).
package handlers

import (
	"fmt"
	"log"
	"net/http"
)

// PingHandler handles requests for `/ping“.
func PingHandler(logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Println("ping")
		fmt.Fprint(w, "pong")
	})
}
