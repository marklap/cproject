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

// TailHandler handles requests to tail a log file.
func TailHandler(logger *log.Logger, host string, pathPrefixes []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// validate request and transform to tail config
		cfg, err := tailCfgFromRequest(r, pathPrefixes)
		if err != nil {
			logger.Printf("bad tail request - error: %s", err)
			WriteJSONBadRequest(w, err)
			return
		}
		logger.Printf("tail request - %s", cfg)

		// create a log file wrapper
		logFile, err := cproject.NewLogFile(cfg.path)
		if err != nil {
			logger.Print(err)
			WriteJSONBadRequest(w, err)
			return
		}
		defer logFile.Close()

		// initialize stats
		start := time.Now()
		lineBytesOut := int64(0)

		// tail log file
		lines, errChan := logFile.YieldLines(cfg.numLines, cfg.filters...)
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

type tailCfg struct {
	path          string
	numLines      int
	filters       []cproject.Filter
	caseSensitive bool
}

func (c tailCfg) String() string {
	return fmt.Sprintf("path: %s, numLines: %d, filters(count): %d, caseSensitive: %t",
		c.path, c.numLines, len(c.filters), c.caseSensitive)
}

func tailCfgFromRequest(r *http.Request, pathPrefixes []string) (*tailCfg, error) {
	req, err := decodeRequest(r)
	if err != nil {
		return nil, err
	}

	if !isValidPath(req.Path, pathPrefixes) {
		return nil, fmt.Errorf("invalid path: %s", req.Path)
	}

	return &tailCfg{
		path:          req.Path,
		numLines:      numLinesFromRequest(req),
		filters:       filtersFromRequest(req),
		caseSensitive: req.CaseSensitive,
	}, nil
}

func decodeRequest(r *http.Request) (TailRequest, error) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	var req TailRequest
	if err := decoder.Decode(&req); err != nil {
		return req, err
	}
	return req, nil
}

func isValidPath(path string, pathPrefixes []string) bool {
	// only valid path prefixes are allowed
	for _, pathPrefix := range pathPrefixes {
		if strings.HasPrefix(filepath.Clean(path), pathPrefix) {
			return true
		}
	}
	return false
}

func numLinesFromRequest(req TailRequest) int {
	n := req.NumLines
	if n == 0 {
		n = DefaultNumLines
	}
	return n
}

func filtersFromRequest(req TailRequest) []cproject.Filter {
	filters := []cproject.Filter{}
	if len(req.MatchSubstrings) > 0 {
		filters = append(filters,
			cproject.NewMatchAnySubstring(
				cproject.WithSubstrings(req.MatchSubstrings),
				cproject.WithCaseSensitivity(req.CaseSensitive)),
		)
	}
	return filters
}
