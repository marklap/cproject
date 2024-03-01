package cproject

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCountTrailingNewlines(t *testing.T) {
	testCases := []struct {
		desc    string
		content []byte
		want    int64
	}{
		{
			desc:    "one",
			content: []byte("a\n"),
			want:    1,
		}, {
			desc:    "many",
			content: []byte("a\n\n\n\n"),
			want:    4,
		}, {
			desc:    "none",
			content: []byte("a"),
			want:    0,
		}, {
			desc:    "onlyPrefix",
			content: []byte("\n\n\na"),
			want:    0,
		}, {
			desc:    "onlyNewlines",
			content: []byte("\n\n\n\n"),
			want:    4,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := countTrailingNewlines(tC.content)
			if tC.want != got {
				t.Errorf("unexpected count of trailing newlines - want: %d, got: %d", tC.want, got)
			}
		})
	}
}

func TestStartPos(t *testing.T) {
	content := FxtContent()
	file, err := FxtFile(t, content)
	if err != nil {
		t.Error(err)
	}

	stat, err := file.Stat()
	if err != nil {
		t.Error(err)
	}

	logFile, err := FxtLogFile(file.Name(), file)
	if err != nil {
		t.Error(err)
	}

	testCases := []struct {
		desc  string
		bufSz int64
		want  int64
	}{
		{
			desc:  "fileSzGreater",
			bufSz: stat.Size() - 1,
			want:  stat.Size() - 1,
		}, {
			desc:  "fileSzEqual",
			bufSz: stat.Size(),
			want:  0,
		}, {
			desc:  "fileSzLessThan",
			bufSz: stat.Size() + 1,
			want:  0,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := startPos(tC.bufSz, logFile.file)
			if err != nil {
				t.Error(err)
			}
			if tC.want != got {
				t.Errorf("unexpected start position - want: %d, got: %d", tC.want, got)
			}
		})
	}
}

func TestYieldLines(t *testing.T) {
	testCases := []struct {
		desc     string
		numLines int
		want     []string
	}{
		{
			desc:     "lastOne",
			numLines: 1,
			want: []string{
				"and off-by-1 errors.",
			},
		}, {
			desc:     "lastTwo",
			numLines: 2,
			want: []string{
				"and off-by-1 errors.",
				"naming things,",
			},
		}, {
			desc:     "lastThree",
			numLines: 3,
			want: []string{
				"and off-by-1 errors.",
				"naming things,",
				"cache invalidation,",
			},
		}, {
			desc:     "sameAsActual",
			numLines: 4,
			want: []string{
				"and off-by-1 errors.",
				"naming things,",
				"cache invalidation,",
				"There are 2 hard problems in computer science:",
			},
		}, {
			desc:     "moreThanActual",
			numLines: 10,
			want: []string{
				"and off-by-1 errors.",
				"naming things,",
				"cache invalidation,",
				"There are 2 hard problems in computer science:",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			content := FxtContent()
			file, err := FxtFile(t, content)
			if err != nil {
				t.Error(err)
			}

			logFile, err := FxtLogFile(file.Name(), file)
			if err != nil {
				t.Error(err)
			}

			tail := NewTailEndFirst(logFile.file)

			linesChan := tail.LinesChan()
			errChan := tail.ErrChan()

			go tail.YieldLines(tC.numLines, nil)

			var got []string

			for line := range linesChan {
				got = append(got, line)
			}
			err = <-errChan
			if err != nil {
				t.Error(err)
			}

			if !StringSlicesEqual(tC.want, got) {
				t.Errorf("unexpected results - want: %#v, got: %#v", tC.want, got)
			}
		})
	}
}

func TestYieldLinesFiltered(t *testing.T) {
	testCases := []struct {
		desc     string
		numLines int
		filters  []Filter
		want     []string
	}{
		{
			desc:     "lastOneMoreMatches",
			numLines: 1,
			filters: []Filter{
				NewMatchAnySubstring(WithSubstrings([]string{"cache", "thing"})),
			},
			want: []string{
				"naming things,",
			},
		}, {
			desc:     "lastTwoSameMatches",
			numLines: 2,
			filters: []Filter{
				NewMatchAnySubstring(WithSubstrings([]string{"cache", "thing"})),
			},
			want: []string{
				"naming things,",
				"cache invalidation,",
			},
		}, {
			desc:     "lastThreeLessMatches",
			numLines: 3,
			filters: []Filter{
				NewMatchAnySubstring(WithSubstrings([]string{"cache", "thing"})),
			},
			want: []string{
				"naming things,",
				"cache invalidation,",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			content := FxtContent()
			file, err := FxtFile(t, content)
			if err != nil {
				t.Error(err)
			}

			logFile, err := FxtLogFile(file.Name(), file)
			if err != nil {
				t.Error(err)
			}

			tail := NewTailEndFirst(logFile.file)

			linesChan := tail.LinesChan()
			errChan := tail.ErrChan()

			go tail.YieldLines(tC.numLines, tC.filters)

			var got []string

			for line := range linesChan {
				got = append(got, line)
			}
			err = <-errChan
			if err != nil {
				t.Error(err)
			}

			if !StringSlicesEqual(tC.want, got) {
				t.Errorf("unexpected results - want: %#v, got: %#v", tC.want, got)
			}
		})
	}
}

func TestListDir(t *testing.T) {
	var thisFullPath string
	var ok bool
	if _, thisFullPath, _, ok = runtime.Caller(0); !ok {
		t.Error("failed to get test file path")
	}

	testDataDir := filepath.Join(filepath.Dir(thisFullPath), "testdata")

	wantFiles := []string{
		filepath.Join(testDataDir, "bartender.txt"),
		filepath.Join(testDataDir, "number_lines.txt"),
	}
	gotFiles, err := ListDir(testDataDir)
	if err != nil {
		t.Error(err)
	}

	for _, want := range wantFiles {
		found := false
		for _, got := range gotFiles {
			if want == got {
				found = true
				break
			}
		}
		if !found {
			t.Error(fmt.Errorf("file not returned in dir listing - want: %s, got: %s", want, gotFiles))
		}
	}
}

func NewTailLineBufferFromString(s string) LineBuffer {
	buf := NewTailLineBuffer()
	buf.buf.WriteString(s)
	return buf
}
