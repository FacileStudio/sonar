package search

import (
	"bytes"
	"sync"
)

var logMu sync.Mutex

type syncWriter struct {
	buf *bytes.Buffer
}

func (w *syncWriter) Write(p []byte) (int, error) {
	logMu.Lock()
	defer logMu.Unlock()
	return w.buf.Write(p)
}

func read(buf *bytes.Buffer) string {
	logMu.Lock()
	defer logMu.Unlock()
	return buf.String()
}
