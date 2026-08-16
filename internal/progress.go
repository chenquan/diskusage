package internal

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

type progress struct {
	enabled   bool
	dirs      atomic.Int64
	files     atomic.Int64
	start     time.Time
	lastPrint atomic.Int64 // unix nano of last render, for throttling
}

var prog progress

func initProgress(enabled bool) {
	prog = progress{enabled: enabled, start: time.Now()}
}

// tickDir records a scanned directory and refreshes the progress line.
func tickDir() {
	if !prog.enabled {
		return
	}
	n := prog.dirs.Add(1)

	now := time.Now().UnixNano()
	last := prog.lastPrint.Load()
	if now-last < 100*time.Millisecond.Nanoseconds() && n != 1 {
		return
	}
	prog.lastPrint.Store(now)

	fmt.Fprintf(os.Stderr, "\rScanning... %d dirs, %d files (%.1fs)",
		n, prog.files.Load(), time.Since(prog.start).Seconds())
}

func tickFile() {
	if prog.enabled {
		prog.files.Add(1)
	}
}

// finishProgress clears the progress line and reports totals.
func finishProgress() {
	if !prog.enabled {
		return
	}

	fmt.Fprintf(os.Stderr, "\rScanned %d dirs, %d files in %.1fs\n",
		prog.dirs.Load(), prog.files.Load(), time.Since(prog.start).Seconds())
}

// stderrIsTerminal reports whether stderr is an interactive terminal
// (progress display is skipped when stderr is redirected).
func stderrIsTerminal() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
