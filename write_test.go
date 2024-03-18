package lit_test

import (
	"bytes"
	"fmt"
	"log"
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

	want := `\begin{equation}\label{eq:numtests}
T_H(x) = \begin{cases}
1 & \text{if } \num{H} = 1 \\
1 + \num{H} S_H(x) & \text{otherwise}
\end{cases}
\end{equation}`

	//	log.Print(out.String())
	//	log.Print(want)
	if out.String() != want {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(out.String(), want, false)
		fmt.Println(dmp.DiffPrettyText(diffs))
		t.Errorf("got %q, want %q", out.String(), want)
	}
}

func TestEquationWriteAfterRunAddsNewLine(t *testing.T) {
	raw := `
¶ ⦊
  ‖ Given the statuses $x$ and a group $H ⊂ P$, the number of
    tests required to determine the status of every specimen in
    $H$ is ⦉
  <equation id='eq:numtests'>
    ‖ T_H(x) = \begin{cases} ⦉

    ‖ 1 & \text{if } \num{H} = 1 \\ ⦉

    ‖ 1 + \num{H} S_H(x) & \text{otherwise} ⦉

    ‖ \end{cases} ⦉
  </equation>
⦉ `
	n, err := lit.ParseLit(raw)
	if err != nil {
		t.Fatal(err)
	}
	var opts = lit.DefaultWriteOpts
	var out bytes.Buffer
	lit.WriteLit(&out, n, opts)

	want := `¶ ⦊
  ‖ Given the statuses $x$ and a group $H ⊂ P$, the number of
    tests required to determine the status of every specimen in
    $H$ is ⦉

  <equation id='eq:numtests'>
    ‖ T_H(x) = \begin{cases} ⦉

    ‖ 1 & \text{if } \num{H} = 1 \\ ⦉

    ‖ 1 + \num{H} S_H(x) & \text{otherwise} ⦉

    ‖ \end{cases} ⦉
  </equation>
⦉`
	if out.String() != want {
		log.Print("got: ", out.String())
		log.Print("want: ", want)
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(out.String(), want, false)
		fmt.Println(dmp.DiffPrettyText(diffs))
		t.Errorf("got %q, want %q", out.String(), want)
	}
}
