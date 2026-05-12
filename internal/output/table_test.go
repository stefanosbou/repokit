package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestTable_AlignsColumns(t *testing.T) {
	var buf bytes.Buffer
	tw := NewTable(&buf, []string{"NAME", "PATH"})
	tw.Row("alpha", "/tmp/a")
	tw.Row("beta-very-long", "/tmp/b")
	tw.Flush()
	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "PATH") {
		t.Errorf("missing headers: %q", out)
	}
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta-very-long") {
		t.Errorf("missing rows: %q", out)
	}
}

func TestRelTime(t *testing.T) {
	cases := []struct {
		secs int64
		want string
	}{
		{30, "30s ago"},
		{120, "2m ago"},
		{3600, "1h ago"},
		{86400, "1d ago"},
	}
	for _, c := range cases {
		if got := RelTime(c.secs); got != c.want {
			t.Errorf("RelTime(%d) = %q, want %q", c.secs, got, c.want)
		}
	}
}
