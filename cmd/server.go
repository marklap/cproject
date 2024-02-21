// Command line interface for the application.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/marklap/cproject/handlers"
)

const (
	// ProjectName is the name of this project.
	ProjectName = "C-Project"

	// DefaultListenIP is the default IP address to listen on.
	DefaultListenIP = "0.0.0.0"

	// DefaultListenPort is the default port to listen on.
	DefaultListenPort = 8080

	// DefaultPathPrefixes is the default path to use for path validation.
	DefaultPathPrefixes = "/var/log"

	// PathPrefixesEnvVar is the environment variable that specifies the allowable path prefixes.
	PathPrefixesEnvVar = "CPROJECT_PATH_PREFIXES"
)

var (
	hostname         string
	listenIP         string
	listenPort       int
	pathPrefixesList string
	pathPrefixes     []string
)

var logger = log.Default()

func init() {
	hostname, _ = os.Hostname()

	flag.StringVar(&listenIP, "ip", DefaultListenIP, "IP address to listen on")
	flag.IntVar(&listenPort, "port", DefaultListenPort, "port to listen on")
	flag.StringVar(&pathPrefixesList, "prefixes", DefaultPathPrefixes,
		fmt.Sprintf("path prefixes to use for path validation [%q deliminted]", os.PathListSeparator))
	flag.Parse()

	if pathPrefixesList == "" {
		pathPrefixesList = DefaultPathPrefixes
	}
	pathPrefixes = strings.Split(pathPrefixesList, string(os.PathListSeparator))
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/ping", handlers.PingHandler(logger))
	mux.Handle("/tail", handlers.TailHandler(logger, fmt.Sprintf("%s:%d", hostname, listenPort), pathPrefixes))

	listenAddr := fmt.Sprintf("%s:%d", listenIP, listenPort)
	logger.Printf("%s API server is listening on %s...", ProjectName, listenAddr)
	logger.Printf(" - root paths: %v", pathPrefixesList)
	logger.Fatal(http.ListenAndServe(listenAddr, mux))
}
