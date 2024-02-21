// Test utilities.
package cproject

import (
	"io"
	"os"
	"testing"
)

// StringSlicesEqual returns true if two string slices have identical strings in the exact same order.
func StringSlicesEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

// Fixture for the content in a synthetic file or temporary file.
func FxtContent() string {
	return "There are 2 hard problems in computer science:\ncache invalidation,\nnaming things,\nand off-by-1 errors."
}

// Creates a temporary file with the contents of `content`.
func FxtFile(t *testing.T, content string) (*os.File, error) {
	fh, err := os.CreateTemp(os.TempDir(), PackageName)
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() { fh.Close(); os.Remove(fh.Name()) })

	_, err = fh.WriteString(content)
	if err != nil {
		return nil, err
	}

	_, err = fh.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return fh, nil
}

// Create a new `LogFile` with the given path and file handle.
func FxtLogFile(path string, file *os.File) (*LogFile, error) {
	return NewLogFile(path, WithFile(file))
}
