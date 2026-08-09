package routemanager

import (
	"io"
	"strings"
	"sync"
	"time"
)

const maxLogLineRunes = 4096

type LogLine struct {
	Time     time.Time
	TunnelID string
	Text     string
}

type LogBuffer struct {
	mu       sync.Mutex
	perLimit int
	per      map[string]*lineRing
	all      *lineRing
	partial  map[string]string
	versions map[string]uint64
	version  uint64
}

func NewLogBuffer(perLimit, allLimit int) *LogBuffer {
	return &LogBuffer{
		perLimit: perLimit,
		per:      make(map[string]*lineRing),
		all:      newLineRing(allLimit),
		partial:  make(map[string]string),
		versions: make(map[string]uint64),
	}
}

func (b *LogBuffer) Writer(id string) io.Writer {
	return logWriter{buffer: b, id: id}
}

type logWriter struct {
	buffer *LogBuffer
	id     string
}

func (w logWriter) Write(p []byte) (int, error) {
	w.buffer.write(w.id, string(p))
	return len(p), nil
}

func (b *LogBuffer) write(id, text string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	combined := b.partial[id] + text
	parts := strings.Split(combined, "\n")
	for _, line := range parts[:len(parts)-1] {
		b.appendLocked(id, strings.TrimSuffix(line, "\r"))
	}
	b.partial[id] = truncateLogLine(parts[len(parts)-1])
}

func (b *LogBuffer) Append(id, text string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if partial := b.partial[id]; partial != "" {
		b.appendLocked(id, partial)
		b.partial[id] = ""
	}
	for _, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		b.appendLocked(id, strings.TrimSuffix(line, "\r"))
	}
}

func (b *LogBuffer) Flush(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if partial := b.partial[id]; partial != "" {
		b.appendLocked(id, partial)
		b.partial[id] = ""
	}
}

func (b *LogBuffer) appendLocked(id, text string) {
	line := LogLine{Time: time.Now(), TunnelID: id, Text: truncateLogLine(text)}
	ring := b.per[id]
	if ring == nil {
		ring = newLineRing(b.perLimit)
		b.per[id] = ring
	}
	ring.append(line)
	b.all.append(line)
	b.versions[id]++
	b.version++
}

func (b *LogBuffer) Lines(id string, limit int) []LogLine {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.per[id] == nil {
		return nil
	}
	return b.per[id].tail(limit)
}

func (b *LogBuffer) All(limit int) []LogLine {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.all.tail(limit)
}

func (b *LogBuffer) Version(id string) uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.versions[id]
}

func (b *LogBuffer) AllVersion() uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.version
}

func (b *LogBuffer) LinesSnapshot(id string, limit int) ([]LogLine, uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.per[id] == nil {
		return nil, b.versions[id]
	}
	return b.per[id].tail(limit), b.versions[id]
}

func (b *LogBuffer) AllSnapshot(limit int) ([]LogLine, uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.all.tail(limit), b.version
}

type lineRing struct {
	lines []LogLine
	next  int
	full  bool
}

func newLineRing(limit int) *lineRing {
	if limit < 0 {
		limit = 0
	}
	return &lineRing{lines: make([]LogLine, limit)}
}

func (r *lineRing) append(line LogLine) {
	if len(r.lines) == 0 {
		return
	}
	r.lines[r.next] = line
	r.next = (r.next + 1) % len(r.lines)
	if r.next == 0 {
		r.full = true
	}
}

func (r *lineRing) tail(limit int) []LogLine {
	count := r.next
	start := 0
	if r.full {
		count = len(r.lines)
		start = r.next
	}
	if limit <= 0 || count == 0 {
		return nil
	}
	if limit > count {
		limit = count
	}
	result := make([]LogLine, limit)
	first := (start + count - limit) % len(r.lines)
	for i := range result {
		result[i] = r.lines[(first+i)%len(r.lines)]
	}
	return result
}

func truncateLogLine(text string) string {
	runes := []rune(text)
	if len(runes) <= maxLogLineRunes {
		return text
	}
	return string(runes[:maxLogLineRunes-1]) + "…"
}
