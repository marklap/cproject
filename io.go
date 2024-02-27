package cproject

import (
	"bytes"
	"io"
	"os"
)

const (
	// stdBufSize is the number of bytes read at a time when reading the log file.
	stdBufSize int64 = 4096
	// newline is a newline
	newline byte = '\n'
)

// LineBuffer describes methods of interacting with a single line of a file.
type LineBuffer interface {
	// Offset returns the byte offset position of the start of the line in the file. A -1 indicates no offset has
	// been recorded (expected at initialization and after reset).
	Offset() int
	// Buffer returns the buffer that holds the line content.
	Buffer() *bytes.Buffer
	// Len returns the byte count of the content currently in the buffer.
	Len() int
	// Reset clears the buffer.
	Reset()
	// WriteByte writes a single byte to the buffer also indicating it's byte offset in the file.
	WriteByte(int, byte) error
	// Line returns the offset of the start of the line and the string value of the bytes in the buffer.
	Line() (int, string)
}

// TailLineBuffer stores bytes in reverse order of how they're found in the log file (i.e. end first).
type TailLineBuffer struct {
	offset int
	buf    *bytes.Buffer
}

// NewTailLineBuffer creates a new empty buffer.
func NewTailLineBuffer() *TailLineBuffer {
	return &TailLineBuffer{
		offset: -1,
		buf:    bytes.NewBuffer([]byte{}),
	}
}

// Returns the length of the buffer.
func (b *TailLineBuffer) Len() int {
	return b.buf.Len()
}

// Reset resets the buffer.
func (b *TailLineBuffer) Reset() {
	b.offset = -1
	b.buf.Reset()
}

// WriteByte writes a single byte to the buffer.
func (b *TailLineBuffer) WriteByte(offset int, c byte) error {
	b.offset = offset
	return b.buf.WriteByte(c)
}

// Line returns the offset of the start of the line and the string value of the bytes in the buffer.
func (b *TailLineBuffer) Line() (int, string) {
	bufLen := b.buf.Len()
	if bufLen == 0 {
		return -1, ""
	}
	if bufLen == 1 {
		return b.offset, b.buf.String()
	}

	buf := make([]byte, b.buf.Len())
	copy(buf, b.buf.Bytes())

	for i, j := 0, bufLen-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}

	return b.offset, string(buf)
}

// LineYielder describes the behavior of a value that reads a text file looking for lines that match a set of filters.
type LineYielder interface {
	// YieldLines yields (to a "lines" channel) a single line matching any of the provided filters up to
	// n number of lines.
	YieldLines(*os.File, int, []Filter)
	// SendLine yields a line from the file wrapped in a LineBuffer interface.
	SendLine(LineBuffer) error
	// Close indicates there no more lines will be yielded. An error may be included if the close action was triggered
	// by an error.
	Close(error)
	// LinesChan returns a channel for consuming the lines that are yielded.
	LinesChan() chan LineBuffer
	// ErrChan returns a channel for consuming an error that interrupted processing of the file. No further lines will
	// be yielded if an error is received on this channel and both the lines and error channel will be closed.
	ErrChan() chan error
}

// TailEndFirst is a LineYielder that reads lines starting from the end of the file reading toward the beginning
// of the file and returning matching lines in the order they are identified.
type TailEndFirst struct {
	lines chan LineBuffer
	err   chan error
}

// SendLine delivers a line to the lines channel.
func (t *TailEndFirst) SendLine(line LineBuffer) {
	t.lines <- line
}

// Lines returns a channel for communicating lines. Each line that matches a filter will be delivered via the lines
// channel. The lines and errors channels are always closed together.
func (t *TailEndFirst) LinesChan() chan LineBuffer {
	return t.lines
}

// Errors returns a channel used for communicating errors. Any error sent over the errors channel closes both the
// lines and errors channels.
func (t *TailEndFirst) ErrChan() chan error {
	return t.err
}

// Close sends an error over the error channel (if provided) and closes the lines and error channels.
func (t *TailEndFirst) Close(err error) {
	defer close(t.err)
	defer close(t.lines)
	if err != nil {
		t.err <- err
	}
}

// tailEndFirstOpt are options for customizing a TailEndFirst value.
type tailEndFirstOpt func(*TailEndFirst)

// WithLinesChannel creates a TailEndFirst LineYielder with the provided channel.
func WithLinesChannel(lines chan LineBuffer) tailEndFirstOpt {
	return func(t *TailEndFirst) {
		t.lines = lines
	}
}

// WithLinesChannel creates a TailEndFirst LineYielder with the provided channel.
func WithErrChannel(err chan error) tailEndFirstOpt {
	return func(t *TailEndFirst) {
		t.err = err
	}
}

