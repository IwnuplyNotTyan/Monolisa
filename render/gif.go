package render

import (
	"bufio"
	"context"
	"image"
	"image/draw"
	"image/gif"
	"io"
	"os"
	"time"
)

type Size struct{ Cols, Rows int }

// composite разворачивает GIF в полные RGBA-кадры (один раз).
func composite(g *gif.GIF) []*image.RGBA {
	canvas := image.NewRGBA(image.Rect(0, 0, g.Config.Width, g.Config.Height))
	draw.Draw(canvas, canvas.Bounds(), image.Black, image.Point{}, draw.Src)
	var snap []uint8
	frames := make([]*image.RGBA, len(g.Image))

	for i, p := range g.Image {
		if i > 0 {
			switch g.Disposal[i-1] {
			case gif.DisposalBackground:
				draw.Draw(canvas, g.Image[i-1].Bounds(), image.Black, image.Point{}, draw.Src)
			case gif.DisposalPrevious:
				copy(canvas.Pix, snap)
			}
		}
		if g.Disposal[i] == gif.DisposalPrevious {
			snap = append(snap[:0], canvas.Pix...)
		}
		draw.Draw(canvas, p.Bounds(), p, p.Bounds().Min, draw.Over)

		cp := image.NewRGBA(canvas.Bounds())
		copy(cp.Pix, canvas.Pix)
		frames[i] = cp
	}
	return frames
}

// build считает дифф-строки под конкретный размер терминала.
func build(frames []*image.RGBA, sz Size) (steps []string, loop string) {
	n := len(frames)
	cells := make([][]Cell, n)
	for i, f := range frames {
		cells[i] = ToCells(f, sz.Cols, sz.Rows)
	}
	steps = make([]string, n)
	for i := range cells {
		var prev []Cell
		if i > 0 {
			prev = cells[i-1]
		}
		steps[i] = Diff(prev, cells[i], sz.Cols)
	}
	loop = Diff(cells[n-1], cells[0], sz.Cols)
	return
}

// Gif рисует анимацию. resize может быть nil (локальный запуск).
func Gif(ctx context.Context, w io.Writer, gifSource string, size Size, resize <-chan Size) error {
	if size.Cols <= 0 || size.Rows <= 0 {
		size = Size{Cols: 80, Rows: 24}
	}
	f, err := os.Open(gifSource)
	if err != nil {
		return err
	}
	g, err := gif.DecodeAll(f)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}

	frames := composite(g)
	steps, loop := build(frames, size)

	out := bufio.NewWriterSize(w, 1<<20)
	defer func() {
		_, _ = out.WriteString("\033[?25h\033[?1049l")
		_ = out.Flush()
	}()

	flush := func(s string) error {
		if _, err := out.WriteString(s); err != nil {
			return err
		}
		return out.Flush()
	}

	if err := flush("\033[?1049h\033[?25l\033[2J"); err != nil {
		return err
	}

	first := true
	for i := 0; ; {
		s := steps[i]
		if i == 0 && !first {
			s = loop
		}
		if err := flush(s); err != nil {
			return err
		}

		d := time.Duration(g.Delay[i]) * 10 * time.Millisecond
		if d < 20*time.Millisecond {
			d = 80 * time.Millisecond
		}

		select {
		case <-ctx.Done():
			return nil
		case sz := <-resize:
			if sz.Cols <= 0 || sz.Rows <= 0 || sz == size {
				continue
			}
			size = sz
			steps, loop = build(frames, size)
			_, _ = out.WriteString("\033[2J\033[H") // очистить экран
			// начать с кадра 0 как "первого" — Diff от пустого
			i, first = 0, true
			continue
		case <-time.After(d):
		}

		i++
		if i == len(frames) {
			i, first = 0, false
		}
	}
}
