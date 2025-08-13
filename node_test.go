package lit

import (
	"testing"

	"github.com/nlandolfi/lit/internal/ast"
)

func TestSetAttrReplace(t *testing.T) {
	n := &Node{}
	n.Attr = []ast.Attribute{{Key: "foo", Val: "bar"}}
	n.SetAttr("foo", "baz")
	if n.Attr[0].Val != "baz" {
		t.Fatalf("expected value 'baz', got %q", n.Attr[0].Val)
	}
	if len(n.Attr) != 1 {
		t.Fatalf("expected 1 attribute, got %d", len(n.Attr))
	}
	if val := n.Attr[0].Val; val != "baz" {
		t.Fatalf("expected value 'baz', got %q", val)
	}
}

func TestSetAttrAppend(t *testing.T) {
	n := &Node{}
	n.SetAttr("foo", "bar")
	if len(n.Attr) != 1 || n.Attr[0].Val != "bar" {
		t.Fatalf("unexpected attrs: %+v", n.Attr)
	}
}
