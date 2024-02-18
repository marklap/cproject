package cproject

import (
	"io"
	"os"
)

const (
	stdBufSize int64 = 4096
	newline    byte  = '\n'
)

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

func posNthLineFromEnd(file *os.File, numLines int, filters []Filter) (int64, error) {
	// ensure we rewind the pointer when we're done
	defer func() { file.Seek(0, io.SeekStart) }()

	var (
		bufSz int64  = stdBufSize
		buf   []byte = make([]byte, bufSz)
		pos   int64  = 0

		nlCount   = 0
		firstRead = true
	)

	pos, err := startPos(bufSz, file)
	if err != nil {
		return -1, err
	}
	if pos > 0 {
		pos, err = file.Seek(-pos, io.SeekEnd)
		if err != nil {
			return -1, err
		}
	}

	for {
		// after this, the seek pos will be pos + sz
		sz, err := file.Read(buf)
		if err != nil {
			return -1, err
		}

		start, end := sz-1, 0
		if firstRead {
			start -= int(countTrailingNewlines(buf))
			firstRead = false
		}

		for i := start; i >= end; i-- {
			if buf[i] == newline {
				nlCount++
				if nlCount == numLines {
					return pos + int64(i) + 1, nil
				}
			}
		}

		// if we've read less than a full buffer size then we've truncated the buffer
		// on the previous pass and reached the beginning of the file and we haven't counted
		// the number of lines we're targetting so we should just dump the whole file (i.e. seek to pos 0)
		if sz < int(stdBufSize) {
			return 0, nil
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
			return -1, err
		}
	}
}
