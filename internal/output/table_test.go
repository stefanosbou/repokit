package output

import (
	"bytes"
	"strings"
	"testing"
	"time"
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
		d    time.Duration
		want string
	}{
		{30 * time.Second, "30s ago"},
		{120 * time.Second, "2m ago"},
		{3600 * time.Second, "1h ago"},
		{86400 * time.Second, "1d ago"},
	}
	for _, c := range cases {
		if got := RelTime(c.d); got != c.want {
			t.Errorf("RelTime(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}
