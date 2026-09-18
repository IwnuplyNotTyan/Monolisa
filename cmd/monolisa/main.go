package main

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"os"
	"os/signal"
	"strings"
	"time"

	"golang.org/x/term"
)

type cell struct{ t, b [3]uint8 }

func rgb(c [3]uint8) string { return fmt.Sprintf("%d;%d;%d", c[0], c[1], c[2]) }

func toCells(img *image.RGBA, cols, rows int) []cell {
	W, H := img.Bounds().Dx(), img.Bounds().Dy()
	pw, ph := cols, rows*2

	xs := make([]int, pw+1)
	for x := 0; x <= pw; x++ {
		xs[x] = x * W / pw
	}
	ys := make([]int, ph+1)
	for y := 0; y <= ph; y++ {
		ys[y] = y * H / ph
	}

	px := func(x, y int) [3]uint8 {
		x0, x1 := xs[x], xs[x+1]
		y0, y1 := ys[y], ys[y+1]
		if x1 <= x0 {
			x1 = x0 + 1
		}
		if y1 <= y0 {
			y1 = y0 + 1
		}
		if x1 > W {
			x0, x1 = W-1, W
		}
		if y1 > H {
			y0, y1 = H-1, H
		}
		var r, g, b, n uint32
		for sy := y0; sy < y1; sy++ {
			i := img.PixOffset(x0, sy)
			for sx := x0; sx < x1; sx++ {
				r += uint32(img.Pix[i])
				g += uint32(img.Pix[i+1])
				b += uint32(img.Pix[i+2])
				n++
				i += 4
			}
		}
		return [3]uint8{uint8(r / n), uint8(g / n), uint8(b / n)}
	}

	cells := make([]cell, cols*rows)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			cells[r*cols+c] = cell{px(c, r*2), px(c, r*2+1)}
		}
	}
	return cells
}

func diff(prev, cur []cell, cols int) string {
	var sb strings.Builder
	lastIdx := -2
	var lastT, lastB [3]uint8
	haveColor := false
	for i, c := range cur {
		if prev != nil && prev[i] == c {
			continue
		}
		if i != lastIdx+1 || i%cols == 0 {
			fmt.Fprintf(&sb, "\033[%d;%dH", i/cols+1, i%cols+1)
		}
		if !haveColor || c.t != lastT || c.b != lastB {
			fmt.Fprintf(&sb, "\033[38;2;%s;48;2;%sm", rgb(c.t), rgb(c.b))
			lastT, lastB, haveColor = c.t, c.b, true
		}
		sb.WriteString("▀")
		lastIdx = i
	}
	sb.WriteString("\033[0m")
	return sb.String()
}

func main() {
	f, err := os.Open("animated.gif")
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
	cells := make([][]cell, len(g.Image))
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
		cells[i] = toCells(canvas, cols, rows)
	}

	n := len(cells)
	steps := make([]string, n)
	for i := range cells {
		var prev []cell
		if i > 0 {
			prev = cells[i-1]
		}
		steps[i] = diff(prev, cells[i], cols)
	}
	loop := diff(cells[n-1], cells[0], cols)

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
