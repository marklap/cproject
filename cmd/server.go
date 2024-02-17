package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/marklap/cproject"
)

const (
	ProjectName       = "C-Project"
	DefaultListenIP   = "0.0.0.0"
	DefaultListenPort = 8080
)

var hostname string
var logBuf bytes.Buffer
var logger = log.New(&logBuf, ProjectName, log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)

func init() {
	hostname, _ = os.Hostname()
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
}

type ReadLinesRequest struct {
	Path            string   `json:"path"`
	MatchSubstrings []string `json:"match_substrings"`
	CaseSensitive   bool     `json:"case_sensitive"`
}

type ReadLinesResponse struct {
	Hostname string `json:"hostname"`
	Content  string `json:"content"`
}

func ReadLinesHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req ReadLinesRequest
	if err := decoder.Decode(&req); err != nil {
		logger.Printf("bad /readlines request - error: %s", err)
		WriteJSONBadRequest(w, err)
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
	if len(req.MatchSubstrings) > 0 {
		filter := cproject.NewMatchAnySubstring(cproject.WithSubstrings(req.MatchSubstrings),
			cproject.WithCaseSensitivity(req.CaseSensitive))
		logFile = cproject.NewFilteredLogFile(logFile.(*cproject.LogFile), cproject.AddFilter(filter))
	}

	content, err := logFile.ReadLines()
	if err != nil {
		logger.Print(err)
		WriteJSONServerError(w, err)
		return
	}

	resp := ReadLinesResponse{
		Hostname: hostname,
		Content:  strings.Join(content, "\n"),
	}
	WriteJSON(w, &resp)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", PingHandler)
	mux.HandleFunc("/readlines", ReadLinesHandler)

	listenAddr := fmt.Sprintf("%s:%d", DefaultListenIP, DefaultListenPort)
	logger.Printf("%s API server is listening on %s...", ProjectName, listenAddr)
	logger.Fatal(http.ListenAndServe(listenAddr, mux))
}
