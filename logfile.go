package cproject

import (
	"bufio"
	"fmt"
	"io"
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

	// Close closes the log file.
	Close() error
}

// LogFile works with logs in a very basic way. It's capable of reading the entire contents of the file and tailing
// `n` lines of the log file.
type LogFile struct {
	path   string
	reader io.ReadCloser
}

type logFileOpt func(*LogFile)

// WithReader is a LogFile option that applies the provided reader to the LogFile value.
func WithReader(reader io.ReadCloser) logFileOpt {
	return func(lf *LogFile) {
		lf.reader = reader
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

	if lf.reader != nil {
		return lf, nil
	}

	fh, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			return nil, fmt.Errorf("error opening file: %w", err)
		}
	}

	return &LogFile{
		path:   path,
		reader: fh,
	}, nil
}

// Path is the path to the log file being read.
func (f *LogFile) Path() string {
	return f.path
}

// ReadLines returns a string slice, one string per line in the log file.
func (l *LogFile) ReadLines() ([]string, error) {
	var content []string
	scanner := bufio.NewScanner(l.reader)
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

// Close closes the log file handle.
func (l *LogFile) Close() error {
	return l.reader.Close()
}

// FilteredLogFile does everything a LogFile can and can also filter lines if filters are provided.
type FilteredLogFile struct {
	*LogFile
	filters []Filter
}

type filteredLogFileOpt func(*FilteredLogFile)

// AddFilter is a FilteredLogFile option that appends the provided filter to the list of filters.
func AddFilter(f Filter) filteredLogFileOpt {
	return func(flf *FilteredLogFile) {
		flf.filters = append(flf.filters, f)
	}
}

// NewFilterLogFile creates a new FilteredLogFile given the provided LogFile and options.
func NewFilteredLogFile(logFile *LogFile, opts ...filteredLogFileOpt) *FilteredLogFile {
	flf := &FilteredLogFile{
		LogFile: logFile,
	}

	for _, opt := range opts {
		opt(flf)
	}

	return flf
}

// Filters returns the list of filters assigned to this FilteredLogFile.
func (l *FilteredLogFile) Filters() []Filter {
	return l.filters
}

// ReadLines returnes all lines in the log file that pass through all filters.
func (l *FilteredLogFile) ReadLines() ([]string, error) {
	allLines, err := l.LogFile.ReadLines()
	if err != nil {
		return nil, err
	}

	var content []string
	for _, line := range allLines {
		for _, filter := range l.Filters() {
			if filter.Include(line) {
				content = append(content, line)
			}
		}
	}

	return content, nil
}

// TailLines returns a string slice of the last `n` lines that pass through all filters.
// TODO: make this more efficient - this will suck for large files.
func (l *FilteredLogFile) TailLines(n int) ([]string, error) {
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
