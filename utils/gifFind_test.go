package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func setArgs(t *testing.T, args ...string) {
	t.Helper()
	old := os.Args
	os.Args = append([]string{"monolisa"}, args...)
	t.Cleanup(func() { os.Args = old })
}

func writeGif(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("GIF89a"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

func TestIsDirectory(t *testing.T) {
	dir := t.TempDir()
	if !isDirectory(dir) {
		t.Errorf("isDirectory(%q) = false, want true", dir)
	}
	if isDirectory(filepath.Join(dir, "missing")) {
		t.Error("isDirectory(missing path) = true, want false")
	}

	file := writeGif(t, dir, "a.gif")
	if isDirectory(file) {
		t.Errorf("isDirectory(%q) = true, want false", file)
	}
}

func TestGifFindExplicitFile(t *testing.T) {
	want := writeGif(t, t.TempDir(), "only.gif")
	setArgs(t, want)

	if got := GifFind(); got != want {
		t.Errorf("GifFind() = %q, want %q", got, want)
	}
}

func TestGifFindMissingFileFallsBackToEnv(t *testing.T) {
	dir := t.TempDir()
	want := writeGif(t, dir, "a.gif")
	t.Setenv("MONOLISA_DIR", dir)
	setArgs(t, filepath.Join(dir, "missing.gif"))

	if got := GifFind(); got != want {
		t.Errorf("GifFind() = %q, want %q", got, want)
	}
}

func TestGifFindDirectoryArgument(t *testing.T) {
	dir := t.TempDir()
	a := writeGif(t, dir, "a.gif")
	b := writeGif(t, dir, "b.gif")
	setArgs(t, dir)

	got := GifFind()
	if got != a && got != b {
		t.Errorf("GifFind() = %q, want one of %q, %q", got, a, b)
	}
}

func TestGifFindEnvDir(t *testing.T) {
	dir := t.TempDir()
	want := writeGif(t, dir, "env.gif")
	t.Setenv("MONOLISA_DIR", dir)
	setArgs(t)

	if got := GifFind(); got != want {
		t.Errorf("GifFind() = %q, want %q", got, want)
	}
}

func TestGifFindWorkingDirectoryFallback(t *testing.T) {
	dir := t.TempDir()
	want := writeGif(t, dir, "cwd.gif")
	t.Setenv("MONOLISA_DIR", "")
	t.Chdir(dir)
	setArgs(t)

	if got := GifFind(); got != want {
		t.Errorf("GifFind() = %q, want %q", got, want)
	}
}
