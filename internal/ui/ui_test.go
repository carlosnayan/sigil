package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintTable(t *testing.T) {
	var buf bytes.Buffer
	printTable(&buf,
		[]string{"NAME", "VALUE", "SCOPE"},
		[][]string{{"config", "dev", "/tmp/proj"}},
	)
	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "config") || !strings.Contains(out, "/tmp/proj") {
		t.Fatalf("unexpected output:\n%s", out)
	}
	if !strings.Contains(out, "┌") || !strings.Contains(out, "└") {
		t.Fatalf("expected box borders:\n%s", out)
	}
}
