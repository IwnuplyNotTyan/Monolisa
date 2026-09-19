package render

import (
	"bufio"
	"image"
	"image/draw"
	"image/gif"
	"os"
	"os/signal"
	"time"

	"golang.org/x/term"
)

func Gif(gifSource string) {
	f, err := os.Open(gifSource)
	if err != nil {
		panic(err)
	}
	g, err := gif.DecodeAll(f)
	f.Close()
	if err != nil {
		panic(err)
	}

	cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}

	canvas := image.NewRGBA(image.Rect(0, 0, g.Config.Width, g.Config.Height))
	draw.Draw(canvas, canvas.Bounds(), image.Black, image.Point{}, draw.Src)
	var snap []uint8
	cells := make([][]Cell, len(g.Image))
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
		cells[i] = ToCells(canvas, cols, rows)
	}

	n := len(cells)
	steps := make([]string, n)
	for i := range cells {
		var prev []Cell
		if i > 0 {
			prev = cells[i-1]
		}
		steps[i] = Diff(prev, cells[i], cols)
	}
	loop := Diff(cells[n-1], cells[0], cols)

	out := bufio.NewWriterSize(os.Stdout, 1<<20)
	out.WriteString("\033[?1049h\033[?25l\033[2J")
	out.Flush()
	defer func() {
		out.WriteString("\033[?25h\033[?1049l")
		out.Flush()
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	for first := true; ; first = false {
		for i := range cells {
			s := steps[i]
			if i == 0 && !first {
				s = loop
			}
			out.WriteString(s)
			out.Flush()

			d := time.Duration(g.Delay[i]) * 10 * time.Millisecond
			if d < 20*time.Millisecond {
				d = 80 * time.Millisecond
			}
			select {
			case <-sig:
				return
			case <-time.After(d):
			}
		}
	}
}
