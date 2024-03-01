package cproject_test

import (
	"io"
	"testing"

	"github.com/marklap/cproject"
)

func TestNext(t *testing.T) {
	type chunkSpec struct {
		len int
		eof bool
	}

	sysPageSz := int(cproject.SysPageSz)
	halfSysPageSz := int(sysPageSz / 2)
	twiceSysPageSz := sysPageSz * 2

	testCases := []struct {
		desc       string
		contentLen int
		want       []chunkSpec
	}{
		{
			desc:       "lessThanPageSize",
			contentLen: halfSysPageSz,
			want: []chunkSpec{
				{len: halfSysPageSz, eof: true},
			},
		}, {
			desc:       "twicePageSize",
			contentLen: twiceSysPageSz,
			want: []chunkSpec{
				{len: sysPageSz, eof: false},
				{len: sysPageSz, eof: true},
			},
		}, {
			desc:       "moreThanTwicePageSize",
			contentLen: twiceSysPageSz + 10,
			want: []chunkSpec{
				{len: sysPageSz, eof: false},
				{len: sysPageSz, eof: false},
				{len: 10, eof: true},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			file, err := cproject.FxtFile(t, cproject.FxtRandomContent(tC.contentLen))
			if err != nil {
				t.Error(err)
			}

			rdr, err := cproject.NewReadChunksFromEnd(file)
			if err != nil {
				t.Error(err)
			}

			for _, want := range tC.want {
				chunk, err := rdr.Next()
				if want.eof && err != io.EOF {
					t.Error("expected io.EOF")
					return
				}
				if len(chunk) != want.len {
					t.Errorf("unexpected length of chunk - want: %d, got: %d", want.len, len(chunk))
				}

			}
		})
	}
}
