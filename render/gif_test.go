package render

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	imagegif "image/gif"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

var testPalette = color.Palette{
	color.RGBA{0, 0, 0, 255},
	color.RGBA{255, 0, 0, 255},
	color.RGBA{0, 255, 0, 255},
	color.RGBA{0, 0, 255, 255},
	color.RGBA{255, 255, 255, 255},
}

const (
	idxBlack = uint8(0)
	idxRed   = uint8(1)
	idxGreen = uint8(2)
	idxBlue  = uint8(3)
)

func newFrame(rect image.Rectangle, idx uint8) *image.Paletted {
	p := image.NewPaletted(rect, testPalette)
	for i := range p.Pix {
		p.Pix[i] = idx
	}
	return p
}

func writeGIF(t *testing.T, g *imagegif.GIF) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.gif")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create gif: %v", err)
	}
	if err := imagegif.EncodeAll(f, g); err != nil {
		f.Close()
		t.Fatalf("encode gif: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close gif: %v", err)
	}
	return path
}

func decodeGIF(t *testing.T, path string) *imagegif.GIF {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open gif: %v", err)
	}
	defer f.Close()
	g, err := imagegif.DecodeAll(f)
	if err != nil {
		t.Fatalf("decode gif: %v", err)
	}
	return g
}

func simpleGIF(t *testing.T, disposal []byte, frames ...*image.Paletted) string {
	t.Helper()
	delays := make([]int, len(frames))
	return writeGIF(t, &imagegif.GIF{
		Image:    frames,
		Delay:    delays,
		Disposal: disposal,
		Config:   image.Config{ColorModel: testPalette, Width: 4, Height: 4},
	})
}

func rgbaAt(img *image.RGBA, x, y int) color.RGBA {
	return img.RGBAAt(x, y)
}

func TestCompositeFullFrames(t *testing.T) {
	path := simpleGIF(t, []byte{imagegif.DisposalNone, imagegif.DisposalNone},
		newFrame(image.Rect(0, 0, 4, 4), idxRed),
		newFrame(image.Rect(0, 0, 2, 2), idxBlue),
	)
	frames := composite(decodeGIF(t, path))

	if len(frames) != 2 {
		t.Fatalf("got %d frames, want 2", len(frames))
	}
	for i, f := range frames {
		if b := f.Bounds(); b != image.Rect(0, 0, 4, 4) {
			t.Errorf("frame %d bounds = %v, want 0,0-4,4", i, b)
		}
	}

	if got := rgbaAt(frames[0], 3, 3); got != (color.RGBA{255, 0, 0, 255}) {
		t.Errorf("frame0(3,3) = %v, want red", got)
	}
	if got := rgbaAt(frames[1], 0, 0); got != (color.RGBA{0, 0, 255, 255}) {
		t.Errorf("frame1(0,0) = %v, want blue", got)
	}
	if got := rgbaAt(frames[1], 3, 3); got != (color.RGBA{255, 0, 0, 255}) {
		t.Errorf("frame1(3,3) = %v, want red from previous frame", got)
	}
}

func TestCompositeDisposalBackground(t *testing.T) {
	path := simpleGIF(t, []byte{imagegif.DisposalBackground, imagegif.DisposalNone},
		newFrame(image.Rect(0, 0, 4, 4), idxRed),
		newFrame(image.Rect(0, 0, 2, 2), idxGreen),
	)
	frames := composite(decodeGIF(t, path))

	if got := rgbaAt(frames[1], 0, 0); got != (color.RGBA{0, 255, 0, 255}) {
		t.Errorf("frame1(0,0) = %v, want green", got)
	}
	if got := rgbaAt(frames[1], 3, 3); got != (color.RGBA{0, 0, 0, 255}) {
		t.Errorf("frame1(3,3) = %v, want black after background disposal", got)
	}
}

func TestCompositeDisposalPrevious(t *testing.T) {
	path := simpleGIF(t, []byte{imagegif.DisposalPrevious, imagegif.DisposalNone},
		newFrame(image.Rect(0, 0, 4, 4), idxRed),
		newFrame(image.Rect(0, 0, 2, 2), idxGreen),
	)
	frames := composite(decodeGIF(t, path))

	if got := rgbaAt(frames[1], 0, 0); got != (color.RGBA{0, 255, 0, 255}) {
		t.Errorf("frame1(0,0) = %v, want green", got)
	}
	if got := rgbaAt(frames[1], 3, 3); got != (color.RGBA{0, 0, 0, 255}) {
		t.Errorf("frame1(3,3) = %v, want black after previous disposal", got)
	}
}

