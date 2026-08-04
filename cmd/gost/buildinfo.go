package main

import (
	"fmt"
	"io"
	"runtime"

	"github.com/go-gost/x/backend/singbox"
)

func printBuildInfo(w io.Writer) {
	m := singbox.BuildManifest()
	fmt.Fprintf(w, "gost %s (%s %s/%s) flavor=%s", version, runtime.Version(), runtime.GOOS, runtime.GOARCH, m.Flavor)
	if m.Compiled {
		fmt.Fprintf(w, " sing-box=%s features=%s", m.PinnedVersion, m.Features)
	}
	fmt.Fprintln(w)
}
