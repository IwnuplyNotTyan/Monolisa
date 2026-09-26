package render

import (
	"image"
	"testing"
)

func fill(img *image.RGBA, c [3]uint8) {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i+0] = c[0]
			img.Pix[i+1] = c[1]
			img.Pix[i+2] = c[2]
			img.Pix[i+3] = 255
		}
	}
}

func setPx(img *image.RGBA, x, y int, c [3]uint8) {
	i := img.PixOffset(x, y)
	img.Pix[i+0] = c[0]
	img.Pix[i+1] = c[1]
	img.Pix[i+2] = c[2]
	img.Pix[i+3] = 255
}

func TestToCellsCount(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for _, sz := range []struct{ cols, rows int }{{1, 1}, {4, 3}, {8, 8}, {16, 16}, {3, 5}} {
		cells := ToCells(img, sz.cols, sz.rows)
		if len(cells) != sz.cols*sz.rows {
			t.Errorf("ToCells(cols=%d, rows=%d) returned %d cells, want %d",
				sz.cols, sz.rows, len(cells), sz.cols*sz.rows)
		}
	}
}

func TestToCellsSolidColor(t *testing.T) {
	want := [3]uint8{12, 34, 56}
	img := image.NewRGBA(image.Rect(0, 0, 9, 7)) // odd size on purpose
	fill(img, want)

	cells := ToCells(img, 5, 3)
	if len(cells) != 15 {
		t.Fatalf("got %d cells, want 15", len(cells))
	}
	for i, c := range cells {
		if c.t != want || c.b != want {
			t.Fatalf("cell %d = {%v %v}, want top and bottom %v", i, c.t, c.b, want)
		}
	}
}

func TestToCellsIgnoresAlpha(t *testing.T) {
	want := [3]uint8{200, 100, 50}
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	fill(img, want)
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 0
	}

	cells := ToCells(img, 2, 2)
	for i, c := range cells {
		if c.t != want || c.b != want {
			t.Fatalf("cell %d = {%v %v}, alpha must not affect colors, want %v", i, c.t, c.b, want)
		}
	}
}

func TestToCellsHalfBlocks(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	rowColor := [4][3]uint8{{10, 10, 10}, {20, 20, 20}, {30, 30, 30}, {40, 40, 40}}
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			setPx(img, x, y, rowColor[y])
		}
	}

	cells := ToCells(img, 2, 2)
	want := [2][2][3]uint8{ // [row][top/bottom]
		{{10, 10, 10}, {20, 20, 20}},
		{{30, 30, 30}, {40, 40, 40}},
	}
	for r := 0; r < 2; r++ {
		for c := 0; c < 2; c++ {
			got := cells[r*2+c]
			if got.t != want[r][0] {
				t.Errorf("cell(%d,%d).t = %v, want %v", c, r, got.t, want[r][0])
			}
			if got.b != want[r][1] {
				t.Errorf("cell(%d,%d).b = %v, want %v", c, r, got.b, want[r][1])
			}
		}
	}
}

func TestToCellsAveragesBlock(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	fill(img, [3]uint8{0, 0, 0})

	setPx(img, 0, 0, [3]uint8{100, 0, 0})
	setPx(img, 1, 0, [3]uint8{200, 0, 0})
	setPx(img, 0, 1, [3]uint8{0, 100, 0})
	setPx(img, 1, 1, [3]uint8{0, 200, 0})

	cells := ToCells(img, 2, 2)
	if got := cells[0].t; got != [3]uint8{150, 0, 0} {
		t.Errorf("top half = %v, want [150 0 0]", got)
	}
	if got := cells[0].b; got != [3]uint8{0, 150, 0} {
		t.Errorf("bottom half = %v, want [0 150 0]", got)
	}
	// Не тронутые блоки остаются чёрными.
	if got := cells[3].t; got != [3]uint8{0, 0, 0} {
		t.Errorf("cell 3 top = %v, want [0 0 0]", got)
	}
}

func TestToCellsSmallerThanGrid(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	setPx(img, 0, 0, [3]uint8{255, 0, 0})

	cells := ToCells(img, 4, 3)
	if len(cells) != 12 {
		t.Fatalf("got %d cells, want 12", len(cells))
	}
	want := [3]uint8{255, 0, 0}
	for i, c := range cells {
		if c.t != want || c.b != want {
			t.Fatalf("cell %d = {%v %v}, want %v", i, c.t, c.b, want)
		}
	}
}

func TestToCellsRectangularImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 8))
	fill(img, [3]uint8{7, 8, 9})

	cells := ToCells(img, 10, 4)
	if len(cells) != 40 {
		t.Fatalf("got %d cells, want 40", len(cells))
	}
	want := [3]uint8{7, 8, 9}
	for i, c := range cells {
		if c.t != want || c.b != want {
			t.Fatalf("cell %d = {%v %v}, want %v", i, c.t, c.b, want)
		}
	}
}
