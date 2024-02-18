package cproject

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
