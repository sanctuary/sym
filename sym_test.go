package sym_test

import (
	"crypto/sha1"
	"fmt"
	"testing"

	"github.com/sanctuary/sym"
)

func TestParseFile(t *testing.T) {
	// Hash sums based on the output of DUMPSYM.EXE from the Psy-Q SDK, as
	// contained in https://github.com/diasurgical/scalpel, with the last line
	// removed and with line endings converted to UNIX format.
	golden := []struct {
		path string
		want string // SHA1 hash of output in Psy-Q format.
	}{
		// psx/symbols/jap_05291998.out
		{
			path: "testdata/DIABPSX_SLPS-01416.sym",
			want: "19f823986500a369f60e78406fada915a1d18aca",
		},
		// psx/symbols/pal_12121997.out
		{
			path: "testdata/DIABPSX_easy_as_pie.sym",
			want: "ef1e5d733560794b66cc5710d2e6211a24afab7c",
		},
	}
	for _, g := range golden {
		f, err := sym.ParseFile(g.path)
		if err != nil {
			t.Errorf("unable to parse %q; %v", g.path, err)
			continue
		}
		got := fmt.Sprintf("%040x", sha1.Sum([]byte(f.String())))
		if g.want != got {
			t.Errorf("%q: SHA1 hash mismatch; expected %v, got %v", g.path, g.want, got)
		}
	}
}
