package teamcity

import (
	"fmt"
	"strings"
	"testing"
)

func intPtr(i int) *int { return &i }

// makeLog builds a log with n numbered lines: "line 1".."line n".
func makeLog(n int) string {
	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	return strings.Join(lines, "\n")
}

// body returns everything after the blank line that separates the header from
// the log content, as a slice of lines.
func body(t *testing.T, out string) []string {
	t.Helper()
	idx := strings.Index(out, "\n\n")
	if idx < 0 {
		t.Fatalf("no header/body separator in output:\n%s", out)
	}
	content := out[idx+2:]
	if content == "" {
		return nil
	}
	return strings.Split(content, "\n")
}

func TestRenderBuildLog_DefaultCap(t *testing.T) {
	// An unconstrained call over a large log must be capped by default so it
	// doesn't dump the whole thing into the context.
	out := renderBuildLog("42", makeLog(10000), buildLogView{})

	lines := body(t, out)
	if len(lines) != defaultBuildLogMaxLines {
		t.Fatalf("expected %d lines, got %d", defaultBuildLogMaxLines, len(lines))
	}
	if lines[0] != "line 1" {
		t.Fatalf("expected first line 'line 1', got %q", lines[0])
	}
	if !strings.Contains(out, "Total lines: 10000") {
		t.Errorf("header missing total line count:\n%s", firstLines(out))
	}
	if !strings.Contains(out, "capped at 500 lines") {
		t.Errorf("expected default-cap NOTE, got:\n%s", firstLines(out))
	}
	if !strings.Contains(out, "startLine=501") {
		t.Errorf("expected next-page hint startLine=501, got:\n%s", firstLines(out))
	}
}

func TestRenderBuildLog_MaxLinesZeroReturnsEverything(t *testing.T) {
	out := renderBuildLog("42", makeLog(1200), buildLogView{MaxLines: intPtr(0)})

	lines := body(t, out)
	if len(lines) != 1200 {
		t.Fatalf("expected all 1200 lines, got %d", len(lines))
	}
	if strings.Contains(out, "NOTE:") {
		t.Errorf("did not expect a truncation NOTE when returning everything:\n%s", firstLines(out))
	}
}

func TestRenderBuildLog_Pagination(t *testing.T) {
	// Second page of 100 lines: startLine=101, maxLines=100 -> lines 101..200.
	out := renderBuildLog("42", makeLog(1000), buildLogView{
		StartLine: intPtr(101),
		MaxLines:  intPtr(100),
	})

	lines := body(t, out)
	if len(lines) != 100 {
		t.Fatalf("expected 100 lines, got %d", len(lines))
	}
	if lines[0] != "line 101" || lines[99] != "line 200" {
		t.Fatalf("unexpected page window: first=%q last=%q", lines[0], lines[99])
	}
	if !strings.Contains(out, "Showing lines 101-200") {
		t.Errorf("expected 'Showing lines 101-200', got:\n%s", firstLines(out))
	}
	if !strings.Contains(out, "startLine=201") {
		t.Errorf("expected next-page hint startLine=201, got:\n%s", firstLines(out))
	}
}

func TestRenderBuildLog_StartLineBeyondEnd(t *testing.T) {
	out := renderBuildLog("42", makeLog(10), buildLogView{StartLine: intPtr(50)})

	if lines := body(t, out); len(lines) != 1 || lines[0] != "(No lines to display for the requested range)" {
		t.Fatalf("expected empty-range message, got %#v", lines)
	}
}

func TestRenderBuildLog_StartLineBeyondFilteredEnd(t *testing.T) {
	// The filter DID match lines; only the requested page is empty. The body
	// must report an out-of-range window, not claim that nothing matched.
	log := "x\n[ERROR] a\nx\n[ERROR] b\nx"
	out := renderBuildLog("42", log, buildLogView{
		FilterPattern: `\[ERROR\]`,
		StartLine:     intPtr(50),
	})

	if !strings.Contains(out, "Matched lines: 2") {
		t.Errorf("expected 'Matched lines: 2', got:\n%s", firstLines(out))
	}
	if lines := body(t, out); len(lines) != 1 || lines[0] != "(No lines to display for the requested range)" {
		t.Fatalf("expected empty-range message (filters matched), got %#v", lines)
	}
}

