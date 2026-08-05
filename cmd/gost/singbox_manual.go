package main

import (
	_ "embed"
	"io"
)

// singboxManual is kept as Markdown so the command output and the repository
// manual have one canonical source.
//
//go:embed SINGBOX_MANUAL.md
var singboxManual string

func writeSingboxManual(w io.Writer) error {
	_, err := io.WriteString(w, singboxManual)
	return err
}
