package handlers

import (
	"log"
	"net/http"

	"github.com/marklap/cproject"
)

type ListingResponse []string

func ListingHandler(logger *log.Logger, pathPrefixes []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp ListingResponse

		for _, prefix := range pathPrefixes {
			prefixFiles, err := cproject.ListDir(prefix)

			if err != nil {
				logger.Printf("failed to list prefix dir: %s", prefix)
				continue
			}

			resp = append(resp, prefixFiles...)
		}

		w.Header().Set("Content-Type", "application/json")
		WriteJSON(w, resp)
	})
}
