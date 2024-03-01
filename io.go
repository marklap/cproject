package cproject

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const newline byte = '\n'

// LineBuffer describes methods of interacting with a single line of a file.
type LineBuffer interface {
	// Offset returns the byte offset position of the start of the line in the file. A -1 indicates no offset has
	// been recorded (expected at initialization and after reset).
	Offset() int64
	// Buffer returns the buffer that holds the line content.
	Buffer() *bytes.Buffer
	// Len returns the byte count of the content currently in the buffer.
	Len() int64
	// Reset clears the buffer.
	Reset()
	// AppendByte writes a single byte to the buffer also indicating it's byte offset in the file.
	AppendByte(int64, byte) error
	// Line returns the offset of the start of the line and the string value of the bytes in the buffer.
	Line() (int64, string)
	// String returns the string value held in the buffer
	String() string
}

// TailLineBuffer stores bytes in reverse order of how they're found in the log file (i.e. end first).
type TailLineBuffer struct {
	offset int64
	buf    *bytes.Buffer
}

// NewTailLineBuffer creates a new empty buffer.
func NewTailLineBuffer() *TailLineBuffer {
	return &TailLineBuffer{
		offset: -1,
		buf:    bytes.NewBuffer([]byte{}),
	}
}

// Offset returns the byte offset position of the start of this line.
func (b *TailLineBuffer) Offset() int64 {
	return b.offset
}

// Buffer returns the buffer that holds the line content.
func (b *TailLineBuffer) Buffer() *bytes.Buffer {
	return b.buf
}

// Returns the length of the buffer.
func (b *TailLineBuffer) Len() int64 {
	return int64(b.buf.Len())
}

// Reset resets the buffer.
func (b *TailLineBuffer) Reset() {
	b.offset = -1
	b.buf.Reset()
}

// AppendByte writes a single byte to the buffer.
func (b *TailLineBuffer) AppendByte(offset int64, c byte) error {
	if err := b.buf.WriteByte(c); err != nil {
		return err
	}
	b.offset = offset
	return nil
}

// Line returns the offset of the start of the line and the string value of the bytes in the buffer.
func (b *TailLineBuffer) Line() (int64, string) {
	bufLen := b.buf.Len()
	if bufLen == 0 {
		return -1, ""
	}
	return b.offset, b.String()
}

// String returns the string value held in the buffer.
func (b *TailLineBuffer) String() string {
	bufLen := b.buf.Len()
	if bufLen == 0 {
		return ""
	}
	if bufLen == 1 {
		return b.buf.String()
	}

	buf := make([]byte, b.buf.Len())
	copy(buf, b.buf.Bytes())

	for i, j := 0, bufLen-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}

	return strings.TrimSpace(string(buf))
}

// LineYielder describes the behavior of a value that reads a text file looking for lines that match a set of filters.
type LineYielder interface {
	// File returnes the file that is being read.
	File() *os.File
	// YieldLines yields (to a "lines" channel) a single line matching any of the provided filters up to
	// n number of lines.
	YieldLines(int, []Filter)
	// SendLine yields a line from the file wrapped in a LineBuffer interface.
	SendLine(LineBuffer) error
	// Close indicates there no more lines will be yielded. An error may be included if the close action was triggered
	// by an error.
	Close(error)
	// LinesChan returns a channel for consuming the lines that are yielded.
	LinesChan() chan string
	// ErrChan returns a channel for consuming an error that interrupted processing of the file. No further lines will
	// be yielded if an error is received on this channel and both the lines and error channel will be closed.
	ErrChan() chan error
}

// TailEndFirst is a LineYielder that reads lines starting from the end of the file reading toward the beginning
// of the file and returning matching lines in the order they are identified.
type TailEndFirst struct {
	file      *os.File
	linesChan chan string
	errChan   chan error
	sentLines int
}

// File returns the file that is being read.
func (t *TailEndFirst) File() *os.File {
	return t.file
}

// SendLine delivers a line to the lines channel.
func (t *TailEndFirst) SendLine(line string) {
	t.linesChan <- line
	t.sentLines++
}

// Lines returns a channel for communicating lines. Each line that matches a filter will be delivered via the lines
// channel. The lines and errors channels are always closed together.
func (t *TailEndFirst) LinesChan() chan string {
	return t.linesChan
}

// Errors returns a channel used for communicating errors. Any error sent over the errors channel closes both the
// lines and errors channels.
func (t *TailEndFirst) ErrChan() chan error {
	return t.errChan
}

// Close sends an error over the error channel (if provided) and closes the lines and error channels.
func (t *TailEndFirst) Close(err error) {
	defer func() { close(t.errChan) }()
	defer func() { close(t.linesChan) }()
	if err != nil {
		t.errChan <- err
	}
}

func (t *TailEndFirst) yieldIfIncluded(lineBuf LineBuffer, filters []Filter) {
	lineLen := lineBuf.Len()

	if lineLen == 0 {
		return
	}

	line := lineBuf.String()
	if len(filters) == 0 {
		t.SendLine(line)
		return
	}

	for _, filter := range filters {
		if filter.Include(line) {
			t.SendLine(line)
			continue
		}
	}
}

// YieldLines yields into the lines channel a single line matching any of the provided filters up to
// numLines number of lines.
func (t *TailEndFirst) YieldLines(numLines int, filters []Filter) {
	defer func() { t.File().Seek(0, io.SeekStart) }()

	rdr, err := NewReadChunksFromEnd(t.File())
	if err != nil {
		t.Close(err)
		return
	}

	lineBuf := NewTailLineBuffer()

	for {
		chunk, readErr := rdr.Next()
		if readErr != nil && readErr != io.EOF {
			t.Close(readErr)
			return
		}

		for i := len(chunk) - 1; i >= 0; i-- {
			if chunk[i] == newline {
				t.yieldIfIncluded(lineBuf, filters)
				if numLines > 0 && t.sentLines == numLines {
					t.Close(nil)
					return
				}
				lineBuf.Reset()
			} else {
				if err := lineBuf.AppendByte(rdr.Offset()+int64(i), chunk[i]); err != nil {
					t.Close(err)
					return
				}
			}
		}

		if readErr == io.EOF {
			t.yieldIfIncluded(lineBuf, filters)
			t.Close(nil)
			return
		}
	}

}

// tailEndFirstOpt are options for customizing a TailEndFirst value.
type tailEndFirstOpt func(*TailEndFirst)

// WithLinesChannel sets the lines chanel to the provided channel.
func WithLinesChannel(linesChan chan string) tailEndFirstOpt {
	return func(t *TailEndFirst) {
		t.linesChan = linesChan
	}
}

// WithLinesChannel sets the error channel to the provided channel.
func WithErrChannel(errChan chan error) tailEndFirstOpt {
	return func(t *TailEndFirst) {
		t.errChan = errChan
	}
}

func NewTailEndFirst(file *os.File, opts ...tailEndFirstOpt) *TailEndFirst {
	t := &TailEndFirst{
		file:      file,
		linesChan: make(chan string, 1),
		errChan:   make(chan error, 1),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
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

// ListDir lists all regular files in a pathPrefix.
func ListDir(pathPrefix string) ([]string, error) {
	var files []string

	err := fs.WalkDir(os.DirFS(pathPrefix), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Type().IsRegular() {
			files = append(files, filepath.Join(pathPrefix, path))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
