package cproject_test

import (
	"testing"

	"github.com/marklap/cproject"
)

func TestMatchAnySubstring(t *testing.T) {
	fxtLine := "Leonardo, Donatello, Raphael and Michelangelo"
	testCases := []struct {
		desc   string
		filter cproject.Filter
		want   bool
	}{
		{
			desc:   "prefixCaseSensitive",
			filter: cproject.NewMatchAnySubstring(cproject.WithSubstrings([]string{"Leonardo"})),
			want:   true,
		},
		{
			desc:   "suffixCaseSensitive",
			filter: cproject.NewMatchAnySubstring(cproject.WithSubstrings([]string{"Michelangelo"})),
			want:   true,
		},
		{
			desc:   "midCaseSensitive",
			filter: cproject.NewMatchAnySubstring(cproject.WithSubstrings([]string{"ello, Raph"})),
			want:   true,
		},
		{
			desc: "prefixCaseInsensitive",
			filter: cproject.NewMatchAnySubstring(cproject.WithSubstrings([]string{"leonardo"}),
				cproject.WithCaseSensitivity(false)),
			want: true,
		},
		{
			desc: "suffixCaseInensitive",
			filter: cproject.NewMatchAnySubstring(
				cproject.WithSubstrings([]string{"michelangelo"}),
				cproject.WithCaseSensitivity(false),
			),
			want: true,
		},
		{
			desc: "midCaseInensitive",
			filter: cproject.NewMatchAnySubstring(
				cproject.WithSubstrings([]string{"ello, raph"}),
				cproject.WithCaseSensitivity(false),
			),
			want: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := tC.filter.Include(fxtLine)
			if tC.want != got {
				t.Errorf("failure to including line - want: %t, got: %t", tC.want, got)
			}
		})
	}
}
