package routemanager

import (
	"fmt"
	"strings"
	"testing"
)

func TestLogBufferKeepsPerTunnelAndGlobalTails(t *testing.T) {
	logs := NewLogBuffer(100, 1000)
	for i := 0; i < 1100; i++ {
		logs.Append(fmt.Sprintf("tunnel-%d", i%2), fmt.Sprintf("line-%04d", i))
	}
	per := logs.Lines("tunnel-0", 100)
	if len(per) != 100 || per[0].Text != "line-0900" || per[99].Text != "line-1098" {
		t.Fatalf("unexpected per-tunnel tail: len=%d first=%q last=%q", len(per), per[0].Text, per[len(per)-1].Text)
	}
	all := logs.All(1000)
	if len(all) != 1000 || all[0].Text != "line-0100" || all[999].Text != "line-1099" {
		t.Fatalf("unexpected global tail: len=%d first=%q last=%q", len(all), all[0].Text, all[len(all)-1].Text)
	}
}

func TestLogBufferJoinsPartialWritesAndBoundsLongLines(t *testing.T) {
	logs := NewLogBuffer(100, 1000)
	writer := logs.Writer("tunnel")
	_, _ = writer.Write([]byte("hello "))
	_, _ = writer.Write([]byte("world\nlast line"))
	logs.Flush("tunnel")
	lines := logs.Lines("tunnel", 100)
	if len(lines) != 2 || lines[0].Text != "hello world" || lines[1].Text != "last line" {
		t.Fatalf("unexpected partial lines: %#v", lines)
	}
	logs.Append("tunnel", strings.Repeat("界", maxLogLineRunes+10))
	last := logs.Lines("tunnel", 1)[0].Text
	if len([]rune(last)) != maxLogLineRunes || !strings.HasSuffix(last, "…") {
		t.Fatalf("long line was not bounded: runes=%d", len([]rune(last)))
	}
}

func TestLogBufferVersionsChangeOnlyWhenACompleteLineIsAdded(t *testing.T) {
	logs := NewLogBuffer(100, 1000)
	writer := logs.Writer("one")
	_, _ = writer.Write([]byte("partial"))
	if got := logs.Version("one"); got != 0 {
		t.Fatalf("partial line version = %d", got)
	}
	_, _ = writer.Write([]byte(" line\n"))
	if got := logs.Version("one"); got != 1 {
		t.Fatalf("tunnel version = %d", got)
	}
	if got := logs.AllVersion(); got != 1 {
		t.Fatalf("global version = %d", got)
	}
	lines, version := logs.LinesSnapshot("one", 100)
	if version != 1 || len(lines) != 1 || lines[0].Text != "partial line" {
		t.Fatalf("snapshot = %#v, version = %d", lines, version)
	}
	logs.Append("two", "next")
	if logs.Version("one") != 1 || logs.Version("two") != 1 || logs.AllVersion() != 2 {
		t.Fatal("per-task and global versions are not independent")
	}
}
