package cproject_test

import (
	"errors"
	"os"
	"reflect"
	"strings"
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
			path:          "./testdata/the_wind_and_the_sun.txt",
			wantPath:      "./testdata/the_wind_and_the_sun.txt",
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

func TestLogFileReadLines(t *testing.T) {
	file, err := cproject.FxtFile(t, cproject.FxtContent())
	if err != nil {
		t.Error(err)
	}
	logFile, err := cproject.FxtLogFile(file.Name(), file)
	if err != nil {
		t.Error(err)
	}

	want := strings.Split(cproject.FxtContent(), "\n")
	got, err := logFile.ReadLines()
	if err != nil {
		t.Error(err)
	}

	if !cproject.StringSlicesEqual(want, got) {
		t.Errorf("unexpected content - want: %#v, got: %#v", want, got)
	}
}

func TestLogFileTailLines(t *testing.T) {
	file, err := cproject.FxtFile(t, cproject.FxtContent())
	if err != nil {
		t.Error(err)
	}
	logFile, err := cproject.FxtLogFile(file.Name(), file)
	if err != nil {
		t.Error(err)
	}

	content := strings.Split(cproject.FxtContent(), "\n")
	want := content[len(content)-2:]
	got, err := logFile.TailLines(2)
	if err != nil {
		t.Error(err)
	}

	if !cproject.StringSlicesEqual(want, got) {
		t.Errorf("unexpected content - want: %#v, got: %#v", want, got)
	}
}

func TestNewFilteredLogFile(t *testing.T) {
	fxtLogFile, err := cproject.NewLogFile("./testdata/the_wind_and_the_sun.txt")
	if err != nil {
		t.Errorf("error creating log file fixture: %s", err)
	}

	testCases := []struct {
		desc            string
		logFile         *cproject.LogFile
		addFilter       cproject.Filter
		wantLogFilePath string
		wantFilters     []cproject.Filter
	}{
		{
			desc:            "addFilter",
			logFile:         fxtLogFile,
			addFilter:       &cproject.MockFilter{},
			wantLogFilePath: "./testdata/the_wind_and_the_sun.txt",
			wantFilters: []cproject.Filter{
				&cproject.MockFilter{},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := cproject.NewFilteredLogFile(tC.logFile, cproject.AddFilter(tC.addFilter))
			if tC.wantLogFilePath != got.Path() {
				t.Errorf("unexpected log file path - want: %s, got: %s", tC.wantLogFilePath, got.Path())
			}
			if !reflect.DeepEqual(tC.wantFilters, got.Filters()) {
				t.Errorf("unexpected filter members - want: %v, got: %v", tC.wantFilters, got.Filters())
			}
		})
	}
}

func TestFilteredLogFileReadLines(t *testing.T) {
	file, err := cproject.FxtFile(t, cproject.FxtContent())
	if err != nil {
		t.Error(err)
	}
	logFile, err := cproject.FxtLogFile(file.Name(), file)
	if err != nil {
		t.Error(err)
	}

	filtLogFile := cproject.FxtFilteredLogFile(logFile, []string{"things", "cache"})
	want := []string{"cache invalidation,", "naming things,"}
	got, err := filtLogFile.ReadLines()
	if err != nil {
		t.Error(err)
	}

	if !cproject.StringSlicesEqual(want, got) {
		t.Errorf("unexpected content - want: %#v, got: %#v", want, got)
	}
}

func TestFilteredLogFileTailLines(t *testing.T) {
	file, err := cproject.FxtFile(t, cproject.FxtContent())
	if err != nil {
		t.Error(err)
	}
	logFile, err := cproject.FxtLogFile(file.Name(), file)
	if err != nil {
		t.Error(err)
	}

	filtLogFile := cproject.FxtFilteredLogFile(logFile, []string{"things", "cache"})
	want := []string{"naming things,"}
	got, err := filtLogFile.TailLines(1)
	if err != nil {
		t.Error(err)
	}

	if !cproject.StringSlicesEqual(want, got) {
		t.Errorf("unexpected content - want: %#v, got: %#v", want, got)
	}
}
