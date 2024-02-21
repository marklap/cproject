package cproject

import (
	"bufio"
	"fmt"
	"os"
)

// LogFileReader describes the behavior of a log file that will be read.
type LogFileReader interface {
	// Path is the path to the log file being read.
	Path() string

	// ReadLines returns a string slice, one string per line in the log file.
	ReadLines() ([]string, error)

	// TailLines returns a string slice of the last `n` lines of the log file; one string per line in the log file.
	TailLines(int) ([]string, error)

	// YieldLines returns a string channel and an error channel for streaming lines from a log file.
	YieldLines(int, []Filter) (chan string, chan error)

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

// ReadLines returns a string slice, one string per line in the log file.
func (l *LogFile) ReadLines() ([]string, error) {
	var content []string
	scanner := bufio.NewScanner(l.file)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		content = append(content, scanner.Text())
	}
	return content, nil
}

// TailLines returns a string slice of the last `n` lines in the log file.
// TODO: make this more efficient - this will suck for large files.
func (l *LogFile) TailLines(n int) ([]string, error) {
	content, err := l.ReadLines()
	if err != nil {
		return nil, err
	}

	lc := len(content)
	if n > lc {
		return content, nil
	}

	return content[lc-n:], nil
}

// YieldLines returns a string channel and an error channel for streaming lines from a log file.
func (l *LogFile) YieldLines(numLines int, filters []Filter) (chan string, chan error) {
	lines := make(chan string, 1)
	errChan := make(chan error, 1)

	go yieldLines(l.file, numLines, filters, lines, errChan)

	return lines, errChan
}

// Close closes the log file handle.
func (l *LogFile) Close() error {
	return l.file.Close()
}
