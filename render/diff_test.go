package render

import (
	"strings"
	"testing"
)

func cell(t, b [3]uint8) Cell { return Cell{t: t, b: b} }

func TestRGB(t *testing.T) {
	for _, tc := range []struct {
		c    [3]uint8
		want string
	}{
		{[3]uint8{0, 0, 0}, "0;0;0"},
		{[3]uint8{1, 2, 3}, "1;2;3"},
		{[3]uint8{255, 128, 7}, "255;128;7"},
	} {
		if got := rgb(tc.c); got != tc.want {
			t.Errorf("rgb(%v) = %q, want %q", tc.c, got, tc.want)
		}
	}
}

func TestDiffNoPreviousRendersEverything(t *testing.T) {
	c := cell([3]uint8{1, 2, 3}, [3]uint8{4, 5, 6})
	got := Diff(nil, []Cell{c, c}, 2)

	want := "\033[1;1H\033[38;2;1;2;3;48;2;4;5;6m▀▀\033[0m"
	if got != want {
		t.Errorf("Diff = %q, want %q", got, want)
	}
}

func TestDiffUnchangedProducesResetOnly(t *testing.T) {
	cur := []Cell{
		cell([3]uint8{1, 1, 1}, [3]uint8{2, 2, 2}),
		cell([3]uint8{3, 3, 3}, [3]uint8{4, 4, 4}),
	}
	if got := Diff(cur, cur, 2); got != "\033[0m" {
		t.Errorf("Diff of identical frames = %q, want %q", got, "\033[0m")
	}
}

func TestDiffCursorPosition(t *testing.T) {
	c := cell([3]uint8{9, 9, 9}, [3]uint8{1, 1, 1})
	got := Diff(nil, []Cell{c, c, c, c, c, c, c, c, c}, 3)

	want := "\033[1;1H\033[38;2;9;9;9;48;2;1;1;1m▀▀▀" +
		"\033[2;1H▀▀▀" +
		"\033[3;1H▀▀▀" +
		"\033[0m"
	if got != want {
		t.Errorf("Diff = %q, want %q", got, want)
	}
}

func TestDiffNonContiguousCellsMoveCursor(t *testing.T) {
	red := cell([3]uint8{255, 0, 0}, [3]uint8{0, 0, 0})
	blue := cell([3]uint8{0, 0, 255}, [3]uint8{0, 0, 0})
	prev := []Cell{blue, blue, blue, blue}
	cur := []Cell{red, blue, red, blue}

	got := Diff(prev, cur, 4)

	want := "\033[1;1H\033[38;2;255;0;0;48;2;0;0;0m▀" +
		"\033[1;3H▀" +
		"\033[0m"
	if got != want {
		t.Errorf("Diff = %q, want %q", got, want)
	}
}

func TestDiffEmitsColorOnChange(t *testing.T) {
	a := cell([3]uint8{10, 20, 30}, [3]uint8{40, 50, 60})
	b := cell([3]uint8{60, 50, 40}, [3]uint8{30, 20, 10})

	got := Diff(nil, []Cell{a, b}, 2)
	want := "\033[1;1H\033[38;2;10;20;30;48;2;40;50;60m▀" +
		"\033[38;2;60;50;40;48;2;30;20;10m▀" +
		"\033[0m"
	if got != want {
		t.Errorf("Diff = %q, want %q", got, want)
	}
}

func TestDiffOnlyChangedCells(t *testing.T) {
	base := cell([3]uint8{1, 1, 1}, [3]uint8{2, 2, 2})
	other := cell([3]uint8{3, 3, 3}, [3]uint8{4, 4, 4})

	prev := []Cell{base, base, base, base}
	cur := []Cell{base, base, other, base}

	got := Diff(prev, cur, 2)
	want := "\033[2;1H\033[38;2;3;3;3;48;2;4;4;4m▀\033[0m"
	if got != want {
		t.Errorf("Diff = %q, want %q", got, want)
	}
}

func TestDiffAlwaysResets(t *testing.T) {
	out := Diff(nil, []Cell{cell([3]uint8{1, 2, 3}, [3]uint8{4, 5, 6})}, 1)
	if !strings.HasSuffix(out, "\033[0m") {
		t.Errorf("Diff output must end with reset, got %q", out)
	}
	if strings.Count(out, "\033[0m") != 1 {
		t.Errorf("Diff must reset exactly once, got %q", out)
	}
}
