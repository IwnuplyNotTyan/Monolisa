package render

import (
	"fmt"
	"strings"
)

func rgb(c [3]uint8) string { return fmt.Sprintf("%d;%d;%d", c[0], c[1], c[2]) }

func Diff(prev, cur []Cell, cols int) string {
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
