package lit_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/nlandolfi/lit"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestEquationWrite(t *testing.T) {
	raw := `
    <equation id='eq:numtests'>
      ‖ T_H(x) = \begin{cases} ⦉

      ‖ 1 & \text{if } \num{H} = 1 \\ ⦉

      ‖ 1 + \num{H} S_H(x) & \text{otherwise} ⦉

      ‖ \end{cases} ⦉
    </equation>⦉
	`

	// Parse the littex
	n, err := lit.ParseLit(raw)
	if err != nil {
		t.Fatal(err)
	}
	var opts = lit.DefaultWriteOpts
	var out bytes.Buffer
	lit.WriteTex(&out, n, opts)

	dmp := diffmatchpatch.New()

	want := `\begin{equation}\label{eq:numtests}
T_H(x) = \begin{cases}
1 & \text{if } \num{H} = 1 \\
1 + \num{H} S_H(x) & \text{otherwise}
\end{cases}
\end{equation}`

	diffs := dmp.DiffMain(out.String(), want, false)

	//	log.Print(out.String())
	//	log.Print(want)
	if out.String() != want {
		fmt.Println(dmp.DiffPrettyText(diffs))
		t.Errorf("got %q, want %q", out.String(), want)
	}
}
