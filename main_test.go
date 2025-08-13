package lit_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/nlandolfi/lit"
)

func TestMainAPI(t *testing.T) {
	// Test parsing
	node, err := lit.ParseLit("hello world")
	if err != nil {
		t.Fatalf("ParseLit failed: %v", err)
	}
	if node == nil {
		t.Fatal("ParseLit returned nil node")
	}

	// Test writing to different formats
	var buf bytes.Buffer

	// Test Lit output
	buf.Reset()
	err = lit.WriteLit(&buf, node, nil)
	if err != nil {
		t.Fatalf("WriteLit failed: %v", err)
	}
	litOutput := buf.String()
	if !strings.Contains(litOutput, "hello") {
		t.Errorf("Lit output should contain 'hello', got: %s", litOutput)
	}

	// Test TeX output
	buf.Reset()
	lit.WriteTex(&buf, node, nil)
	texOutput := buf.String()
	if !strings.Contains(texOutput, "hello") {
		t.Errorf("TeX output should contain 'hello', got: %s", texOutput)
	}

	// Test HTML output
	buf.Reset()
	err = lit.WriteHTML(&buf, node, nil)
	if err != nil {
		t.Fatalf("WriteHTML failed: %v", err)
	}
	htmlOutput := buf.String()
	if !strings.Contains(htmlOutput, "hello") {
		t.Errorf("HTML output should contain 'hello', got: %s", htmlOutput)
	}

	// Test debug output
	buf.Reset()
	lit.WriteDebug(&buf, node, nil)
	debugOutput := buf.String()
	if !strings.Contains(debugOutput, "fragment") {
		t.Errorf("Debug output should contain 'fragment', got: %s", debugOutput)
	}
}

func TestMust(t *testing.T) {
	// Test successful case
	node, err := lit.ParseLit("test")
	if err != nil {
		t.Fatalf("ParseLit failed: %v", err)
	}

	result := lit.Must(node, nil)
	if result != node {
		t.Error("Must should return the node when no error")
	}

	// Test panic case
	defer func() {
		if r := recover(); r == nil {
			t.Error("Must should panic when error is provided")
		}
	}()
	lit.Must(nil, fmt.Errorf("test error"))
}

func TestWriteOpts(t *testing.T) {
	opts := &lit.WriteOpts{Prefix: "test", Indent: "  ", InMath: false}

	// Test InMath
	mathOpts := lit.InMath(opts)
	if !mathOpts.InMath {
		t.Error("InMath should set InMath to true")
	}
	if mathOpts.Prefix != "test" {
		t.Error("InMath should preserve other options")
	}

	// Test Indented
	indentedOpts := lit.Indented(opts)
	if indentedOpts.Prefix != "test  " {
		t.Errorf("Indented should add indent to prefix, got: %q", indentedOpts.Prefix)
	}

	// Test NoPrefix
	noPrefixOpts := lit.NoPrefix(opts)
	if noPrefixOpts.Prefix != "" {
		t.Error("NoPrefix should clear the prefix")
	}
	if noPrefixOpts.Indent != "  " {
		t.Error("NoPrefix should preserve indent")
	}
}
