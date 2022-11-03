package lit_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/nlandolfi/lit"
)

func TestGolden(t *testing.T) {
	var cases = []struct {
		file       string
		golden     string
		goldenHTML string
	}{
		{
			file:   "./examples/halmos/halmos.lit",
			golden: "./examples/halmos/halmos_golden.lit",
		},
		{
			file:   "./examples/kierkegaard/kierkegaard.lit",
			golden: "./examples/kierkegaard/kierkegaard_golden.lit",
		},
		{
			file:   "./examples/aristotle/aristotle.lit",
			golden: "./examples/aristotle/aristotle_golden.lit",
		},
		{
			file:   "./examples/mathgenomics/mathgenomics.lit",
			golden: "./examples/mathgenomics/mathgenomics_golden.lit",
		},
		{
			file:       "./examples/headerids/headerids.lit",
			golden:     "./examples/headerids/headerids_golden.lit",
			goldenHTML: "./examples/headerids/headerids_golden.html",
		},
	}

	for _, c := range cases {
		bs, err := os.ReadFile(c.file)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf(c.file)

		n, err := lit.ParseLit(string(bs))
		if err != nil {
			t.Fatal(err)
		}

		var b bytes.Buffer

		lit.WriteLit(&b, n, &lit.WriteOpts{Prefix: "", Indent: "  "})

		s := b.String()

		bs, err = os.ReadFile(c.golden)
		if err != nil {
			t.Fatal(err)
		}

		if want, got := string(bs), s; got != want {
			t.Fatalf("%q doesn't match golden\n diff the result of lit on %q \nagainst %q", c.file, c.file, c.golden)
		}

		if c.goldenHTML != "" {
			bs, err := os.ReadFile(c.goldenHTML)
			if err != nil {
				t.Fatal(err)
			}

			var b bytes.Buffer
			lit.WriteHTMLInBody(&b, n, &lit.WriteOpts{Prefix: "", Indent: "  "})

			if want, got := string(bs), b.String(); got != want {
				t.Fatalf("%q doesn't match goldenHTML\n diff the result of lit on %q \nagainst %q", c.file, c.file, c.goldenHTML)
			}

		}

	}
}
