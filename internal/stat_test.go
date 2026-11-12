package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chenquan/diskusage/internal/worker"
	"github.com/jedib0t/go-pretty/v6/list"
)

// buildInfoFile must not produce NaN% when totalSize is 0 (empty tree + --all).
func TestBuildInfoFile_ZeroTotalSize(t *testing.T) {
	files := []*file{
		{name: "empty-dir", isDir: true, print: true},
		{name: "empty-file", print: true},
	}

	infos := buildInfoFile(list.NewWriter(), files, 0, 1, "M", 0, false)
	if len(infos) == 0 {
		t.Fatal("expected info for marked entries")
	}
	for _, info := range infos {
		if info.usageRate != info.usageRate { // NaN check: NaN != NaN
			t.Fatal("usageRate is NaN for zero totalSize")
		}
		if info.usageRate != 0 {
			t.Fatalf("usageRate = %v, want 0", info.usageRate)
		}
	}
}

// find must tolerate a symlink loop entry instead of erroring out.
func TestFind_SymlinkLoopEntry(t *testing.T) {
	dir := t.TempDir()
	if err := os.Symlink(dir, filepath.Join(dir, "loop")); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	files, err := find(dir, func(_ os.FileInfo) bool { return true },
		func(a []*file, i, j int) bool { return a[i].size > a[j].size })
	if err != nil {
		t.Fatalf("find returned error: %v", err)
	}
	if len(files) != 1 || files[0].name != "loop" {
		t.Fatalf("unexpected result: %+v", files)
	}
}

func TestFind_ExcludeSkipsSubtree(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "node_modules", "big"), []byte("xxxx"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "keep.js"), []byte("xx"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := find(dir, func(info os.FileInfo) bool {
		return info.Name() != "node_modules" && info.Name() != "keep.js"
	}, func(a []*file, i, j int) bool { return a[i].size > a[j].size })
	if err != nil {
		t.Fatalf("find returned error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected excluded entries to be skipped, got %+v", files)
	}
}

func TestGetReduce(t *testing.T) {
	// unit downgrades only until the value stays >= 0.1
	val, unit := getReduce("G", 500)
	if val != 500.0/1024 || unit != "K" {
		t.Fatalf("getReduce(G, 500) = %v %q, want %.6f K", val, unit, 500.0/1024)
	}
	val, unit = getReduce("G", 2*GB)
	if val != 2 || unit != "G" {
		t.Fatalf("getReduce(G, 2GB) = %v %q, want 2 G", val, unit)
	}
	val, unit = getReduce("M", 50)
	if val != 50 || unit != "B" {
		t.Fatalf("getReduce(M, 50) = %v %q, want 50 B", val, unit)
	}
}

// Hard links must be counted once and attributed to the lexicographically
// first path, regardless of the concurrent scan order.
func TestFind_HardLinkCountedOnce(t *testing.T) {
	oldWorker := w
	w = worker.New(4)
	defer func() {
		w.Close()
		w = oldWorker
	}()

	dir := t.TempDir()
	first := filepath.Join(dir, "a", "file.bin")
	second := filepath.Join(dir, "b", "file.bin")
	for _, p := range []string{first, second} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(first, []byte("hard link payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(first, second); err != nil {
		t.Skipf("hard links not supported: %v", err)
	}

	info, err := os.Stat(first)
	if err != nil {
		t.Fatal(err)
	}
	want := diskSize(info, first)

	files, err := find(dir, func(os.FileInfo) bool { return true },
		func(a []*file, i, j int) bool { return a[i].size > a[j].size })
	if err != nil {
		t.Fatalf("find returned error: %v", err)
	}

	sizes := make(map[string]int64)
	var total int64
	for _, d := range files {
		total += d.size
		for _, f := range d.sub {
			sizes[d.name+"/"+f.name] = f.size
		}
	}

	if total != want {
		t.Fatalf("total = %d, want %d (hard link counted twice?)", total, want)
	}
	if sizes["a/file.bin"] != want {
		t.Fatalf("a/file.bin = %d, want %d", sizes["a/file.bin"], want)
	}
	if sizes["b/file.bin"] != 0 {
		t.Fatalf("b/file.bin = %d, want 0", sizes["b/file.bin"])
	}
}

// A hard-linked file whose other links are outside the scanned tree must still
// be counted in full.
func TestFind_HardLinkOutsideRootKept(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(dir, "outside.bin")
	insideDir := filepath.Join(dir, "inside")
	if err := os.MkdirAll(insideDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	inside := filepath.Join(insideDir, "inside.bin")
	if err := os.Link(outside, inside); err != nil {
		t.Skipf("hard links not supported: %v", err)
	}

	info, err := os.Stat(outside)
	if err != nil {
		t.Fatal(err)
	}
	want := diskSize(info, outside)

	files, err := find(insideDir, func(os.FileInfo) bool { return true },
		func(a []*file, i, j int) bool { return a[i].size > a[j].size })
	if err != nil {
		t.Fatalf("find returned error: %v", err)
	}
	if len(files) != 1 || files[0].name != "inside.bin" {
		t.Fatalf("unexpected result: %+v", files)
	}
	if files[0].size != want {
		t.Fatalf("size = %d, want %d", files[0].size, want)
	}
}