func TestBuildStepsAndLoop(t *testing.T) {
	path := simpleGIF(t, []byte{imagegif.DisposalNone, imagegif.DisposalNone, imagegif.DisposalNone},
		newFrame(image.Rect(0, 0, 4, 4), idxRed),
		newFrame(image.Rect(0, 0, 4, 4), idxGreen),
		newFrame(image.Rect(0, 0, 4, 4), idxBlue),
	)
	frames := composite(decodeGIF(t, path))
	sz := Size{Cols: 4, Rows: 2}

	steps, loop := build(frames, sz)
	if len(steps) != len(frames) {
		t.Fatalf("got %d steps, want %d", len(steps), len(frames))
	}

	cells := sz.Cols * sz.Rows // клеток на кадр
	wantColors := []string{
		"\033[38;2;255;0;0;48;2;255;0;0m",
		"\033[38;2;0;255;0;48;2;0;255;0m",
		"\033[38;2;0;0;255;48;2;0;0;255m",
	}
	for i, s := range steps {
		if !strings.HasPrefix(s, "\033[1;1H") {
			t.Errorf("step %d must start at cell 1,1: %q", i, s)
		}
		if !strings.Contains(s, wantColors[i]) {
			t.Errorf("step %d = %q, want color %q", i, s, wantColors[i])
		}
		if got := strings.Count(s, "▀"); got != cells {
			t.Errorf("step %d drew %d cells, want %d", i, got, cells)
		}
	}

	if !strings.Contains(loop, wantColors[0]) {
		t.Errorf("loop = %q, want color %q", loop, wantColors[0])
	}
	if got := strings.Count(loop, "▀"); got != cells {
		t.Errorf("loop drew %d cells, want %d", got, cells)
	}
}

func TestBuildSingleFrame(t *testing.T) {
	path := simpleGIF(t, nil, newFrame(image.Rect(0, 0, 4, 4), idxRed))
	frames := composite(decodeGIF(t, path))

	steps, loop := build(frames, Size{Cols: 2, Rows: 2})
	if len(steps) != 1 {
		t.Fatalf("got %d steps, want 1", len(steps))
	}
	if loop != "\033[0m" {
		t.Errorf("loop = %q, want %q", loop, "\033[0m")
	}
	if strings.Count(steps[0], "▀") != 4 {
		t.Errorf("step 0 = %q, want 4 cells", steps[0])
	}
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s", what)
}

func sendSize(t *testing.T, ch chan Size, sz Size) {
	t.Helper()
	select {
	case ch <- sz:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout sending resize")
	}
}

func TestGifMissingFile(t *testing.T) {
	err := Gif(context.Background(), io.Discard, filepath.Join(t.TempDir(), "nope.gif"), Size{Cols: 10, Rows: 5}, nil)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("got %v, want os.ErrNotExist", err)
	}
}

func TestGifCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.gif")
	if err := os.WriteFile(path, []byte("this is not a gif"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Gif(context.Background(), io.Discard, path, Size{Cols: 10, Rows: 5}, nil); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestGifCancelledContextWritesFirstFrame(t *testing.T) {
	path := simpleGIF(t, []byte{imagegif.DisposalNone, imagegif.DisposalNone},
		newFrame(image.Rect(0, 0, 4, 4), idxRed),
		newFrame(image.Rect(0, 0, 4, 4), idxGreen),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	if err := Gif(ctx, &buf, path, Size{Cols: 4, Rows: 2}, nil); err != nil {
		t.Fatalf("Gif returned %v, want nil", err)
	}

	out := buf.String()
	for _, want := range []string{
		"\033[?1049h\033[?25l\033[2J",
		"\033[1;1H",
		"▀",
		"\033[0m",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output misses %q: %q", want, out)
		}
	}
	if !strings.HasSuffix(out, "\033[?25h\033[?1049l") {
		t.Errorf("output must end with screen restore, got %q", out)
	}
}

func TestGifDefaultSize(t *testing.T) {
	path := simpleGIF(t, nil, newFrame(image.Rect(0, 0, 4, 4), idxRed))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	if err := Gif(ctx, &buf, path, Size{}, nil); err != nil {
		t.Fatalf("Gif returned %v, want nil", err)
	}

	out := buf.String()
	if !strings.Contains(out, "\033[24;1H") {
		t.Errorf("default size must be 80x24 (row 24 missing): %q", out)
	}
	if strings.Contains(out, "\033[25;") {
		t.Errorf("default size must not exceed 24 rows: %q", out)
	}
}

func TestGifWriteError(t *testing.T) {
	path := simpleGIF(t, nil, newFrame(image.Rect(0, 0, 4, 4), idxRed))

	err := Gif(context.Background(), errWriter{}, path, Size{Cols: 4, Rows: 2}, nil)
	if err == nil {
		t.Fatal("expected write error")
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestGifResize(t *testing.T) {
	path := simpleGIF(t, []byte{imagegif.DisposalNone, imagegif.DisposalNone},
		newFrame(image.Rect(0, 0, 4, 4), idxRed),
		newFrame(image.Rect(0, 0, 4, 4), idxGreen),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := &syncBuffer{}
	resize := make(chan Size, 1)
	errCh := make(chan error, 1)
	go func() {
		errCh <- Gif(ctx, out, path, Size{Cols: 4, Rows: 2}, resize)
	}()

	waitFor(t, "first frame", func() bool { return strings.Contains(out.String(), "▀") })

	// Нулевой размер должен быть проигнорирован, валидный — перестроить кадры.
	sendSize(t, resize, Size{Cols: 0, Rows: 0})
	sendSize(t, resize, Size{Cols: 2, Rows: 2})

	waitFor(t, "screen clear after resize", func() bool {
		return strings.Contains(out.String(), "\033[2J\033[H")
	})

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Gif returned %v, want nil", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Gif did not stop after cancel")
	}

	final := out.String()
	if got := strings.Count(final, "\033[2J\033[H"); got != 1 {
		t.Errorf("screen cleared %d times, want exactly 1 (invalid size must be ignored)", got)
	}
	if !strings.Contains(final, "\033[2;1H") {
		t.Errorf("expected cursor to row 2 after resize: %q", final)
	}
}
