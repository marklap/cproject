package cproject

import (
	"io"
	"os"
	"syscall"
)

// Page size for the filesystem used as the read buffer size.
var SysPageSz = int64(syscall.Getpagesize())

// ChunkedReader reads chunks of a file until reaching the end of the file.
type ChunkedReader interface {
	// Next reads the next chunk of the file until reaching the end of the file in which case an io.EOF is returned
	// as the error value.
	Next() ([]byte, error)
	// Offset returns the current byte offset position where the next read will start from.
	Offset() int64
}

// ReadChunksFromEnd is a ChunkedReader that reads chunks from the end of the file to the beginning.
type ReadChunksFromEnd struct {
	file    *os.File
	fileSz  int64
	chunkSz int64
	offset  int64
	buf     []byte
}

// NewReadChunksFromEnd creates a new ReadChunksFromEnd reader.
func NewReadChunksFromEnd(file *os.File) (*ReadChunksFromEnd, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return &ReadChunksFromEnd{
		file:    file,
		fileSz:  stat.Size(),
		chunkSz: SysPageSz,
		offset:  stat.Size() - SysPageSz,
		buf:     make([]byte, SysPageSz),
	}, nil
}

// amendOffset ensures the offset and chunk size are valid for the current read.
func (r *ReadChunksFromEnd) amendOffset() error {
	if r.offset <= 0 {
		r.chunkSz += r.offset
		r.offset = 0
		return io.EOF
	}
	return nil
}

// Next reads the next chunk of the file from the end of the file to the beginning. When the beginning of the file is
// reached, io.EOF is returned as the error value.
func (r *ReadChunksFromEnd) Next() ([]byte, error) {

	eof := r.amendOffset()

	// fmt.Printf("before: %s\n", r)
	sz, err := r.file.ReadAt(r.buf[:r.chunkSz], r.offset)
	if err != nil {
		if err != io.EOF {
			return nil, err
		}
	}

	// we've read nothing and/or reached the beginning of the file
	if sz == 0 {
		return nil, io.EOF
	}

	r.offset -= r.chunkSz

	return r.buf[:sz], eof
}

// Offset returns the current byte offset position where the next read will start from.
func (r *ReadChunksFromEnd) Offset() int64 {
	return r.offset
}