func TestRenderBuildLog_TailLines(t *testing.T) {
	out := renderBuildLog("42", makeLog(1000), buildLogView{TailLines: intPtr(5)})

	lines := body(t, out)
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
	if lines[0] != "line 996" || lines[4] != "line 1000" {
		t.Fatalf("unexpected tail window: first=%q last=%q", lines[0], lines[4])
	}
	// The header must say the view is a tail window — "Showing lines 1-5" is
	// relative to it, not to the whole log.
	if !strings.Contains(out, "Tail: last 5 lines") {
		t.Errorf("expected 'Tail: last 5 lines' in header, got:\n%s", firstLines(out))
	}
	// tailLines already bounds the output, so no default-cap NOTE.
	if strings.Contains(out, "NOTE:") {
		t.Errorf("did not expect a NOTE with tailLines set:\n%s", firstLines(out))
	}
}

func TestRenderBuildLog_FilterPattern(t *testing.T) {
	log := "start\n[ERROR] boom\nnormal\n[ERROR] kaboom\nend"
	out := renderBuildLog("42", log, buildLogView{FilterPattern: `\[ERROR\]`})

	lines := body(t, out)
	if len(lines) != 2 {
		t.Fatalf("expected 2 matching lines, got %d: %#v", len(lines), lines)
	}
	if lines[0] != "[ERROR] boom" || lines[1] != "[ERROR] kaboom" {
		t.Fatalf("unexpected matches: %#v", lines)
	}
	if !strings.Contains(out, "Matched lines: 2") {
		t.Errorf("expected 'Matched lines: 2', got:\n%s", firstLines(out))
	}
}

func TestRenderBuildLog_FilterPatternNoMatch(t *testing.T) {
	out := renderBuildLog("42", "a\nb\nc", buildLogView{FilterPattern: "zzz"})

	if lines := body(t, out); len(lines) != 1 || lines[0] != "(No lines match the specified filters)" {
		t.Fatalf("expected no-match message, got %#v", lines)
	}
}

func TestRenderBuildLog_ContextLines(t *testing.T) {
	log := strings.Join([]string{
		"l1", "l2", "MATCH", "l4", "l5", "l6", "l7", "MATCH", "l9",
	}, "\n")
	out := renderBuildLog("42", log, buildLogView{
		FilterPattern: "MATCH",
		ContextLines:  intPtr(1),
	})

	lines := body(t, out)
	want := []string{"l2", "MATCH", "l4", "--", "l7", "MATCH", "l9"}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Fatalf("context window mismatch:\n got: %#v\nwant: %#v", lines, want)
	}
	// Matched lines counts matches only — not context lines or "--" separators.
	if !strings.Contains(out, "Matched lines: 2") {
		t.Errorf("expected 'Matched lines: 2' (matches only), got:\n%s", firstLines(out))
	}
}

func TestFilterBySeverity(t *testing.T) {
	log := []string{
		"[INFO] starting up",
		"[ERROR] Connection failed",
		"[WARN] Deprecated API usage",
		"[ERROR] Test TestFoo failed",
		"just a normal line",
		"",
	}

	tests := []struct {
		severity string
		want     []string
	}{
		{"error", []string{"[ERROR] Connection failed", "[ERROR] Test TestFoo failed"}},
		{"warning", []string{"[WARN] Deprecated API usage"}},
		// info excludes error/warning lines and blank lines.
		{"info", []string{"[INFO] starting up", "just a normal line"}},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			got := filterBySeverity(log, tt.severity)
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Fatalf("severity %q:\n got: %#v\nwant: %#v", tt.severity, got, tt.want)
			}
		})
	}
}

func TestFilterByPatternLiteralFallback(t *testing.T) {
	// An invalid regex must fall back to a literal substring search rather than
	// returning nothing.
	log := []string{"a(b", "cd", "a(b again"}
	got, matched := filterByPattern(log, "a(b", 0)
	if len(got) != 2 || matched != 2 {
		t.Fatalf("expected literal-substring fallback to match 2 lines, got %#v (matched=%d)", got, matched)
	}
}

func firstLines(s string) string {
	parts := strings.SplitN(s, "\n\n", 2)
	return parts[0]
}
