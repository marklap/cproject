package cproject

import (
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

func TestIncludeLine(t *testing.T) {
	testCases := []struct {
		desc     string
		buf      *LineBuffer
		filters  []Filter
		wantBool bool
		wantStr  string
	}{
		{
			desc:     "noFilters",
			buf:      NewLineBufferFromString("cba"),
			filters:  nil,
			wantBool: true,
			wantStr:  "abc",
		}, {
			desc:     "emptyBuffer",
			buf:      NewLineBuffer(),
			filters:  nil,
			wantBool: false,
			wantStr:  "",
		}, {
			desc: "singleFilterMatch",
			buf:  NewLineBufferFromString("cba"),
			filters: []Filter{
				NewMatchAnySubstring(WithSubstrings([]string{"abc"})),
			},
			wantBool: true,
			wantStr:  "abc",
		}, {
			desc: "singleFilterNoMatch",
			buf:  NewLineBufferFromString("cba"),
			filters: []Filter{
				NewMatchAnySubstring(WithSubstrings([]string{"gorilla"})),
			},
			wantBool: false,
			wantStr:  "abc",
		}, {
			desc: "multipleFilterMatch",
			buf:  NewLineBufferFromString("cba"),
			filters: []Filter{
				NewMatchAnySubstring(WithSubstrings([]string{"gorilla"})),
				NewMatchAnySubstring(WithSubstrings([]string{"abc"})),
			},
			wantBool: true,
			wantStr:  "abc",
		}, {
			desc: "multipleFilterNoMatch",
			buf:  NewLineBufferFromString("cba"),
			filters: []Filter{
				NewMatchAnySubstring(WithSubstrings([]string{"gorilla"})),
				NewMatchAnySubstring(WithSubstrings([]string{"monkey"})),
			},
			wantBool: false,
			wantStr:  "abc",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			gotBool, gotStr := includeLine(tC.buf, tC.filters)
			if tC.wantBool != gotBool {
				t.Errorf("unexpected include line boolean - want: %t, got: %t", tC.wantBool, gotBool)
			}
			if tC.wantStr != gotStr {
				t.Errorf("unexpected include line string - want: %s, got: %s", tC.wantStr, gotStr)
			}
		})
	}
}

func TestYieldLines(t *testing.T) {
	testCases := []struct {
		desc  string
		lines int
		want  []string
	}{
		{
			desc:  "lastOne",
			lines: 1,
			want: []string{
				"and off-by-1 errors.",
			},
		}, {
			desc:  "lastTwo",
			lines: 2,
			want: []string{
				"and off-by-1 errors.",
				"naming things,",
			},
		}, {
			desc:  "lastThree",
			lines: 3,
			want: []string{
				"and off-by-1 errors.",
				"naming things,",
				"cache invalidation,",
			},
		}, {
			desc:  "sameAsActual",
			lines: 4,
			want: []string{
				"and off-by-1 errors.",
				"naming things,",
				"cache invalidation,",
				"There are 2 hard problems in computer science:",
			},
		}, {
			desc:  "moreThanActual",
			lines: 10,
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

			lines := make(chan string, 1)
			errChan := make(chan error, 1)

			go yieldLines(logFile.file, tC.lines, nil, lines, errChan)

			var got []string

			for line := range lines {
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
		desc    string
		lines   int
		filters []Filter
		want    []string
	}{
		{
			desc:  "lastOneMoreMatches",
			lines: 1,
			filters: []Filter{
				NewMatchAnySubstring(WithSubstrings([]string{"cache", "thing"})),
			},
			want: []string{
				"naming things,",
			},
		}, {
			desc:  "lastTwoSameMatches",
			lines: 2,
			filters: []Filter{
				NewMatchAnySubstring(WithSubstrings([]string{"cache", "thing"})),
			},
			want: []string{
				"naming things,",
				"cache invalidation,",
			},
		}, {
			desc:  "lastThreeLessMatches",
			lines: 3,
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

			lines := make(chan string, 1)
			errChan := make(chan error, 1)

			go yieldLines(logFile.file, tC.lines, tC.filters, lines, errChan)

			var got []string

			for line := range lines {
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
