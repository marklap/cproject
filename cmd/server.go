package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/marklap/cproject/handlers"
)

const (
	ProjectName       = "C-Project"
	DefaultListenIP   = "0.0.0.0"
	DefaultListenPort = 8080
	DefaultRootPath   = "/var/log"

	ROOTPATHS_ENV_VAR = "CPROJECT_ROOTPATHS"
)

var hostname string
var rootPathsList string
var rootPaths []string

var logger = log.Default()

func init() {
	hostname, _ = os.Hostname()
	rootPathsList = os.Getenv(ROOTPATHS_ENV_VAR)
	if rootPathsList == "" {
		rootPathsList = DefaultRootPath
	}
	rootPaths = strings.Split(rootPathsList, string(os.PathListSeparator))
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/ping", handlers.PingHandler(logger))
	mux.Handle("/tail", handlers.TailHandler(logger, hostname, rootPaths))

	listenAddr := fmt.Sprintf("%s:%d", DefaultListenIP, DefaultListenPort)
	logger.Printf("%s API server is listening on %s...", ProjectName, listenAddr)
	logger.Printf(" - root paths: %v", rootPathsList)
	logger.Fatal(http.ListenAndServe(listenAddr, mux))
}
