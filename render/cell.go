package render

import "image"

type Cell struct{ t, b [3]uint8 }

func ToCells(img *image.RGBA, cols, rows int) []Cell {
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

	Cells := make([]Cell, cols*rows)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			Cells[r*cols+c] = Cell{px(c, r*2), px(c, r*2+1)}
		}
	}
	return Cells
}
