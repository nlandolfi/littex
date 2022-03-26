package gba_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/greatbooksadventure/gba"
)

func TestGolden(t *testing.T) {
	var cases = []struct {
		file   string
		golden string
	}{
		{
			file:   "./cmd/parse3/halmos.gba",
			golden: "./cmd/parse3/halmos_golden.gba",
		},
		{
			file:   "./cmd/parse3/kierkegaard.gba",
			golden: "./cmd/parse3/kierkegaard_golden.gba",
		},
		{
			file:   "./cmd/parse3/aristotle.gba",
			golden: "./cmd/parse3/aristotle_golden.gba",
		},
		{
			file:   "./cmd/parse3/mathgenomics.gba",
			golden: "./cmd/parse3/mathgenomics_golden.gba",
		},
	}

	for _, c := range cases {
		bs, err := os.ReadFile(c.file)
		if err != nil {
			t.Fatal(err)
		}

		n, err := gba.Parse3(string(bs))
		if err != nil {
			t.Fatal(err)
		}

		var b bytes.Buffer

		gba.WriteGBA(&b, n, "", "  ")

		s := b.String()

		bs, err = os.ReadFile(c.golden)
		if err != nil {
			t.Fatal(err)
		}

		if want, got := string(bs), s; got != want {
			t.Fatalf("%q doesn't match golden\n diff the result of gba3 on %q \nagainst %q", c.file, c.file, c.golden)
		}

	}
}
