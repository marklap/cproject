package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/marklap/cproject"
)

const (
	// DefaultNumLines is the default number of lines to return if none specified.
	DefaultNumLines = 10
)

// TailRequest is a request to tail a file.
type TailRequest struct {
	Path            string   `json:"path"`
	NumLines        int      `json:"num_lines"`
	MatchSubstrings []string `json:"match_substrings"`
	CaseSensitive   bool     `json:"case_sensitive"`
}

// String pretty prints a tail request.
func (r *TailRequest) String() string {
	return fmt.Sprintf("path: %s, num_lines: %d, match_substrings: %s, case_sensitive: %t",
		r.Path, r.NumLines, r.MatchSubstrings, r.CaseSensitive)
}

// TailResponseChunk is a response is a single line from a file.
type TailResponseChunk struct {
	Host string `json:"host"`
	Line string `json:"line"`
}

func validPrefix(path string, pathPrefixes []string) bool {
	// only valid path prefixes are allowed
	for _, pathPrefix := range pathPrefixes {
		if strings.HasPrefix(filepath.Clean(path), pathPrefix) {
			return true
		}
	}
	return false
}

// TailHandler handles requests to tail a log file.
func TailHandler(logger *log.Logger, host string, pathPrefixes []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// decode the incoming request
		decoder := json.NewDecoder(r.Body)
		var req TailRequest
		if err := decoder.Decode(&req); err != nil {
			logger.Printf("bad tail request - error: %s", err)
			WriteJSONBadRequest(w, err)
			return
		}
		defer r.Body.Close()
		logger.Printf("tail request: %s", req.String())

		// validation
		if !validPrefix(req.Path, pathPrefixes) {
			logger.Printf("bad tail request - error: invalid path: %s", req.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// create a log file value
		var (
			err     error
			logFile cproject.LogFileReader
		)
		logFile, err = cproject.NewLogFile(req.Path)
		if err != nil {
			logger.Print(err)
			WriteJSONBadRequest(w, err)
			return
		}
		defer logFile.Close()

		// determine num lines to return
		numLines := req.NumLines
		if numLines == 0 {
			numLines = DefaultNumLines
		}

		// create filters if requested
		filters := []cproject.Filter{}
		if len(req.MatchSubstrings) > 0 {
			filters = append(filters,
				cproject.NewMatchAnySubstring(
					cproject.WithSubstrings(req.MatchSubstrings),
					cproject.WithCaseSensitivity(req.CaseSensitive)),
			)
		}

		// tail file
		start := time.Now()
		lineBytesOut := int64(0)
		lines, errChan := logFile.YieldLines(numLines, filters...)
		for line := range lines {
			lineBytesOut += int64(len([]byte(line)))
			chunk := TailResponseChunk{
				Host: host,
				Line: line,
			}
			WriteJSONCompact(w, &chunk)
		}

		// check for errors
		err = <-errChan
		if err != nil {
			logger.Print(err)
			return
		}
		logger.Printf("tail request - line bytes out written: %d [took %s]", lineBytesOut, time.Since(start))
	})
}
