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

func TestPosNthLineFromEnd(t *testing.T) {
	content := FxtContent()
	file, err := FxtFile(t, content)
	if err != nil {
		t.Error(err)
	}

	logFile, err := FxtLogFile(file.Name(), file)
	if err != nil {
		t.Error(err)
	}

	testCases := []struct {
		desc  string
		lines int
		want  int64
	}{
		{
			desc:  "lastOne",
			lines: 1,
			want:  82,
		}, {
			desc:  "lastTwo",
			lines: 2,
			want:  67,
		}, {
			desc:  "lastThree",
			lines: 3,
			want:  47,
		}, {
			desc:  "sameAsActual",
			lines: 4,
			want:  0,
		}, {
			desc:  "moreThanActual",
			lines: 10,
			want:  0,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := posNthLineFromEnd(logFile.file, tC.lines, nil)
			if err != nil {
				t.Error(err)
			}
			if tC.want != got {
				t.Errorf("unexpected position for count of lines - want: %d, got: %d", tC.want, got)
			}
		})
	}
}
