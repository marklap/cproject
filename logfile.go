package cproject

import (
	"fmt"
	"os"
)

// LogFileReader describes the behavior of a log file that will be read.
type LogFileReader interface {
	// Path is the path to the log file being read.
	Path() string

	// YieldLines returns a string channel and an error channel for streaming lines from a log file.
	YieldLines(int, ...Filter) (chan string, chan error)

	// Close closes the log file.
	Close() error
}

// LogFile works with logs in a very basic way. It's capable of reading the entire contents of the file and tailing
// `n` lines of the log file.
type LogFile struct {
	path string
	file *os.File
}

type logFileOpt func(*LogFile)

// WithFile is a LogFile option that applies the provided reader to the LogFile value.
func WithFile(file *os.File) logFileOpt {
	return func(lf *LogFile) {
		lf.file = file
	}
}

// NewLogFile creates a new LogFile and applies the provided options.
func NewLogFile(path string, opts ...logFileOpt) (*LogFile, error) {
	lf := &LogFile{
		path: path,
	}

	for _, opt := range opts {
		opt(lf)
	}

	if lf.file != nil {
		return lf, nil
	}

	fh, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			return nil, fmt.Errorf("error opening file: %w", err)
		}
	}

	return &LogFile{
		path: path,
		file: fh,
	}, nil
}

// Path is the path to the log file being read.
func (f *LogFile) Path() string {
	return f.path
}

// YieldLines returns a string channel and an error channel for streaming lines from a log file.
func (l *LogFile) YieldLines(numLines int, filters ...Filter) (chan string, chan error) {
	lines := make(chan string, 1)
	errChan := make(chan error, 1)

	go yieldLines(l.file, numLines, filters, lines, errChan)

	return lines, errChan
}

// Close closes the log file handle.
func (l *LogFile) Close() error {
	return l.file.Close()
}
