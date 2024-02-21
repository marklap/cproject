package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/marklap/cproject"
)

const (
	DefaultNumLines = 10
)

type TailRequest struct {
	Path            string   `json:"path"`
	NumLines        int      `json:"num_lines"`
	MatchSubstrings []string `json:"match_substrings"`
	CaseSensitive   bool     `json:"case_sensitive"`
}

func (r *TailRequest) String() string {
	return fmt.Sprintf("path: %s, num_lines: %d, match_substrings: %s, case_sensitive: %t",
		r.Path, r.NumLines, r.MatchSubstrings, r.CaseSensitive)
}

type TailResponseChunk struct {
	Hostname string `json:"hostname"`
	Line     string `json:"line"`
}

func validPrefix(path string, rootPaths []string) bool {
	for _, rootPath := range rootPaths {
		if strings.HasPrefix(filepath.Clean(path), rootPath) {
			return true
		}
	}
	return false
}

func TailHandler(logger *log.Logger, hostname string, rootPaths []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var req TailRequest
		if err := decoder.Decode(&req); err != nil {
			logger.Printf("bad tail request - error: %s", err)
			WriteJSONBadRequest(w, err)
			return
		}

		// validation
		if !validPrefix(req.Path, rootPaths) {
			logger.Printf("bad tail request - error: invalid path: %s", req.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

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

		numLines := req.NumLines
		if numLines == 0 {
			numLines = DefaultNumLines
		}

		filters := []cproject.Filter{}
		if len(req.MatchSubstrings) > 0 {
			filters = append(filters,
				cproject.NewMatchAnySubstring(
					cproject.WithSubstrings(req.MatchSubstrings),
					cproject.WithCaseSensitivity(req.CaseSensitive)),
			)
		}

		logger.Printf("tail request: %s", req.String())

		lineBytesOut := int64(0)
		lines, errChan := logFile.YieldLines(numLines, filters)
		for line := range lines {
			lineBytesOut += int64(len([]byte(line)))
			chunk := TailResponseChunk{
				Hostname: hostname,
				Line:     line,
			}
			WriteJSONCompact(w, &chunk)
		}
		err = <-errChan
		if err != nil {
			logger.Print(err)
			return
		}
		logger.Printf("tail request - bytes written: %d", lineBytesOut)
	})
}
