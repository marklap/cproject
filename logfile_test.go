package cproject_test

import (
	"errors"
	"os"
	"testing"

	"github.com/marklap/cproject"
)

func TestNewLogFile(t *testing.T) {
	testCases := []struct {
		desc          string
		path          string
		wantPath      string
		errIs         func(error) bool
		wantErrIsDesc string
	}{
		{
			desc:          "logFileExists",
			path:          "./testdata/bartender.txt",
			wantPath:      "./testdata/bartender.txt",
			errIs:         func(err error) bool { return err == nil },
			wantErrIsDesc: "error is nil",
		}, {
			desc:          "logFileNotExists",
			path:          "this-file-does-not-exist.missing",
			errIs:         func(err error) bool { return os.IsNotExist(errors.Unwrap(err)) },
			wantErrIsDesc: "error is not not-exist",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := cproject.NewLogFile(tC.path)
			if err == nil {
				if tC.wantPath != got.Path() {
					t.Errorf("unexpected path - want: %s, got: %s", tC.wantPath, got.Path())
				}
			} else if !tC.errIs(err) {
				t.Errorf("unexpected error - want: %s, got: %s", tC.wantErrIsDesc, err)
			}
		})
	}
}

func TestLogFileYieldLines(t *testing.T) {
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
			content := cproject.FxtContent()
			file, err := cproject.FxtFile(t, content)
			if err != nil {
				t.Error(err)
			}

			logFile, err := cproject.FxtLogFile(file.Name(), file)
			if err != nil {
				t.Error(err)
			}

			lines, errChan := logFile.YieldLines(tC.lines)

			var got []string

			for line := range lines {
				got = append(got, line)
			}
			err = <-errChan
			if err != nil {
				t.Error(err)
			}

			if !cproject.StringSlicesEqual(tC.want, got) {
				t.Errorf("unexpected results - want: %#v, got: %#v", tC.want, got)
			}
		})
	}
}

func TestLogFileYieldLinesFiltered(t *testing.T) {
	testCases := []struct {
		desc    string
		lines   int
		filters []cproject.Filter
		want    []string
	}{
		{
			desc:  "lastOneMoreMatches",
			lines: 1,
			filters: []cproject.Filter{
				cproject.NewMatchAnySubstring(cproject.WithSubstrings([]string{"cache", "thing"})),
			},
			want: []string{
				"naming things,",
			},
		}, {
			desc:  "lastTwoSameMatches",
			lines: 2,
			filters: []cproject.Filter{
				cproject.NewMatchAnySubstring(cproject.WithSubstrings([]string{"cache", "thing"})),
			},
			want: []string{
				"naming things,",
				"cache invalidation,",
			},
		}, {
			desc:  "lastThreeLessMatches",
			lines: 3,
			filters: []cproject.Filter{
				cproject.NewMatchAnySubstring(cproject.WithSubstrings([]string{"cache", "thing"})),
			},
			want: []string{
				"naming things,",
				"cache invalidation,",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			content := cproject.FxtContent()
			file, err := cproject.FxtFile(t, content)
			if err != nil {
				t.Error(err)
			}

			logFile, err := cproject.FxtLogFile(file.Name(), file)
			if err != nil {
				t.Error(err)
			}

			lines, errChan := logFile.YieldLines(tC.lines, tC.filters...)

			var got []string

			for line := range lines {
				got = append(got, line)
			}
			err = <-errChan
			if err != nil {
				t.Error(err)
			}

			if !cproject.StringSlicesEqual(tC.want, got) {
				t.Errorf("unexpected results - want: %#v, got: %#v", tC.want, got)
			}
		})
	}
}

func TestClose(t *testing.T) {
	content := cproject.FxtContent()
	file, err := cproject.FxtFile(t, content)
	if err != nil {
		t.Error(err)
	}

	logFile, err := cproject.FxtLogFile(file.Name(), file)
	if err != nil {
		t.Error(err)
	}

	err = logFile.Close()
	if err != nil {
		t.Errorf("unsuccessful close of file - expected success")
	}
	err = logFile.Close()
	if err == nil {
		t.Errorf("no error returned - expected error on second attempt at close")
	}

}
