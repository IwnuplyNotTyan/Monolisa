//go:build !ssh

package remote

import (
	"context"
	"monolisa/render"
	"monolisa/utils"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"golang.org/x/term"
)

func Init() {
	fd := int(os.Stdout.Fd())
	if !term.IsTerminal(fd) {
		log.Fatal(os.Stderr, "stdout is not a terminal")
	}

	cols, rows, err := term.GetSize(fd)
	if err != nil {
		log.Fatal(os.Stderr, "failed to get terminal size:", err)
	}
	size := render.Size{Cols: cols, Rows: rows}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	resize := make(chan render.Size, 1)
	go func() {
		t := time.NewTicker(300 * time.Millisecond)
		defer t.Stop()
		last := size
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				c, r, err := term.GetSize(fd)
				if err != nil {
					continue
				}
				sz := render.Size{Cols: c, Rows: r}
				if sz == last {
					continue
				}
				last = sz
				select {
				case <-resize:
				default:
				}
				resize <- sz
			}
		}
	}()

	if err := render.Gif(ctx, os.Stdout, utils.GifFind(), size, resize); err != nil {
		log.Fatal(os.Stderr, err)
	}
}