// NewTailEndFirst creates and returns a TailEndFirst LineYielder. The default lines and err channels have a capacity
// of 1.
func NewTailEndFirst(opts ...tailEndFirstOpt) *TailEndFirst {
	t := &TailEndFirst{
		lines: make(chan LineBuffer, 1),
		err:   make(chan error, 1),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// startPos determines the best starting position of the seek depending on the size of the file.
// If the file size is less than or equal to the buffer size, we'll read the whole thing.
func startPos(bufSz int64, file *os.File) (int64, error) {
	stat, err := file.Stat()
	if err != nil {
		return -1, err
	}
	if stat.Size() <= bufSz {
		return 0, nil
	}
	return bufSz, nil
}

// countTrailingNewlines counts the number of newlines at the end of the buffer.
func countTrailingNewlines(buf []byte) int64 {
	var nlCount int64 = 0
	for i := len(buf) - 1; i >= 0; i-- {
		if buf[i] == newline {
			nlCount++
		} else {
			break
		}
	}
	return nlCount
}

// includeLine determines if the line should be included in the output. It expects
// a LineBuffer. and returns true and the string
// if it should be included. If it should not be included it will return false and the string.
func includeLine(lineBuf *LineBuffer, filters []Filter) (bool, string) {
	lineLen := lineBuf.Len()
	if lineLen == 0 {
		return false, ""
	}

	line := lineBuf.String()
	if len(filters) == 0 {
		return true, line
	}

	for _, filter := range filters {
		if filter.Include(line) {
			return true, line
		}
	}

	return false, line
}

// yieldLines reads up to `numLines` lines from the provided file. The file is read in reverse building
// lines as they are identified (line = SOF,\n; \n,\n; \n,EOF). If `numLines` is 0 or less, all lines are returned.
// If `filters` are provided, only lines that pass the filters are returned. When a line is identified and passes
// filters, it is yielded to the lines channel. If an error is encountered, processes stops and an error is returned
// on the `errChan`. When `yieldLines` is successful, both the lines channel will be closed with no further values.
//
// TODO: Dissect this function into smaller, managable functions.
func yieldLines(file *os.File, numLines int, filters []Filter, lines chan<- string, errChan chan<- error) {
	// ensure we rewind the pointer when we're done
	defer func() { file.Seek(0, io.SeekStart) }()

	var (
		// bufSz is the size of the read buffer.
		bufSz int64 = stdBufSize
		// buf is the read buffer.
		buf []byte = make([]byte, bufSz)
		// pos is the current seek position of the file.
		pos int64 = 0

		// nlCount collects the count of yielded lines.
		nlCount = 0
		// firstRead is a flag that, when true, indicates we haven't read any lines yet.
		firstRead = true
	)

	// Determine the best seek position to start reading from.
	pos, err := startPos(bufSz, file)
	if err != nil {
		close(lines)
		errChan <- err
		close(errChan)
		return
	}

	// If we're not reading from the beginning of the file (because the file is bigger than the buffer), then seek to
	// that position.
	if pos > 0 {
		pos, err = file.Seek(-pos, io.SeekEnd)
		if err != nil {
			close(lines)
			errChan <- err
			close(errChan)
			return
		}
	}

	// lineBuf is buffer that collects bytes from a single line in the file.
	lineBuf := NewTailLineBuffer()

	// Loop, reading chunks of the file and yielding lines as they are identified.
	for {
		// after this, the seek pos will be pos + sz
		sz, err := file.Read(buf)
		if err != nil {
			if err != io.EOF {
				close(lines)
				errChan <- err
				close(errChan)
			}
		}

		// Determine the best start and end index for reading the bytes in the buffer.
		start, end := sz-1, 0
		if firstRead {
			// adjust the start to account for trailing newlines at the end of the file.
			start -= int(countTrailingNewlines(buf))
			firstRead = false
		}

		// Everytime we come across a newline, check to see if we should yield it.
		for i := start; i >= end; i-- {
			if buf[i] == newline {
				if include, line := includeLine(lineBuf, filters); include {
					lines <- line
					nlCount++
					// If we've yielded the requested number of lines, we're done.
					if numLines > 0 && nlCount == numLines {
						close(lines)
						close(errChan)
						return
					}
				}
				// Reset the line buffer regardless if we've yielded the line.
				lineBuf.Reset()
			} else {
				// Append the read byte to the line buffer.
				err := lineBuf.WriteByte(buf[i])
				if err != nil {
					close(lines)
					errChan <- err
					close(errChan)
				}
			}
		}

		// If we've read less than a full buffer size then we've truncated the buffer
		// on the previous pass and reached the beginning of the file and we're done.
		if sz < int(stdBufSize) {
			if include, line := includeLine(lineBuf, filters); include {
				lines <- line
			}
			close(lines)
			close(errChan)
			return
		}

		// determine how far to move the pointer by seeing if there's enough bytes to fill the buffer
		// else seek to pos 0 and truncate the buffer
		offset := bufSz * 2 // rewind to beginning of current buffer plus another buffer
		err = nil
		if pos-offset > 0 {
			pos, err = file.Seek(-offset, io.SeekCurrent)
		} else {
			buf = make([]byte, pos)
			pos, err = file.Seek(0, io.SeekStart)
		}
		if err != nil {
			close(lines)
			errChan <- err
			close(errChan)
		}
	}
}
