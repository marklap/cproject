package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

var SysPageSz = int64(syscall.Getpagesize())

type ChunkedReader interface {
	Next() ([]byte, error)
}

type ReadChunksFromEnd struct {
	file    *os.File
	fileSz  int64
	chunkSz int64
	offset  int64
	buf     []byte
}

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

func (r *ReadChunksFromEnd) amendOffset() error {
	if r.offset < 0 {
		r.chunkSz += r.offset
		r.offset = 0
		return io.EOF
	}
	return nil
}

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

func (r *ReadChunksFromEnd) String() string {
	return fmt.Sprintf("ReadChunksFromEnd{file: %s, chunkSz: %d, offset: %d, cap(buf): %d}",
		r.file.Name(), r.chunkSz, r.offset, cap(r.buf))
}

func readChunks(file *os.File) {
	rdr, err := NewReadChunksFromEnd(file)
	if err != nil {
		panic(err)
	}

	i := 1
	for {
		chunkBytes, err := rdr.Next()
		if err != nil && err != io.EOF {
			panic(err)
		}

		fmt.Printf("[%6d]\n%s\n\n", i, string(chunkBytes))
		// chunk := strings.TrimSpace(strings.ReplaceAll(string(chunkBytes), "\n", " "))
		// chunkParts := strings.Split(chunk, " ")
		// if len(chunkParts) > 2 {
		// 	fmt.Printf("[%6d] %s - %s\n", i, chunkParts[1], chunkParts[len(chunkParts)-2])
		// } else {
		// 	fmt.Printf("[%6d] %s\n", i, chunkParts[0])
		// }

		if err == io.EOF {
			return
		}

		i++

	}

	// buf := make([]byte, StdChunkSz)

	// off, err := file.Seek(-StdChunkSz, io.SeekEnd)
	// if err != nil {
	// 	panic(err)
	// }

	// // done := false
	// i := 1
	// for {
	// 	sz, err := file.ReadAt(buf, off)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	if sz == 0 {
	// 		break
	// 	}

	// 	chunk := strings.TrimSpace(strings.ReplaceAll(string(buf[:sz]), "\n", " "))
	// 	chunkParts := strings.Split(chunk, " ")
	// 	fmt.Printf("[%6d] %s - %s\n", i, chunkParts[1], chunkParts[len(chunkParts)-2])
	// 	i++

	// 	if int64(sz) == off {
	// 		return
	// 	}

	// 	if StdChunkSz > off {
	// 		buf = make([]byte, off)
	// 		off = 0
	// 		// done = true
	// 	} else {
	// 		off -= StdChunkSz
	// 	}
	// 	// file.Seek(-(ChunkSz * 3), io.SeekCurrent)
	// 	// nOff := off - ChunkSz
	// 	// if nOff < 0 {
	// 	// 	file.Seek(0, io.SeekStart)
	// 	// 	buf = make([]byte, off)
	// 	// } else {
	// 	// 	file.Seek(-nOff, io.SeekCurrent)
	// 	// }
	// }
}

func main() {
	fmt.Println("page size:", SysPageSz)
	_, here, _, _ := runtime.Caller(0)
	filePath := filepath.Join(filepath.Dir(filepath.Dir(here)), "testdata", "number_lines.txt")

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	readChunks(file)
}
